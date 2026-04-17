package service

import (
	"context"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"devops-backend/utils"
)

type PodInfo struct {
	Name       string            `json:"name"`
	Namespace  string            `json:"namespace"`
	Status     string            `json:"status"`
	PodIP      string            `json:"pod_ip"`
	NodeName   string            `json:"node_name"`
	Restarts   int32             `json:"restarts"`
	Containers []ContainerInfo   `json:"containers"`
	Labels     map[string]string `json:"labels"`
	CreatedAt  time.Time         `json:"created_at"`
}

type ContainerInfo struct {
	Name   string `json:"name"`
	Image  string `json:"image"`
	Status string `json:"status"`
	Ready  bool   `json:"ready"`
}

type PodService struct{}

func (s *PodService) GetPods(clusterID uint, namespace string) ([]PodInfo, error) {
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

	podList, err := client.Clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	pods := make([]PodInfo, 0, len(podList.Items))
	for _, pod := range podList.Items {
		containers := make([]ContainerInfo, 0)
		for _, container := range pod.Spec.Containers {
			containerStatus := getContainerStatus(pod.Status.ContainerStatuses, container.Name)
			containers = append(containers, ContainerInfo{
				Name:   container.Name,
				Image:  container.Image,
				Status: containerStatus,
				Ready:  isContainerReady(pod.Status.ContainerStatuses, container.Name),
			})
		}

		var restarts int32 = 0
		for _, cs := range pod.Status.ContainerStatuses {
			restarts += cs.RestartCount
		}

		pods = append(pods, PodInfo{
			Name:       pod.Name,
			Namespace:  pod.Namespace,
			Status:     string(pod.Status.Phase),
			PodIP:      pod.Status.PodIP,
			NodeName:   pod.Spec.NodeName,
			Restarts:   restarts,
			Containers: containers,
			Labels:     pod.Labels,
			CreatedAt:  pod.CreationTimestamp.Time,
		})
	}

	return pods, nil
}

