package service

import (
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"devops-backend/utils"
)

type K8sNodeService struct{}

type NodeDetail struct {
	Name              string            `json:"name"`
	IP                string            `json:"ip"`
	InternalIP        string            `json:"internal_ip"`
	ExternalIP        string            `json:"external_ip"`
	Role              string            `json:"role"`
	Status            string            `json:"status"`
	Ready             bool              `json:"ready"`
	CPUCapacity       int64             `json:"cpu_capacity"`
	MemoryCapacity    int64             `json:"memory_capacity"` // bytes
	PodCapacity       int64             `json:"pod_capacity"`
	CPUAllocatable    int64             `json:"cpu_allocatable"`
	MemoryAllocatable int64             `json:"memory_allocatable"` // bytes
	PodAllocatable    int64             `json:"pod_allocatable"`
	Labels            map[string]string `json:"labels"`
	Annotations       map[string]string `json:"annotations"`
	Conditions        []NodeCondition   `json:"conditions"`
	Taints            []NodeTaint       `json:"taints"`
	CreatedAt         time.Time         `json:"created_at"`
	KernelVersion     string            `json:"kernel_version"`
	OSImage           string            `json:"os_image"`
	ContainerRuntime  string            `json:"container_runtime"`
	KubeletVersion    string            `json:"kubelet_version"`
	KubeProxyVersion  string            `json:"kube_proxy_version"`
	OperatingSystem   string            `json:"operating_system"`
	Architecture      string            `json:"architecture"`
}

type NodeCondition struct {
	Type    string `json:"type"`
	Status  string `json:"status"`
	Reason  string `json:"reason"`
	Message string `json:"message"`
}

type NodeTaint struct {
	Key    string `json:"key"`
	Value  string `json:"value"`
	Effect string `json:"effect"`
}

type NodePodInfo struct {
	Name       string    `json:"name"`
	Namespace  string    `json:"namespace"`
	Status     string    `json:"status"`
	PodIP      string    `json:"pod_ip"`
	Containers int       `json:"containers"`
	Restarts   int32     `json:"restarts"`
	CreatedAt  time.Time `json:"created_at"`
	CPURequest int64     `json:"cpu_request"` // millicores
	MemRequest int64     `json:"mem_request"` // bytes
}

