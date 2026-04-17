package utils

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

var (
	clusterClients sync.Map
)

type ClusterClient struct {
	Clientset  *kubernetes.Clientset
	Config     *clientcmdapi.Config
	RestConfig *rest.Config
}

func GetClusterClientFromKubeconfig(kubeconfig string) (*ClusterClient, error) {
	key := kubeconfig

	if client, ok := clusterClients.Load(key); ok {
		return client.(*ClusterClient), nil
	}

	tmpFile, err := os.CreateTemp("", "kubeconfig-*.yaml")
	if err != nil {
		return nil, fmt.Errorf("创建临时文件失败: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("写入kubeconfig失败: %v", err)
	}
	tmpFile.Close()

	config, err := clientcmd.LoadFromFile(tmpFile.Name())
	if err != nil {
		return nil, fmt.Errorf("加载kubeconfig失败: %v", err)
	}

	clientConfig := clientcmd.NewDefaultClientConfig(*config, &clientcmd.ConfigOverrides{})
	restConfig, err := clientConfig.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("创建rest配置失败: %v", err)
	}

	restConfig.Timeout = time.Second * 30

	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("创建k8s client失败: %v", err)
	}

	client := &ClusterClient{
		Clientset:  clientset,
		Config:     config,
		RestConfig: restConfig,
	}

	clusterClients.Store(key, client)

	return client, nil
}

func RemoveClusterClient(kubeconfig string) {
	clusterClients.Delete(kubeconfig)
}

func GetLogStreamClient(kubeconfig string) (*ClusterClient, error) {
	tmpFile, err := os.CreateTemp("", "kubeconfig-*.yaml")
	if err != nil {
		return nil, fmt.Errorf("创建临时文件失败: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("写入kubeconfig失败: %v", err)
	}
	tmpFile.Close()

	config, err := clientcmd.LoadFromFile(tmpFile.Name())
	if err != nil {
		return nil, fmt.Errorf("加载kubeconfig失败: %v", err)
	}

	clientConfig := clientcmd.NewDefaultClientConfig(*config, &clientcmd.ConfigOverrides{})
	restConfig, err := clientConfig.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("创建rest配置失败: %v", err)
	}

	restConfig.Timeout = 0

	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("创建k8s client失败: %v", err)
	}

	return &ClusterClient{
		Clientset:  clientset,
		Config:     config,
		RestConfig: restConfig,
	}, nil
}

func TestClusterConnectionFromKubeconfig(kubeconfig string) error {
	client, err := GetClusterClientFromKubeconfig(kubeconfig)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err = client.Clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{Limit: 1})
	if err != nil {
		return fmt.Errorf("无法连接到集群: %v", err)
	}

	return nil
}

func GetPodLogs(client *ClusterClient, namespace, podName string, opts *corev1.PodLogOptions) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req := client.Clientset.CoreV1().Pods(namespace).GetLogs(podName, opts)

	stream, err := req.Stream(ctx)
	if err != nil {
		return "", fmt.Errorf("获取Pod日志失败: %v", err)
	}
	defer stream.Close()

	buf := new(strings.Builder)
	_, err = io.Copy(buf, stream)
	if err != nil {
		return "", fmt.Errorf("读取日志失败: %v", err)
	}

	return buf.String(), nil
}

func GetPodEvents(client *ClusterClient, namespace, podName string) ([]corev1.Event, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	events, err := client.Clientset.CoreV1().Events(namespace).List(ctx, metav1.ListOptions{
		FieldSelector: fmt.Sprintf("involvedObject.name=%s,involvedObject.kind=Pod", podName),
	})
	if err != nil {
		return nil, fmt.Errorf("获取Pod事件失败: %v", err)
	}

	return events.Items, nil
}

func GetCurrentContext(kubeconfig string) (string, error) {
	client, err := GetClusterClientFromKubeconfig(kubeconfig)
	if err != nil {
		return "", err
	}

	if client.Config == nil || len(client.Config.Contexts) == 0 {
		return "", fmt.Errorf("kubeconfig中没有context")
	}

	return client.Config.CurrentContext, nil
}

func GetClusterServer(kubeconfig string) (string, error) {
	client, err := GetClusterClientFromKubeconfig(kubeconfig)
	if err != nil {
		return "", err
	}

	if client.Config == nil || len(client.Config.Clusters) == 0 {
		return "", fmt.Errorf("kubeconfig中没有cluster配置")
	}

	currentContext := client.Config.CurrentContext
	if currentContext == "" {
		return "", fmt.Errorf("kubeconfig中没有设置current-context")
	}

	ctx, exists := client.Config.Contexts[currentContext]
	if !exists {
		return "", fmt.Errorf("context %s 不存在", currentContext)
	}

	cluster, exists := client.Config.Clusters[ctx.Cluster]
	if !exists {
		return "", fmt.Errorf("cluster %s 不存在", ctx.Cluster)
	}

	return cluster.Server, nil
}
