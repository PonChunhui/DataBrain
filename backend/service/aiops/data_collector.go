package aiops

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"devops-backend/global"
	"devops-backend/model"
	"devops-backend/service"
	"devops-backend/utils"
	"go.uber.org/zap"
)

type ClusterServiceRef struct{}

func (s *ClusterServiceRef) GetClusterByID(id uint) (*model.K8sCluster, error) {
	var cluster model.K8sCluster
	err := global.GVA_DB.First(&cluster, id).Error
	if err != nil {
		return nil, err
	}
	return &cluster, nil
}

var clusterService = &ClusterServiceRef{}
var prometheusService = &service.PrometheusService{}

type DataCollector struct{}

type DiagnosticInput struct {
	ResourceType string                   `json:"resource_type"`
	ResourceInfo map[string]interface{}   `json:"resource_info"`
	Events       []map[string]interface{} `json:"events"`
	Logs         map[string]interface{}   `json:"logs,omitempty"`
	NodeInfo     map[string]interface{}   `json:"node_info,omitempty"`
	Pods         []map[string]interface{} `json:"pods,omitempty"`
	Metrics      map[string]interface{}   `json:"metrics,omitempty"`
	Timestamp    string                   `json:"timestamp"`
}

func NewDataCollector() *DataCollector {
	return &DataCollector{}
}

func (c *DataCollector) Collect(ctx context.Context, clusterID uint, namespace, resourceType, resourceName string) (*DiagnosticInput, error) {
	cluster, err := clusterService.GetClusterByID(clusterID)
	if err != nil {
		return nil, fmt.Errorf("get cluster failed: %v", err)
	}

	client, err := utils.GetClusterClientFromKubeconfig(cluster.Kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("get k8s client failed: %v", err)
	}

	input := &DiagnosticInput{
		ResourceType: resourceType,
		Timestamp:    time.Now().Format(time.RFC3339),
	}

	switch resourceType {
	case "pod":
		err = c.collectPodData(ctx, client, clusterID, namespace, resourceName, input)
	case "deployment":
		err = c.collectDeploymentData(ctx, client, clusterID, namespace, resourceName, input)
	case "service":
		err = c.collectServiceData(ctx, client, namespace, resourceName, input)
	case "ingress":
		err = c.collectIngressData(ctx, client, namespace, resourceName, input)
	case "node":
		err = c.collectNodeData(ctx, client, clusterID, resourceName, input)
	default:
		return nil, fmt.Errorf("unsupported resource type: %s", resourceType)
	}

	if err != nil {
		return nil, err
	}

	return input, nil
}

func (c *DataCollector) collectPodData(ctx context.Context, client *utils.ClusterClient, clusterID uint, namespace, name string, input *DiagnosticInput) error {
	pod, err := client.Clientset.CoreV1().Pods(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("get pod failed: %v", err)
	}

	input.ResourceInfo = c.buildPodInfo(pod)

	events, err := client.Clientset.CoreV1().Events(namespace).List(ctx, metav1.ListOptions{
		FieldSelector: fmt.Sprintf("involvedObject.name=%s,involvedObject.kind=Pod", name),
	})
	if err != nil {
		return fmt.Errorf("get events failed: %v", err)
	}

	input.Events = c.buildEvents(events.Items)

	if pod.Spec.NodeName != "" {
		node, err := client.Clientset.CoreV1().Nodes().Get(ctx, pod.Spec.NodeName, metav1.GetOptions{})
		if err == nil {
			input.NodeInfo = c.buildNodeInfo(node)
		}
	}

	for _, container := range pod.Spec.Containers {
		if container.Name != "" {
			logBytes, err := client.Clientset.CoreV1().Pods(namespace).GetLogs(name, &corev1.PodLogOptions{
				Container: container.Name,
				TailLines: int64Ptr(100),
			}).DoRaw(ctx)
			if err == nil && len(logBytes) > 0 {
				input.Logs = map[string]interface{}{
					"container":  container.Name,
					"tail_lines": 100,
					"content":    string(logBytes),
				}
				break
			}
		}
	}

	cpuLimit, memLimit := c.getPodResourceLimits(pod)
	metrics, err := prometheusService.GetAggregatedPodMetrics(clusterID, namespace, name, 30, cpuLimit, memLimit)
	if err != nil {
		global.GVA_LOG.Warn("获取Pod Prometheus指标失败", zap.Error(err))
	} else if metrics != nil {
		input.Metrics = map[string]interface{}{
			"cpu":             metrics.CPU,
			"memory":          metrics.Memory,
			"network":         metrics.Network,
			"collection_time": "30m",
		}
	}

	return nil
}