func (s *K8sNodeService) GetNodeDetail(clusterID uint, nodeName string) (*NodeDetail, error) {
	cluster, err := clusterService.GetClusterByID(clusterID)
	if err != nil {
		return nil, err
	}

	client, err := utils.GetClusterClientFromKubeconfig(cluster.Kubeconfig)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	node, err := client.Clientset.CoreV1().Nodes().Get(ctx, nodeName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	detail := parseNodeDetail(node)
	return detail, nil
}

func parseNodeDetail(node *corev1.Node) *NodeDetail {
	internalIP := ""
	externalIP := ""
	for _, addr := range node.Status.Addresses {
		if addr.Type == corev1.NodeInternalIP {
			internalIP = addr.Address
		} else if addr.Type == corev1.NodeExternalIP {
			externalIP = addr.Address
		}
	}

	role := "Worker"
	if _, ok := node.Labels["node-role.kubernetes.io/master"]; ok {
		role = "Master"
	} else if _, ok := node.Labels["node-role.kubernetes.io/control-plane"]; ok {
		role = "Control-Plane"
	}

	ready := false
	status := "NotReady"
	for _, cond := range node.Status.Conditions {
		if cond.Type == corev1.NodeReady {
			if cond.Status == corev1.ConditionTrue {
				ready = true
				status = "Ready"
			}
		}
	}

	conditions := make([]NodeCondition, 0)
	for _, cond := range node.Status.Conditions {
		conditions = append(conditions, NodeCondition{
			Type:    string(cond.Type),
			Status:  string(cond.Status),
			Reason:  cond.Reason,
			Message: cond.Message,
		})
	}

	taints := make([]NodeTaint, 0)
	for _, taint := range node.Spec.Taints {
		taints = append(taints, NodeTaint{
			Key:    taint.Key,
			Value:  taint.Value,
			Effect: string(taint.Effect),
		})
	}

	return &NodeDetail{
		Name:              node.Name,
		IP:                internalIP,
		InternalIP:        internalIP,
		ExternalIP:        externalIP,
		Role:              role,
		Status:            status,
		Ready:             ready,
		CPUCapacity:       node.Status.Capacity.Cpu().Value(),
		MemoryCapacity:    node.Status.Capacity.Memory().Value(),
		PodCapacity:       node.Status.Capacity.Pods().Value(),
		CPUAllocatable:    node.Status.Allocatable.Cpu().Value(),
		MemoryAllocatable: node.Status.Allocatable.Memory().Value(),
		PodAllocatable:    node.Status.Allocatable.Pods().Value(),
		Labels:            node.Labels,
		Annotations:       node.Annotations,
		Conditions:        conditions,
		Taints:            taints,
		CreatedAt:         node.CreationTimestamp.Time,
		KernelVersion:     node.Status.NodeInfo.KernelVersion,
		OSImage:           node.Status.NodeInfo.OSImage,
		ContainerRuntime:  node.Status.NodeInfo.ContainerRuntimeVersion,
		KubeletVersion:    node.Status.NodeInfo.KubeletVersion,
		KubeProxyVersion:  node.Status.NodeInfo.KubeProxyVersion,
		OperatingSystem:   node.Status.NodeInfo.OperatingSystem,
		Architecture:      node.Status.NodeInfo.Architecture,
	}
}

func (s *K8sNodeService) GetNodePods(clusterID uint, nodeName string, page, pageSize int) ([]NodePodInfo, int, error) {
	cluster, err := clusterService.GetClusterByID(clusterID)
	if err != nil {
		return nil, 0, err
	}

	client, err := utils.GetClusterClientFromKubeconfig(cluster.Kubeconfig)
	if err != nil {
		return nil, 0, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	podList, err := client.Clientset.CoreV1().Pods("").List(ctx, metav1.ListOptions{
		FieldSelector: fmt.Sprintf("spec.nodeName=%s", nodeName),
	})
	if err != nil {
		return nil, 0, err
	}

	allPods := make([]NodePodInfo, 0, len(podList.Items))
	for _, pod := range podList.Items {
		var restarts int32 = 0
		containerCount := len(pod.Status.ContainerStatuses)
		for _, cs := range pod.Status.ContainerStatuses {
			restarts += cs.RestartCount
		}

		cpuRequest := int64(0)
		memRequest := int64(0)
		for _, container := range pod.Spec.Containers {
			if container.Resources.Requests != nil {
				if cpu, ok := container.Resources.Requests[corev1.ResourceCPU]; ok {
					cpuRequest += cpu.MilliValue()
				}
				if mem, ok := container.Resources.Requests[corev1.ResourceMemory]; ok {
					memRequest += mem.Value()
				}
			}
		}

		podStatus := getPodRealStatus(pod)

		allPods = append(allPods, NodePodInfo{
			Name:       pod.Name,
			Namespace:  pod.Namespace,
			Status:     podStatus,
			PodIP:      pod.Status.PodIP,
			Containers: containerCount,
			Restarts:   restarts,
			CreatedAt:  pod.CreationTimestamp.Time,
			CPURequest: cpuRequest,
			MemRequest: memRequest,
		})
	}

	total := len(allPods)
	start := (page - 1) * pageSize
	end := start + pageSize
	if start >= total {
		return []NodePodInfo{}, total, nil
	}
	if end > total {
		end = total
	}

	return allPods[start:end], total, nil
}

func (s *K8sNodeService) GetNodeEvents(clusterID uint, nodeName string) ([]map[string]interface{}, error) {
	cluster, err := clusterService.GetClusterByID(clusterID)
	if err != nil {
		return nil, err
	}

	client, err := utils.GetClusterClientFromKubeconfig(cluster.Kubeconfig)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	events, err := client.Clientset.CoreV1().Events("").List(ctx, metav1.ListOptions{
		FieldSelector: fmt.Sprintf("involvedObject.name=%s,involvedObject.kind=Node", nodeName),
	})
	if err != nil {
		return nil, err
	}

	eventList := make([]map[string]interface{}, 0)
	for _, event := range events.Items {
		eventList = append(eventList, map[string]interface{}{
			"type":            event.Type,
			"reason":          event.Reason,
			"message":         event.Message,
			"count":           event.Count,
			"last_timestamp":  event.LastTimestamp.Time,
			"first_timestamp": event.FirstTimestamp.Time,
			"source":          event.Source.Component,
		})
	}

	return eventList, nil
}
