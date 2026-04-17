package service

import (
	"context"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"devops-backend/utils"
)

type ConfigMapInfo struct {
	Name      string            `json:"name"`
	Namespace string            `json:"namespace"`
	DataKeys  []string          `json:"data_keys"`
	Data      map[string]string `json:"data,omitempty"`
	Labels    map[string]string `json:"labels"`
	CreatedAt time.Time         `json:"created_at"`
}

type K8sConfigMap struct{}

func (s *K8sConfigMap) GetConfigMaps(clusterID uint, namespace string) ([]ConfigMapInfo, error) {
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

	cmList, err := client.Clientset.CoreV1().ConfigMaps(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	configMaps := make([]ConfigMapInfo, 0, len(cmList.Items))
	for _, cm := range cmList.Items {
		dataKeys := make([]string, 0, len(cm.Data))
		for k := range cm.Data {
			dataKeys = append(dataKeys, k)
		}

		configMaps = append(configMaps, ConfigMapInfo{
			Name:      cm.Name,
			Namespace: cm.Namespace,
			DataKeys:  dataKeys,
			Labels:    cm.Labels,
			CreatedAt: cm.CreationTimestamp.Time,
		})
	}

	return configMaps, nil
}

func (s *K8sConfigMap) GetConfigMap(clusterID uint, namespace, name string) (*corev1.ConfigMap, error) {
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

	cm, err := client.Clientset.CoreV1().ConfigMaps(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return cm, nil
}

func (s *K8sConfigMap) CreateConfigMap(clusterID uint, namespace string, cm *corev1.ConfigMap) error {
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

	_, err = client.Clientset.CoreV1().ConfigMaps(namespace).Create(ctx, cm, metav1.CreateOptions{})
	return err
}

func (s *K8sConfigMap) UpdateConfigMap(clusterID uint, namespace string, cm *corev1.ConfigMap) error {
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

	_, err = client.Clientset.CoreV1().ConfigMaps(namespace).Update(ctx, cm, metav1.UpdateOptions{})
	return err
}

func (s *K8sConfigMap) DeleteConfigMap(clusterID uint, namespace, name string) error {
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

	return client.Clientset.CoreV1().ConfigMaps(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}