func (c *DataCollector) collectDeploymentData(ctx context.Context, client *utils.ClusterClient, clusterID uint, namespace, name string, input *DiagnosticInput) error {
	deploy, err := client.Clientset.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("get deployment failed: %v", err)
	}

	input.ResourceInfo = c.buildDeploymentInfo(deploy)

	events, err := client.Clientset.CoreV1().Events(namespace).List(ctx, metav1.ListOptions{
		FieldSelector: fmt.Sprintf("involvedObject.name=%s,involvedObject.kind=Deployment", name),
	})
	if err != nil {
		return fmt.Errorf("get events failed: %v", err)
	}

	input.Events = c.buildEvents(events.Items)

	podList, err := client.Clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("app=%s", name),
	})
	if err == nil {
		for _, pod := range podList.Items {
			input.Pods = append(input.Pods, map[string]interface{}{
				"name":      pod.Name,
				"status":    string(pod.Status.Phase),
				"node_name": pod.Spec.NodeName,
				"restarts":  getPodRestarts(&pod),
			})
		}

		if len(podList.Items) > 0 && clusterID > 0 {
			totalCpuLimit := 0.0
			totalMemLimit := 0.0
			for _, pod := range podList.Items {
				cpu, mem := c.getPodResourceLimits(&pod)
				totalCpuLimit += cpu
				totalMemLimit += mem
			}

			if totalCpuLimit > 0 || totalMemLimit > 0 {
				podMetrics := []map[string]interface{}{}
				for _, pod := range podList.Items {
					cpu, mem := c.getPodResourceLimits(&pod)
					m, err := prometheusService.GetAggregatedPodMetrics(clusterID, namespace, pod.Name, 30, cpu, mem)
					if err != nil {
						global.GVA_LOG.Warn("获取Pod指标失败", zap.String("pod", pod.Name), zap.Error(err))
						continue
					}
					if m != nil {
						podMetrics = append(podMetrics, map[string]interface{}{
							"pod_name": pod.Name,
							"cpu":      m.CPU,
							"memory":   m.Memory,
						})
					}
				}

				if len(podMetrics) > 0 {
					input.Metrics = map[string]interface{}{
						"pod_metrics":        podMetrics,
						"total_cpu_limit":    totalCpuLimit,
						"total_mem_limit_mb": totalMemLimit / 1024 / 1024,
						"collection_time":    "30m",
					}
				}
			}
		}
	}

	return nil
}

func (c *DataCollector) getPodResourceLimits(pod *corev1.Pod) (float64, float64) {
	var cpuLimit float64 = 0
	var memLimit float64 = 0

	for _, container := range pod.Spec.Containers {
		if container.Resources.Limits != nil {
			if cpu, ok := container.Resources.Limits[corev1.ResourceCPU]; ok {
				cpuLimit += float64(cpu.MilliValue()) / 1000
			}
			if mem, ok := container.Resources.Limits[corev1.ResourceMemory]; ok {
				memLimit += float64(mem.Value())
			}
		}
	}

	return cpuLimit, memLimit
}

func (c *DataCollector) collectServiceData(ctx context.Context, client *utils.ClusterClient, namespace, name string, input *DiagnosticInput) error {
	svc, err := client.Clientset.CoreV1().Services(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("get service failed: %v", err)
	}

	input.ResourceInfo = c.buildServiceInfo(svc)

	events, err := client.Clientset.CoreV1().Events(namespace).List(ctx, metav1.ListOptions{
		FieldSelector: fmt.Sprintf("involvedObject.name=%s,involvedObject.kind=Service", name),
	})
	if err != nil {
		return fmt.Errorf("get events failed: %v", err)
	}

	input.Events = c.buildEvents(events.Items)

	endpoints, err := client.Clientset.CoreV1().Endpoints(namespace).Get(ctx, name, metav1.GetOptions{})
	if err == nil {
		input.ResourceInfo["endpoints"] = c.buildEndpointsInfo(endpoints)
	}

	return nil
}

