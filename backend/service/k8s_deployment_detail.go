package service

import (
	"context"
	"fmt"
	"strconv"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/yaml"

	"devops-backend/global"
	"devops-backend/model"
	"devops-backend/utils"
	"go.uber.org/zap"
)

type DeploymentDetailService struct{}

type DeploymentDetail struct {
	Name              string            `json:"name"`
	Namespace         string            `json:"namespace"`
	Replicas          int32             `json:"replicas"`
	ReadyReplicas     int32             `json:"ready_replicas"`
	AvailableReplicas int32             `json:"available_replicas"`
	Status            string            `json:"status"`
	Images            []string          `json:"images"`
	Labels            map[string]string `json:"labels"`
	Annotations       map[string]string `json:"annotations"`
	Ports             []PortInfo        `json:"ports"`
	Containers        []ContainerDetail `json:"containers"`
	CreatedAt         time.Time         `json:"created_at"`
}

type PortInfo struct {
	Name       string `json:"name"`
	Protocol   string `json:"protocol"`
	Port       int32  `json:"port"`
	TargetPort string `json:"target_port"`
}

type ContainerDetail struct {
	Name      string                      `json:"name"`
	Image     string                      `json:"image"`
	Ports     []corev1.ContainerPort      `json:"ports"`
	Env       []corev1.EnvVar             `json:"env"`
	Resources corev1.ResourceRequirements `json:"resources"`
}

type PodDetail struct {
	Name      string    `json:"name"`
	Namespace string    `json:"namespace"`
	NodeName  string    `json:"node_name"`
	NodeIP    string    `json:"node_ip"`
	PodIP     string    `json:"pod_ip"`
	Status    string    `json:"status"`
	Restarts  int32     `json:"restarts"`
	CreatedAt time.Time `json:"created_at"`
}

type EventInfo struct {
	Type    string    `json:"type"`
	Reason  string    `json:"reason"`
	Time    time.Time `json:"time"`
	Source  string    `json:"source"`
	Message string    `json:"message"`
}

func (s *DeploymentDetailService) GetDeploymentDetail(clusterID uint, namespace, name string) (*DeploymentDetail, error) {
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

	deploy, err := client.Clientset.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	detail := &DeploymentDetail{
		Name:              deploy.Name,
		Namespace:         deploy.Namespace,
		Replicas:          *deploy.Spec.Replicas,
		ReadyReplicas:     deploy.Status.ReadyReplicas,
		AvailableReplicas: deploy.Status.AvailableReplicas,
		Status:            getDeploymentStatus(deploy, *deploy.Spec.Replicas, deploy.Status.ReadyReplicas),
		Labels:            deploy.Labels,
		Annotations:       deploy.Annotations,
		CreatedAt:         deploy.CreationTimestamp.Time,
	}

	images := make([]string, 0)
	ports := make([]PortInfo, 0)
	containers := make([]ContainerDetail, 0)

	for _, container := range deploy.Spec.Template.Spec.Containers {
		images = append(images, container.Image)

		containerDetail := ContainerDetail{
			Name:      container.Name,
			Image:     container.Image,
			Ports:     container.Ports,
			Env:       container.Env,
			Resources: container.Resources,
		}
		containers = append(containers, containerDetail)

		for _, port := range container.Ports {
			ports = append(ports, PortInfo{
				Name:       port.Name,
				Protocol:   string(port.Protocol),
				Port:       port.ContainerPort,
				TargetPort: strconv.Itoa(int(port.ContainerPort)),
			})
		}
	}

	detail.Images = images
	detail.Ports = ports
	detail.Containers = containers

	return detail, nil
}

