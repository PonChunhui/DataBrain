package v1

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"

	corev1 "k8s.io/api/core/v1"

	"devops-backend/global"
	"devops-backend/model/response"
	"devops-backend/service"
	"devops-backend/utils"
)

type PodLogController struct{}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type LogSession struct {
	Conn      *websocket.Conn
	Client    *utils.ClusterClient
	Namespace string
	PodName   string
	Container string
	Cancel    context.CancelFunc
	Mutex     sync.Mutex
}

var logSessions sync.Map

func (c *PodLogController) StreamLogs(ctx *gin.Context) {
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
	tailLines := ctx.DefaultQuery("tail_lines", "100")
	follow := ctx.DefaultQuery("follow", "true") == "true"

	cluster, err := (&service.ClusterService{}).GetClusterByID(uint(clusterID))
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(err.Error()))
		return
	}

	logClient, err := utils.GetLogStreamClient(cluster.Kubeconfig)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(err.Error()))
		return
	}

	conn, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		global.GVA_LOG.Error("WebSocket升级失败", zap.Error(err))
		return
	}

	sessionID := fmt.Sprintf("%d-%s-%s-%d", clusterID, namespace, podName, time.Now().UnixNano())

	session := &LogSession{
		Conn:      conn,
		Client:    logClient,
		Namespace: namespace,
		PodName:   podName,
		Container: container,
	}

	logSessions.Store(sessionID, session)

	go c.streamLogsToWebSocket(session, tailLines, follow, sessionID)
}

func (c *PodLogController) streamLogsToWebSocket(session *LogSession, tailLines string, follow bool, sessionID string) {
	ctx, cancel := context.WithCancel(context.Background())
	session.Cancel = cancel

	defer func() {
		cancel()
		session.Conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(1000, "日志流正常结束"))
		session.Conn.Close()
		logSessions.Delete(sessionID)
	}()

	tailLinesInt, _ := strconv.ParseInt(tailLines, 10, 64)

	opts := &corev1.PodLogOptions{
		Container: session.Container,
		Follow:    follow,
		TailLines: &tailLinesInt,
	}

	req := session.Client.Clientset.CoreV1().Pods(session.Namespace).GetLogs(session.PodName, opts)

	stream, err := req.Stream(ctx)
	if err != nil {
		session.Mutex.Lock()
		session.Conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("\x1b[31m✗ 无法获取日志: %v\x1b[0m\n", err)))
		session.Mutex.Unlock()
		return
	}
	defer stream.Close()

	session.Mutex.Lock()
	session.Conn.WriteMessage(websocket.TextMessage, []byte("\x1b[32m✓ 日志流已开始\x1b[0m\n\n"))
	session.Mutex.Unlock()

	buf := make([]byte, 4096)
	for {
		select {
		case <-ctx.Done():
			return
		default:
			n, err := stream.Read(buf)
			if n > 0 {
				session.Mutex.Lock()
				writeErr := session.Conn.WriteMessage(websocket.TextMessage, buf[:n])
				session.Mutex.Unlock()
				if writeErr != nil {
					global.GVA_LOG.Info("WebSocket写入失败", zap.Error(writeErr))
					return
				}
			}
			if err != nil {
				session.Mutex.Lock()
				if err.Error() == "EOF" {
					session.Conn.WriteMessage(websocket.TextMessage, []byte("\n\x1b[33m■ 日志已全部输出（Pod可能已完成）\x1b[0m\n"))
				} else {
					session.Conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("\n\x1b[31m✗ 日志流异常结束: %v\x1b[0m\n", err)))
				}
				session.Mutex.Unlock()
				return
			}
		}
	}
}

func (c *PodLogController) StopStream(ctx *gin.Context) {
	sessionID := ctx.Query("session_id")
	if sessionID == "" {
		ctx.JSON(http.StatusOK, response.Fail("session_id不能为空"))
		return
	}

	if session, ok := logSessions.Load(sessionID); ok {
		s := session.(*LogSession)
		if s.Cancel != nil {
			s.Cancel()
		}
		logSessions.Delete(sessionID)
		ctx.JSON(http.StatusOK, response.Success(nil))
		return
	}

	ctx.JSON(http.StatusOK, response.Fail("session不存在"))
}