func (c *DataCollector) collectIngressData(ctx context.Context, client *utils.ClusterClient, namespace, name string, input *DiagnosticInput) error {
	ingress, err := client.Clientset.NetworkingV1().Ingresses(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("get ingress failed: %v", err)
	}

	input.ResourceInfo = c.buildIngressInfo(ingress)

	events, err := client.Clientset.CoreV1().Events(namespace).List(ctx, metav1.ListOptions{
		FieldSelector: fmt.Sprintf("involvedObject.name=%s,involvedObject.kind=Ingress", name),
	})
	if err != nil {
		return fmt.Errorf("get events failed: %v", err)
	}

	input.Events = c.buildEvents(events.Items)

	return nil
}

func (c *DataCollector) collectNodeData(ctx context.Context, client *utils.ClusterClient, clusterID uint, name string, input *DiagnosticInput) error {
	node, err := client.Clientset.CoreV1().Nodes().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("get node failed: %v", err)
	}

	input.ResourceInfo = c.buildNodeInfo(node)

	events, err := client.Clientset.CoreV1().Events("").List(ctx, metav1.ListOptions{
		FieldSelector: fmt.Sprintf("involvedObject.name=%s,involvedObject.kind=Node", name),
	})
	if err != nil {
		return fmt.Errorf("get events failed: %v", err)
	}

	input.Events = c.buildEvents(events.Items)

	metrics, err := prometheusService.GetNodeMetrics(clusterID, name, 30)
	if err != nil {
		global.GVA_LOG.Warn("获取节点 Prometheus指标失败", zap.Error(err))
	} else if metrics != nil && len(metrics) > 0 {
		input.Metrics = metrics
	}

	return nil
}

func (c *DataCollector) buildPodInfo(pod *corev1.Pod) map[string]interface{} {
	containers := make([]map[string]interface{}, 0)
	for _, cs := range pod.Status.ContainerStatuses {
		container := map[string]interface{}{
			"name":     cs.Name,
			"image":    cs.Image,
			"ready":    cs.Ready,
			"restarts": cs.RestartCount,
		}
		if cs.State.Waiting != nil {
			container["status"] = "Waiting"
			container["reason"] = cs.State.Waiting.Reason
			container["message"] = cs.State.Waiting.Message
		} else if cs.State.Running != nil {
			container["status"] = "Running"
			container["started_at"] = cs.State.Running.StartedAt.Format(time.RFC3339)
		} else if cs.State.Terminated != nil {
			container["status"] = "Terminated"
			container["reason"] = cs.State.Terminated.Reason
			container["exit_code"] = cs.State.Terminated.ExitCode
		} else {
			container["status"] = "Unknown"
		}
		containers = append(containers, container)
	}

	return map[string]interface{}{
		"name":        pod.Name,
		"namespace":   pod.Namespace,
		"status":      string(pod.Status.Phase),
		"pod_ip":      pod.Status.PodIP,
		"node_name":   pod.Spec.NodeName,
		"created_at":  pod.CreationTimestamp.Format(time.RFC3339),
		"restarts":    getPodRestarts(pod),
		"containers":  containers,
		"labels":      pod.Labels,
		"annotations": pod.Annotations,
	}
}

func (c *DataCollector) buildDeploymentInfo(deploy *appsv1.Deployment) map[string]interface{} {
	conditions := make([]map[string]interface{}, 0)
	for _, cond := range deploy.Status.Conditions {
		conditions = append(conditions, map[string]interface{}{
			"type":    string(cond.Type),
			"status":  string(cond.Status),
			"reason":  cond.Reason,
			"message": cond.Message,
		})
	}

	containers := make([]map[string]interface{}, 0)
	for _, container := range deploy.Spec.Template.Spec.Containers {
		containers = append(containers, map[string]interface{}{
			"name":  container.Name,
			"image": container.Image,
		})
	}

	replicas := int32(0)
	if deploy.Spec.Replicas != nil {
		replicas = *deploy.Spec.Replicas
	}

	return map[string]interface{}{
		"name":               deploy.Name,
		"namespace":          deploy.Namespace,
		"replicas":           replicas,
		"ready_replicas":     deploy.Status.ReadyReplicas,
		"available_replicas": deploy.Status.AvailableReplicas,
		"updated_replicas":   deploy.Status.UpdatedReplicas,
		"conditions":         conditions,
		"containers":         containers,
		"labels":             deploy.Labels,
		"created_at":         deploy.CreationTimestamp.Format(time.RFC3339),
	}
}