func (s *DeploymentDetailService) GetDeploymentPods(clusterID uint, namespace, deploymentName string) ([]PodDetail, error) {
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

	deploy, err := client.Clientset.AppsV1().Deployments(namespace).Get(ctx, deploymentName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	labelSelector := labels.Set(deploy.Spec.Selector.MatchLabels).AsSelectorPreValidated()
	podList, err := client.Clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labelSelector.String(),
	})
	if err != nil {
		return nil, err
	}

	nodeList, err := client.Clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	nodeIPMap := make(map[string]string)
	for _, node := range nodeList.Items {
		for _, addr := range node.Status.Addresses {
			if addr.Type == corev1.NodeInternalIP {
				nodeIPMap[node.Name] = addr.Address
			}
		}
	}

	pods := make([]PodDetail, 0, len(podList.Items))
	for _, pod := range podList.Items {
		var restarts int32 = 0
		for _, cs := range pod.Status.ContainerStatuses {
			restarts += cs.RestartCount
		}

		nodeIP := nodeIPMap[pod.Spec.NodeName]

		podStatus := getPodRealStatus(pod)

		pods = append(pods, PodDetail{
			Name:      pod.Name,
			Namespace: pod.Namespace,
			NodeName:  pod.Spec.NodeName,
			NodeIP:    nodeIP,
			PodIP:     pod.Status.PodIP,
			Status:    podStatus,
			Restarts:  restarts,
			CreatedAt: pod.CreationTimestamp.Time,
		})
	}

	return pods, nil
}

func getPodRealStatus(pod corev1.Pod) string {
	if pod.DeletionTimestamp != nil {
		return "Terminating"
	}

	phase := string(pod.Status.Phase)

	if phase == "Pending" || phase == "Failed" || phase == "Succeeded" || phase == "Unknown" {
		return phase
	}

	for _, condition := range pod.Status.Conditions {
		if condition.Type == corev1.PodScheduled && condition.Status == corev1.ConditionFalse {
			if condition.Reason == "Unschedulable" {
				return "Unschedulable"
			}
			return condition.Reason
		}
		if condition.Type == corev1.ContainersReady && condition.Status == corev1.ConditionFalse {
			if condition.Reason != "" && condition.Reason != "ContainersNotReady" {
				return condition.Reason
			}
			for _, cs := range pod.Status.ContainerStatuses {
				if !cs.Ready {
					if cs.State.Waiting != nil && cs.State.Waiting.Reason != "" {
						return cs.State.Waiting.Reason
					}
					if cs.State.Terminated != nil && cs.State.Terminated.Reason != "" {
						return cs.State.Terminated.Reason
					}
				}
			}
			return "ContainersNotReady"
		}
		if condition.Type == corev1.PodReady && condition.Status == corev1.ConditionFalse {
			if condition.Reason != "" {
				return condition.Reason
			}
			return "NotReady"
		}
	}

	if phase == "Running" {
		allReady := true
		for _, cs := range pod.Status.ContainerStatuses {
			if !cs.Ready {
				allReady = false
				break
			}
		}
		if allReady {
			return "Running"
		}
		return "ContainersNotReady"
	}

	return phase
}

func (s *DeploymentDetailService) GetDeploymentEvents(clusterID uint, namespace, deploymentName string) ([]EventInfo, error) {
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

	events, err := client.Clientset.CoreV1().Events(namespace).List(ctx, metav1.ListOptions{
		FieldSelector: fmt.Sprintf("involvedObject.name=%s,involvedObject.kind=Deployment", deploymentName),
	})
	if err != nil {
		return nil, err
	}

	eventInfos := make([]EventInfo, 0, len(events.Items))
	for _, event := range events.Items {
		source := ""
		if event.Source.Component != "" {
			source = event.Source.Component
		}
		if event.Source.Host != "" {
			source += "/" + event.Source.Host
		}

		eventInfos = append(eventInfos, EventInfo{
			Type:    event.Type,
			Reason:  event.Reason,
			Time:    event.FirstTimestamp.Time,
			Source:  source,
			Message: event.Message,
		})
	}

	return eventInfos, nil
}

func (s *DeploymentDetailService) SaveDeploymentHistory(clusterID uint, namespace, deploymentName string, yamlContent, changeType, changeReason, changedBy string, replicas int32) error {
	var maxVersion int = 0
	global.GVA_DB.Model(&model.DeploymentHistory{}).
		Where("cluster_id = ? AND namespace = ? AND deployment_name = ?", clusterID, namespace, deploymentName).
		Select("COALESCE(MAX(version), 0)").
		Scan(&maxVersion)

	history := model.DeploymentHistory{
		ClusterID:      clusterID,
		Namespace:      namespace,
		DeploymentName: deploymentName,
		YAMLContent:    yamlContent,
		ChangeType:     changeType,
		ChangeReason:   changeReason,
		ChangedBy:      changedBy,
		Replicas:       replicas,
		Version:        maxVersion + 1,
	}

	if err := global.GVA_DB.Create(&history).Error; err != nil {
		global.GVA_LOG.Error("保存Deployment历史失败", zap.Error(err))
		return err
	}

	return nil
}

