package v1

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"

	"devops-backend/global"
	"devops-backend/model/response"
	"devops-backend/service"
	"devops-backend/utils"
)

type PodExecController struct{}

var execUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type ExecSession struct {
	Conn      *websocket.Conn
	Client    *utils.ClusterClient
	Namespace string
	PodName   string
	Container string
	Cancel    context.CancelFunc
	Mutex     sync.Mutex
	SizeChan  chan remotecommand.TerminalSize
}

var execSessions sync.Map

func (c *PodExecController) ExecPod(ctx *gin.Context) {
	clusterID, err := resolveClusterID(ctx.Query("cluster"))
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(err.Error()))
		return
	}

	namespace := ctx.Query("namespace")
	if namespace == "" {
		namespace = "default"
	}

	podName := ctx.Query("pod")
	if podName == "" {
		ctx.JSON(http.StatusOK, response.Fail("Pod名称不能为空"))
		return
	}

	container := ctx.Query("container")
	shell := ctx.DefaultQuery("shell", "/bin/sh")

	cluster, err := (&service.ClusterService{}).GetClusterByID(uint(clusterID))
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(err.Error()))
		return
	}

	client, err := utils.GetLogStreamClient(cluster.Kubeconfig)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(err.Error()))
		return
	}

	conn, err := execUpgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		global.GVA_LOG.Error("WebSocket升级失败", zap.Error(err))
		return
	}

	sessionID := fmt.Sprintf("exec-%d-%s-%s-%d", clusterID, namespace, podName, time.Now().UnixNano())

	sizeChan := make(chan remotecommand.TerminalSize, 1)

	session := &ExecSession{
		Conn:      conn,
		Client:    client,
		Namespace: namespace,
		PodName:   podName,
		Container: container,
		SizeChan:  sizeChan,
	}

	execSessions.Store(sessionID, session)

	go c.handleExecSession(session, shell, sessionID)
}

func (c *PodExecController) handleExecSession(session *ExecSession, shell string, sessionID string) {
	ctx, cancel := context.WithCancel(context.Background())
	session.Cancel = cancel

	defer func() {
		cancel()
		session.Conn.Close()
		close(session.SizeChan)
		execSessions.Delete(sessionID)
	}()

	req := session.Client.Clientset.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(session.PodName).
		Namespace(session.Namespace).
		SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			Container: session.Container,
			Command:   []string{shell},
			Stdin:     true,
			Stdout:    true,
			Stderr:    true,
			TTY:       true,
		}, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(session.Client.RestConfig, "POST", req.URL())
	if err != nil {
		session.Mutex.Lock()
		session.Conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("\x1b[31m✗ 无法创建终端: %v\x1b[0m\n", err)))
		session.Mutex.Unlock()
		return
	}

	stream := &WebSocketStream{
		Conn:     session.Conn,
		SizeChan: session.SizeChan,
		Mutex:    &session.Mutex,
	}

	// session.Mutex.Lock()
	// session.Conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("\x1b[32m✓ 终端已连接\x1b[0m\n")))
	// session.Conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("\x1b[36mPod: %s | Container: %s | Shell: %s\x1b[0m\n", session.PodName, session.Container, shell)))
	// session.Conn.WriteMessage(websocket.TextMessage, []byte("\n"))
	// session.Mutex.Unlock()

	sizeQueue := make(chan remotecommand.TerminalSize, 1)

	go func() {
		for {
			select {
			case size := <-session.SizeChan:
				select {
				case sizeQueue <- size:
				default:
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	tsq := chanTerminalSizeQueue(sizeQueue)

	err = exec.StreamWithContext(ctx, remotecommand.StreamOptions{
		Stdin:             stream,
		Stdout:            stream,
		Stderr:            stream,
		Tty:               true,
		TerminalSizeQueue: tsq,
	})

	if err != nil {
		session.Mutex.Lock()
		session.Conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("\n\x1b[31m✗ 终端已断开: %v\x1b[0m\n", err)))
		session.Mutex.Unlock()
	} else {
		session.Mutex.Lock()
		session.Conn.WriteMessage(websocket.TextMessage, []byte("\n\x1b[33m■ 终端已关闭\x1b[0m\n"))
		session.Mutex.Unlock()
	}
}

func (c *PodExecController) StopExec(ctx *gin.Context) {
	sessionID := ctx.Query("session_id")
	if sessionID == "" {
		ctx.JSON(http.StatusOK, response.Fail("session_id不能为空"))
		return
	}

	if session, ok := execSessions.Load(sessionID); ok {
		s := session.(*ExecSession)
		if s.Cancel != nil {
			s.Cancel()
		}
		execSessions.Delete(sessionID)
		ctx.JSON(http.StatusOK, response.Success(nil))
		return
	}

	ctx.JSON(http.StatusOK, response.Fail("session不存在"))
}

type WebSocketStream struct {
	Conn     *websocket.Conn
	SizeChan chan remotecommand.TerminalSize
	Mutex    *sync.Mutex
}

func (s *WebSocketStream) Read(p []byte) (int, error) {
	for {
		msgType, data, err := s.Conn.ReadMessage()
		if err != nil {
			return 0, err
		}

		if msgType == websocket.TextMessage {
			if len(data) > 0 && data[0] == '\x01' {
				var width, height uint16
				fmt.Sscanf(string(data[1:]), "%d,%d", &width, &height)
				if s.SizeChan != nil {
					s.SizeChan <- remotecommand.TerminalSize{Width: width, Height: height}
				}
				continue
			}
			return copy(p, data), nil
		}

		if msgType == websocket.BinaryMessage {
			return copy(p, data), nil
		}
	}
}

func (s *WebSocketStream) Write(p []byte) (int, error) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()
	err := s.Conn.WriteMessage(websocket.BinaryMessage, p)
	if err != nil {
		return 0, err
	}
	return len(p), nil
}

type chanTerminalSizeQueue chan remotecommand.TerminalSize

func (q chanTerminalSizeQueue) Next() *remotecommand.TerminalSize {
	select {
	case size := <-q:
		return &size
	default:
		return nil
	}
}