func (c *DataCollector) buildServiceInfo(svc *corev1.Service) map[string]interface{} {
	ports := make([]map[string]interface{}, 0)
	for _, port := range svc.Spec.Ports {
		ports = append(ports, map[string]interface{}{
			"name":       port.Name,
			"port":       port.Port,
			"targetPort": port.TargetPort.String(),
			"protocol":   string(port.Protocol),
		})
	}

	return map[string]interface{}{
		"name":       svc.Name,
		"namespace":  svc.Namespace,
		"type":       string(svc.Spec.Type),
		"cluster_ip": svc.Spec.ClusterIP,
		"ports":      ports,
		"labels":     svc.Labels,
		"created_at": svc.CreationTimestamp.Format(time.RFC3339),
	}
}

func (c *DataCollector) buildIngressInfo(ingress *networkingv1.Ingress) map[string]interface{} {
	rules := make([]map[string]interface{}, 0)
	for _, rule := range ingress.Spec.Rules {
		paths := make([]map[string]interface{}, 0)
		if rule.HTTP != nil {
			for _, path := range rule.HTTP.Paths {
				paths = append(paths, map[string]interface{}{
					"path":         path.Path,
					"path_type":    string(*path.PathType),
					"service_name": path.Backend.Service.Name,
					"service_port": path.Backend.Service.Port.Number,
				})
			}
		}
		rules = append(rules, map[string]interface{}{
			"host":  rule.Host,
			"paths": paths,
		})
	}

	ingressClass := ""
	if ingress.Spec.IngressClassName != nil {
		ingressClass = *ingress.Spec.IngressClassName
	}

	return map[string]interface{}{
		"name":          ingress.Name,
		"namespace":     ingress.Namespace,
		"ingress_class": ingressClass,
		"rules":         rules,
		"labels":        ingress.Labels,
		"created_at":    ingress.CreationTimestamp.Format(time.RFC3339),
	}
}

func (c *DataCollector) buildNodeInfo(node *corev1.Node) map[string]interface{} {
	conditions := make([]map[string]interface{}, 0)
	for _, cond := range node.Status.Conditions {
		conditions = append(conditions, map[string]interface{}{
			"type":    string(cond.Type),
			"status":  string(cond.Status),
			"reason":  cond.Reason,
			"message": cond.Message,
		})
	}

	return map[string]interface{}{
		"name":       node.Name,
		"status":     getNodeStatus(node),
		"conditions": conditions,
		"labels":     node.Labels,
		"created_at": node.CreationTimestamp.Format(time.RFC3339),
	}
}

func (c *DataCollector) buildEvents(events []corev1.Event) []map[string]interface{} {
	result := make([]map[string]interface{}, 0)
	for _, event := range events {
		result = append(result, map[string]interface{}{
			"type":    event.Type,
			"reason":  event.Reason,
			"message": event.Message,
			"time":    event.FirstTimestamp.Format(time.RFC3339),
			"count":   event.Count,
			"source":  event.Source.Component,
		})
	}
	return result
}

func (c *DataCollector) buildEndpointsInfo(endpoints *corev1.Endpoints) map[string]interface{} {
	addresses := make([]map[string]interface{}, 0)
	for _, subset := range endpoints.Subsets {
		for _, addr := range subset.Addresses {
			addresses = append(addresses, map[string]interface{}{
				"ip":    addr.IP,
				"ready": true,
			})
		}
		for _, addr := range subset.NotReadyAddresses {
			addresses = append(addresses, map[string]interface{}{
				"ip":    addr.IP,
				"ready": false,
			})
		}
	}
	return map[string]interface{}{
		"addresses": addresses,
	}
}

func (c *DataCollector) ToJSON(input *DiagnosticInput) (string, error) {
	data, err := json.MarshalIndent(input, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func int64Ptr(i int) *int64 {
	return &[]int64{int64(i)}[0]
}

func getPodRestarts(pod *corev1.Pod) int32 {
	var restarts int32 = 0
	for _, cs := range pod.Status.ContainerStatuses {
		restarts += cs.RestartCount
	}
	return restarts
}

func getNodeStatus(node *corev1.Node) string {
	for _, cond := range node.Status.Conditions {
		if cond.Type == corev1.NodeReady {
			if cond.Status == corev1.ConditionTrue {
				return "Ready"
			}
		}
	}
	return "NotReady"
}
