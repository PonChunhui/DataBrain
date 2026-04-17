package service

import (
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"devops-backend/utils"
)

type ClusterStats struct {
	Nodes          []NodeInfo `json:"nodes"`
	PodCount       int        `json:"pod_count"`
	RunningPods    int        `json:"running_pods"`
	PendingPods    int        `json:"pending_pods"`
	FailedPods     int        `json:"failed_pods"`
	NamespaceCount int        `json:"namespace_count"`
}

type NodeInfo struct {
	Name        string `json:"name"`
	IP          string `json:"ip"`
	Role        string `json:"role"`
	Status      string `json:"status"`
	CpuCapacity string `json:"cpu_capacity"`
	MemCapacity string `json:"mem_capacity"`
	PodCount    int    `json:"pod_count"`
}

type StatsService struct{}

func convertMemoryToGB(mem resource.Quantity) string {
	bytes := mem.Value()
	gb := float64(bytes) / (1024 * 1024 * 1024)
	if gb < 1 {
		mb := float64(bytes) / (1024 * 1024)
		return fmt.Sprintf("%.0fMB", mb)
	}
	return fmt.Sprintf("%.1fG", gb)
}

func (s *StatsService) GetClusterStats(clusterID uint) (*ClusterStats, error) {
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

	nodes, err := client.Clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	nsList, err := client.Clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	namespaceCount := 0
	if err == nil {
		namespaceCount = len(nsList.Items)
	}

	allPods, err := client.Clientset.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	nodePodCounts := make(map[string]int)
	for _, pod := range allPods.Items {
		if pod.Spec.NodeName != "" {
			nodePodCounts[pod.Spec.NodeName]++
		}
	}

	stats := &ClusterStats{
		PodCount:       len(allPods.Items),
		NamespaceCount: namespaceCount,
		RunningPods:    0,
		PendingPods:    0,
		FailedPods:     0,
	}

	for _, pod := range allPods.Items {
		switch pod.Status.Phase {
		case corev1.PodRunning:
			stats.RunningPods++
		case corev1.PodPending:
			stats.PendingPods++
		case corev1.PodFailed:
			stats.FailedPods++
		}
	}

	nodeInfos := make([]NodeInfo, 0, len(nodes.Items))
	for _, node := range nodes.Items {
		status := "Ready"
		for _, condition := range node.Status.Conditions {
			if condition.Type == corev1.NodeReady && condition.Status != corev1.ConditionTrue {
				status = "NotReady"
			}
		}

		nodeIP := ""
		for _, addr := range node.Status.Addresses {
			if addr.Type == corev1.NodeInternalIP {
				nodeIP = addr.Address
				break
			}
		}

		nodeRole := "Worker"
		if _, ok := node.Labels["node-role.kubernetes.io/master"]; ok {
			nodeRole = "Master"
		} else if _, ok := node.Labels["node-role.kubernetes.io/control-plane"]; ok {
			nodeRole = "Control-Plane"
		} else if _, ok := node.Labels["node-role.kubernetes.io/work"]; ok {
			nodeRole = "Worker"
		}

		cpuCapacity := node.Status.Capacity[corev1.ResourceCPU]
		memCapacity := node.Status.Capacity[corev1.ResourceMemory]

		nodeInfos = append(nodeInfos, NodeInfo{
			Name:        node.Name,
			IP:          nodeIP,
			Role:        nodeRole,
			Status:      status,
			CpuCapacity: cpuCapacity.String(),
			MemCapacity: convertMemoryToGB(memCapacity),
			PodCount:    nodePodCounts[node.Name],
		})
	}

	stats.Nodes = nodeInfos

	return stats, nil
}
