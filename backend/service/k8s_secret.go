package service

import (
	"context"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"devops-backend/utils"
)

type SecretInfo struct {
	Name      string            `json:"name"`
	Namespace string            `json:"namespace"`
	Type      string            `json:"type"`
	DataKeys  []string          `json:"data_keys"`
	Labels    map[string]string `json:"labels"`
	CreatedAt time.Time         `json:"created_at"`
}

type K8sSecret struct{}

func (s *K8sSecret) GetSecrets(clusterID uint, namespace string) ([]SecretInfo, error) {
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

	secretList, err := client.Clientset.CoreV1().Secrets(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	secrets := make([]SecretInfo, 0, len(secretList.Items))
	for _, secret := range secretList.Items {
		dataKeys := make([]string, 0, len(secret.Data))
		for k := range secret.Data {
			dataKeys = append(dataKeys, k)
		}

		secrets = append(secrets, SecretInfo{
			Name:      secret.Name,
			Namespace: secret.Namespace,
			Type:      string(secret.Type),
			DataKeys:  dataKeys,
			Labels:    secret.Labels,
			CreatedAt: secret.CreationTimestamp.Time,
		})
	}

	return secrets, nil
}

func (s *K8sSecret) GetSecret(clusterID uint, namespace, name string) (*corev1.Secret, error) {
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

	secret, err := client.Clientset.CoreV1().Secrets(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return secret, nil
}

func (s *K8sSecret) CreateSecret(clusterID uint, namespace string, secret *corev1.Secret) error {
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

	_, err = client.Clientset.CoreV1().Secrets(namespace).Create(ctx, secret, metav1.CreateOptions{})
	return err
}

func (s *K8sSecret) UpdateSecret(clusterID uint, namespace string, secret *corev1.Secret) error {
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

	_, err = client.Clientset.CoreV1().Secrets(namespace).Update(ctx, secret, metav1.UpdateOptions{})
	return err
}

func (s *K8sSecret) DeleteSecret(clusterID uint, namespace, name string) error {
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

	return client.Clientset.CoreV1().Secrets(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}