func (s *DeploymentDetailService) GetDeploymentHistories(clusterID uint, namespace, deploymentName string) ([]model.DeploymentHistory, error) {
	var histories []model.DeploymentHistory
	if err := global.GVA_DB.Where("cluster_id = ? AND namespace = ? AND deployment_name = ?", clusterID, namespace, deploymentName).
		Order("version desc").
		Find(&histories).Error; err != nil {
		return nil, err
	}
	return histories, nil
}

func (s *DeploymentDetailService) CompareDeploymentYAML(historyID1, historyID2 uint) (string, string, error) {
	var history1, history2 model.DeploymentHistory
	if err := global.GVA_DB.First(&history1, historyID1).Error; err != nil {
		return "", "", err
	}
	if err := global.GVA_DB.First(&history2, historyID2).Error; err != nil {
		return "", "", err
	}
	return history1.YAMLContent, history2.YAMLContent, nil
}

func (s *DeploymentDetailService) RollbackDeployment(clusterID uint, namespace, deploymentName string, historyID uint) error {
	var history model.DeploymentHistory
	if err := global.GVA_DB.First(&history, historyID).Error; err != nil {
		return err
	}

	cluster, err := clusterService.GetClusterByID(clusterID)
	if err != nil {
		return err
	}

	client, err := utils.GetClusterClientFromKubeconfig(cluster.Kubeconfig)
	if err != nil {
		return err
	}

	decoder := scheme.Codecs.UniversalDeserializer()
	obj, _, err := decoder.Decode([]byte(history.YAMLContent), nil, nil)
	if err != nil {
		return err
	}

	deploy, ok := obj.(*appsv1.Deployment)
	if !ok {
		return fmt.Errorf("invalid deployment YAML")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err = client.Clientset.AppsV1().Deployments(namespace).Update(ctx, deploy, metav1.UpdateOptions{})
	if err != nil {
		return err
	}

	return nil
}

func (s *DeploymentDetailService) ScaleDeploymentWithHistory(clusterID uint, namespace, deploymentName string, replicas int32, changedBy string) error {
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

	deploy, err := client.Clientset.AppsV1().Deployments(namespace).Get(ctx, deploymentName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	yamlBytes, err := yaml.Marshal(deploy)
	if err != nil {
		return err
	}

	if err := s.SaveDeploymentHistory(clusterID, namespace, deploymentName, string(yamlBytes), "scale", fmt.Sprintf("副本数从 %d 调整为 %d", *deploy.Spec.Replicas, replicas), changedBy, replicas); err != nil {
		return err
	}

	patch := []byte(`{"spec":{"replicas":` + strconv.Itoa(int(replicas)) + `}}`)
	_, err = client.Clientset.AppsV1().Deployments(namespace).Patch(
		ctx,
		deploymentName,
		types.MergePatchType,
		patch,
		metav1.PatchOptions{},
	)

	return err
}

func (s *DeploymentDetailService) UpdateDeploymentWithHistory(clusterID uint, namespace, deploymentName string, deploy *appsv1.Deployment, changeReason, changedBy string) error {
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

	yamlBytes, err := yaml.Marshal(deploy)
	if err != nil {
		return err
	}

	if err := s.SaveDeploymentHistory(clusterID, namespace, deploymentName, string(yamlBytes), "update", changeReason, changedBy, *deploy.Spec.Replicas); err != nil {
		return err
	}

	_, err = client.Clientset.AppsV1().Deployments(namespace).Update(ctx, deploy, metav1.UpdateOptions{})
	return err
}

func (s *DeploymentDetailService) GetDeploymentYAML(clusterID uint, namespace, name string) (string, error) {
	cluster, err := clusterService.GetClusterByID(clusterID)
	if err != nil {
		return "", err
	}

	client, err := utils.GetClusterClientFromKubeconfig(cluster.Kubeconfig)
	if err != nil {
		return "", err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	deploy, err := client.Clientset.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return "", err
	}

	deploy.TypeMeta = metav1.TypeMeta{
		APIVersion: "apps/v1",
		Kind:       "Deployment",
	}

	yamlBytes, err := yaml.Marshal(deploy)
	if err != nil {
		return "", err
	}

	return string(yamlBytes), nil
}

var deploymentDetailService = &DeploymentDetailService{}