func (s *PodService) GetPod(clusterID uint, namespace, name string) (*corev1.Pod, error) {
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

	pod, err := client.Clientset.CoreV1().Pods(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return pod, nil
}

func (s *PodService) GetPodLogs(clusterID uint, namespace, podName, containerName string, tailLines int64) (string, error) {
	cluster, err := clusterService.GetClusterByID(clusterID)
	if err != nil {
		return "", err
	}

	client, err := utils.GetClusterClientFromKubeconfig(cluster.Kubeconfig)
	if err != nil {
		return "", err
	}

	opts := &corev1.PodLogOptions{
		Container: containerName,
		TailLines: &tailLines,
	}

	return utils.GetPodLogs(client, namespace, podName, opts)
}

func (s *PodService) GetPodEvents(clusterID uint, namespace, podName string) ([]corev1.Event, error) {
	cluster, err := clusterService.GetClusterByID(clusterID)
	if err != nil {
		return nil, err
	}

	client, err := utils.GetClusterClientFromKubeconfig(cluster.Kubeconfig)
	if err != nil {
		return nil, err
	}

	return utils.GetPodEvents(client, namespace, podName)
}

func (s *PodService) DeletePod(clusterID uint, namespace, name string) error {
	cluster, err := clusterService.GetClusterByID(clusterID)
	if err != nil {
		return err
	}

	client, err := utils.GetClusterClientFromKubeconfig(cluster.Kubeconfig)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	return client.Clientset.CoreV1().Pods(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

func getContainerStatus(statuses []corev1.ContainerStatus, name string) string {
	for _, status := range statuses {
		if status.Name == name {
			if status.State.Running != nil {
				return "Running"
			}
			if status.State.Waiting != nil {
				return "Waiting: " + status.State.Waiting.Reason
			}
			if status.State.Terminated != nil {
				return "Terminated: " + status.State.Terminated.Reason
			}
		}
	}
	return "Unknown"
}

func isContainerReady(statuses []corev1.ContainerStatus, name string) bool {
	for _, status := range statuses {
		if status.Name == name {
			return status.Ready
		}
	}
	return false
}

func (s *PodService) GetPodDetail(clusterID uint, namespace, name string) (map[string]interface{}, error) {
	pod, err := s.GetPod(clusterID, namespace, name)
	if err != nil {
		return nil, err
	}

	containers := make([]map[string]interface{}, 0)
	for _, container := range pod.Spec.Containers {
		containerStatus := getContainerStatusMap(pod.Status.ContainerStatuses, container.Name)
		envVars := make([]map[string]interface{}, 0)
		for _, env := range container.Env {
			value := env.Value
			if env.ValueFrom != nil {
				if env.ValueFrom.SecretKeyRef != nil {
					value = "[from secret: " + env.ValueFrom.SecretKeyRef.Name + "/" + env.ValueFrom.SecretKeyRef.Key + "]"
				} else if env.ValueFrom.ConfigMapKeyRef != nil {
					value = "[from configmap: " + env.ValueFrom.ConfigMapKeyRef.Name + "/" + env.ValueFrom.ConfigMapKeyRef.Key + "]"
				} else if env.ValueFrom.FieldRef != nil {
					value = "[from field: " + env.ValueFrom.FieldRef.FieldPath + "]"
				}
			}
			envVars = append(envVars, map[string]interface{}{
				"name":  env.Name,
				"value": value,
			})
		}

		ports := make([]map[string]interface{}, 0)
		for _, port := range container.Ports {
			ports = append(ports, map[string]interface{}{
				"name":           port.Name,
				"container_port": port.ContainerPort,
				"protocol":       string(port.Protocol),
			})
		}

		containers = append(containers, map[string]interface{}{
			"name":      container.Name,
			"image":     container.Image,
			"status":    containerStatus,
			"env":       envVars,
			"ports":     ports,
			"resources": container.Resources,
			"ready":     isContainerReady(pod.Status.ContainerStatuses, container.Name),
			"restarts":  getContainerRestarts(pod.Status.ContainerStatuses, container.Name),
		})
	}

	initContainers := make([]map[string]interface{}, 0)
	for _, container := range pod.Spec.InitContainers {
		initContainers = append(initContainers, map[string]interface{}{
			"name":  container.Name,
			"image": container.Image,
		})
	}

	volumeMounts := make(map[string][]map[string]interface{})
	for _, container := range pod.Spec.Containers {
		for _, mount := range container.VolumeMounts {
			if volumeMounts[mount.Name] == nil {
				volumeMounts[mount.Name] = make([]map[string]interface{}, 0)
			}
			volumeMounts[mount.Name] = append(volumeMounts[mount.Name], map[string]interface{}{
				"container":  container.Name,
				"mount_path": mount.MountPath,
				"sub_path":   mount.SubPath,
				"read_only":  mount.ReadOnly,
			})
		}
	}

	volumes := make([]map[string]interface{}, 0)
	for _, vol := range pod.Spec.Volumes {
		mounts := volumeMounts[vol.Name]
		if mounts == nil {
			mounts = []map[string]interface{}{}
		}
		volInfo := map[string]interface{}{
			"name":   vol.Name,
			"mounts": mounts,
		}
		if vol.PersistentVolumeClaim != nil {
			volInfo["type"] = "PVC"
			volInfo["source"] = vol.PersistentVolumeClaim.ClaimName
		} else if vol.ConfigMap != nil {
			volInfo["type"] = "ConfigMap"
			volInfo["source"] = vol.ConfigMap.Name
		} else if vol.Secret != nil {
			volInfo["type"] = "Secret"
			volInfo["source"] = vol.Secret.SecretName
		} else if vol.EmptyDir != nil {
			volInfo["type"] = "EmptyDir"
			volInfo["source"] = "临时存储"
		} else if vol.HostPath != nil {
			volInfo["type"] = "HostPath"
			volInfo["source"] = vol.HostPath.Path
		} else {
			volInfo["type"] = "Other"
			volInfo["source"] = "-"
		}
		volumes = append(volumes, volInfo)
	}

	podIP := pod.Status.PodIP
	if podIP == "" {
		podIP = "-"
	}

	hostIP := pod.Status.HostIP
	if hostIP == "" {
		hostIP = "-"
	}

	detail := map[string]interface{}{
		"name":            pod.Name,
		"namespace":       pod.Namespace,
		"status":          string(pod.Status.Phase),
		"pod_ip":          podIP,
		"host_ip":         hostIP,
		"node_name":       pod.Spec.NodeName,
		"labels":          pod.Labels,
		"annotations":     pod.Annotations,
		"created_at":      pod.CreationTimestamp.Time,
		"containers":      containers,
		"init_containers": initContainers,
		"volumes":         volumes,
		"restarts":        getTotalRestarts(pod.Status.ContainerStatuses),
		"service_account": pod.Spec.ServiceAccountName,
		"priority_class":  pod.Spec.PriorityClassName,
		"qos_class":       getPodQOSClass(pod),
		"scheduler_name":  pod.Spec.SchedulerName,
		"tolerations":     getTolerations(pod.Spec.Tolerations),
		"affinity":        getAffinity(pod.Spec.Affinity),
		"nominated_node":  pod.Status.NominatedNodeName,
		"start_time":      pod.Status.StartTime,
		"message":         pod.Status.Message,
		"reason":          pod.Status.Reason,
	}

	return detail, nil
}

func (s *PodService) GetPodMetrics(clusterID uint, namespace, name string) ([]map[string]interface{}, error) {
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

	pod, err := client.Clientset.CoreV1().Pods(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	metrics := make([]map[string]interface{}, 0)
	for _, container := range pod.Spec.Containers {
		cpuRequest := ""
		cpuLimit := ""
		memRequest := ""
		memLimit := ""

		if container.Resources.Requests != nil {
			if cpu, ok := container.Resources.Requests["cpu"]; ok {
				cpuRequest = cpu.String()
			}
			if mem, ok := container.Resources.Requests["memory"]; ok {
				memRequest = mem.String()
			}
		}
		if container.Resources.Limits != nil {
			if cpu, ok := container.Resources.Limits["cpu"]; ok {
				cpuLimit = cpu.String()
			}
			if mem, ok := container.Resources.Limits["memory"]; ok {
				memLimit = mem.String()
			}
		}

		metrics = append(metrics, map[string]interface{}{
			"name":        container.Name,
			"cpu_request": cpuRequest,
			"cpu_limit":   cpuLimit,
			"mem_request": memRequest,
			"mem_limit":   memLimit,
		})
	}

	return metrics, nil
}

func getContainerStatusMap(statuses []corev1.ContainerStatus, name string) map[string]interface{} {
	for _, status := range statuses {
		if status.Name == name {
			state := map[string]interface{}{}
			if status.State.Running != nil {
				state["status"] = "Running"
				state["started_at"] = status.State.Running.StartedAt.Time
			} else if status.State.Waiting != nil {
				state["status"] = "Waiting"
				state["reason"] = status.State.Waiting.Reason
				state["message"] = status.State.Waiting.Message
			} else if status.State.Terminated != nil {
				state["status"] = "Terminated"
				state["reason"] = status.State.Terminated.Reason
				state["exit_code"] = status.State.Terminated.ExitCode
			} else {
				state["status"] = "Unknown"
			}
			return state
		}
	}
	return map[string]interface{}{"status": "Unknown"}
}

func getContainerRestarts(statuses []corev1.ContainerStatus, name string) int32 {
	for _, status := range statuses {
		if status.Name == name {
			return status.RestartCount
		}
	}
	return 0
}

func getTotalRestarts(statuses []corev1.ContainerStatus) int32 {
	var total int32 = 0
	for _, status := range statuses {
		total += status.RestartCount
	}
	return total
}

func getPodQOSClass(pod *corev1.Pod) string {
	return string(pod.Status.QOSClass)
}

func getTolerations(tolerations []corev1.Toleration) []map[string]interface{} {
	result := make([]map[string]interface{}, 0)
	for _, t := range tolerations {
		result = append(result, map[string]interface{}{
			"key":      t.Key,
			"operator": string(t.Operator),
			"value":    t.Value,
			"effect":   string(t.Effect),
		})
	}
	return result
}

func getAffinity(affinity *corev1.Affinity) map[string]interface{} {
	if affinity == nil {
		return nil
	}

	result := map[string]interface{}{}

	if affinity.NodeAffinity != nil {
		nodeAff := map[string]interface{}{}
		if affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution != nil {
			nodeAff["required"] = affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms
		}
		if len(affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution) > 0 {
			nodeAff["preferred"] = affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution
		}
		result["node_affinity"] = nodeAff
	}

	if affinity.PodAffinity != nil {
		podAff := map[string]interface{}{
			"required":  affinity.PodAffinity.RequiredDuringSchedulingIgnoredDuringExecution,
			"preferred": affinity.PodAffinity.PreferredDuringSchedulingIgnoredDuringExecution,
		}
		result["pod_affinity"] = podAff
	}

	if affinity.PodAntiAffinity != nil {
		podAntiAff := map[string]interface{}{
			"required":  affinity.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution,
			"preferred": affinity.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution,
		}
		result["pod_antiaffinity"] = podAntiAff
	}

	return result
}
