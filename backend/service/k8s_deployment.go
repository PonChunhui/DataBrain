package service

import (
	"context"
	"strconv"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/yaml"

	"devops-backend/utils"
)

type DeploymentInfo struct {
	Name              string            `json:"name"`
	Namespace         string            `json:"namespace"`
	Status            string            `json:"status"`
	Replicas          int32             `json:"replicas"`
	ReadyReplicas     int32             `json:"ready_replicas"`
	AvailableReplicas int32             `json:"available_replicas"`
	Images            []string          `json:"images"`
	Labels            map[string]string `json:"labels"`
	CreatedAt         time.Time         `json:"created_at"`
}

type DeploymentService struct{}

func (s *DeploymentService) GetDeployments(clusterID uint, namespace string) ([]DeploymentInfo, error) {
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

	deployList, err := client.Clientset.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	deployments := make([]DeploymentInfo, 0, len(deployList.Items))
	for _, deploy := range deployList.Items {
		images := make([]string, 0)
		for _, container := range deploy.Spec.Template.Spec.Containers {
			images = append(images, container.Image)
		}

		readyReplicas := deploy.Status.ReadyReplicas

		replicas := int32(0)
		if deploy.Spec.Replicas != nil {
			replicas = *deploy.Spec.Replicas
		}

		status := getDeploymentStatus(&deploy, replicas, readyReplicas)

		deployments = append(deployments, DeploymentInfo{
			Name:              deploy.Name,
			Namespace:         deploy.Namespace,
			Status:            status,
			Replicas:          replicas,
			ReadyReplicas:     readyReplicas,
			AvailableReplicas: deploy.Status.AvailableReplicas,
			Images:            images,
			Labels:            deploy.Labels,
			CreatedAt:         deploy.CreationTimestamp.Time,
		})
	}

	return deployments, nil
}

func (s *DeploymentService) GetDeployment(clusterID uint, namespace, name string) (*appsv1.Deployment, error) {
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

	return deploy, nil
}

func (s *DeploymentService) ScaleDeployment(clusterID uint, namespace, name string, replicas int32) error {
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

	patch := []byte(`{"spec":{"replicas":` + strconv.Itoa(int(replicas)) + `}}`)
	_, err = client.Clientset.AppsV1().Deployments(namespace).Patch(
		ctx,
		name,
		types.MergePatchType,
		patch,
		metav1.PatchOptions{},
	)

	return err
}

func (s *DeploymentService) RestartDeployment(clusterID uint, namespace, name string) error {
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

	data := `{"spec":{"template":{"metadata":{"annotations":{"kubectl.kubernetes.io/restartedAt":"` + time.Now().Format("2006-01-02T15:04:05Z") + `"}}}}}`
	_, err = client.Clientset.AppsV1().Deployments(namespace).Patch(
		ctx,
		name,
		types.StrategicMergePatchType,
		[]byte(data),
		metav1.PatchOptions{},
	)

	return err
}

func (s *DeploymentService) DeleteDeployment(clusterID uint, namespace, name string) error {
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

	return client.Clientset.AppsV1().Deployments(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

func (s *DeploymentService) CreateDeployment(clusterID uint, namespace string, deploy *appsv1.Deployment) error {
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

	_, err = client.Clientset.AppsV1().Deployments(namespace).Create(ctx, deploy, metav1.CreateOptions{})
	return err
}

func (s *DeploymentService) GetDeploymentYAML(clusterID uint, namespace, name string) (string, error) {
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

func (s *DeploymentService) UpdateDeploymentYAML(clusterID uint, namespace, name, yamlStr string) error {
	cluster, err := clusterService.GetClusterByID(clusterID)
	if err != nil {
		return err
	}

	client, err := utils.GetClusterClientFromKubeconfig(cluster.Kubeconfig)
	if err != nil {
		return err
	}

	decoder := scheme.Codecs.UniversalDeserializer()
	obj, _, err := decoder.Decode([]byte(yamlStr), nil, nil)
	if err != nil {
		return err
	}

	deploy, ok := obj.(*appsv1.Deployment)
	if !ok {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err = client.Clientset.AppsV1().Deployments(namespace).Update(ctx, deploy, metav1.UpdateOptions{})
	return err
}

var clusterService = &ClusterService{}

func getDeploymentStatus(deploy *appsv1.Deployment, replicas, readyReplicas int32) string {
	if deploy.Spec.Paused {
		return "Paused"
	}

	if replicas == 0 {
		return "ScaledDown"
	}

	if readyReplicas == replicas && deploy.Status.AvailableReplicas == replicas {
		return "Running"
	}

	for _, cond := range deploy.Status.Conditions {
		if cond.Type == appsv1.DeploymentProgressing && cond.Status == "False" {
			return "Failed"
		}
		if cond.Type == appsv1.DeploymentReplicaFailure && cond.Status == "True" {
			return "Failed"
		}
	}

	if readyReplicas > 0 && readyReplicas < replicas {
		return "Updating"
	}

	if readyReplicas == 0 {
		return "Starting"
	}

	return "Degraded"
}
