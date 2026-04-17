package service

import (
	"context"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"devops-backend/global"
	"devops-backend/model"
	"devops-backend/utils"
)

type ClusterService struct{}

func (s *ClusterService) GetClusters() ([]model.K8sCluster, error) {
	var clusters []model.K8sCluster
	err := global.GVA_DB.Find(&clusters).Error

	for i := range clusters {
		clusters[i].Kubeconfig = ""
		clusters[i].PrometheusBasicAuthPass = ""
	}

	return clusters, err
}

func (s *ClusterService) GetClusterByID(id uint) (*model.K8sCluster, error) {
	var cluster model.K8sCluster
	err := global.GVA_DB.First(&cluster, id).Error
	if err != nil {
		return nil, err
	}
	return &cluster, nil
}

func (s *ClusterService) GetClusterByAlias(alias string) (*model.K8sCluster, error) {
	var cluster model.K8sCluster
	err := global.GVA_DB.Where("alias = ?", alias).First(&cluster).Error
	if err != nil {
		return nil, err
	}
	cluster.Kubeconfig = ""
	cluster.PrometheusBasicAuthPass = ""
	return &cluster, nil
}

func (s *ClusterService) GetClusterByIDSecure(id uint) (*model.K8sCluster, error) {
	cluster, err := s.GetClusterByID(id)
	if err != nil {
		return nil, err
	}

	return &model.K8sCluster{
		ID:                      cluster.ID,
		Name:                    cluster.Name,
		Alias:                   cluster.Alias,
		Kubeconfig:              "",
		Namespace:               cluster.Namespace,
		Description:             cluster.Description,
		Status:                  cluster.Status,
		PrometheusUrl:           cluster.PrometheusUrl,
		PrometheusAuthEnabled:   cluster.PrometheusAuthEnabled,
		PrometheusBasicAuthUser: cluster.PrometheusBasicAuthUser,
		PrometheusBasicAuthPass: "",
		CreatedAt:               cluster.CreatedAt,
		UpdatedAt:               cluster.UpdatedAt,
	}, nil
}

func (s *ClusterService) CreateCluster(cluster *model.K8sCluster) error {
	if err := utils.TestClusterConnectionFromKubeconfig(cluster.Kubeconfig); err != nil {
		return err
	}

	return global.GVA_DB.Create(cluster).Error
}

func (s *ClusterService) UpdateCluster(cluster *model.K8sCluster) error {
	if cluster.Kubeconfig != "" {
		if err := utils.TestClusterConnectionFromKubeconfig(cluster.Kubeconfig); err != nil {
			return err
		}
	}

	utils.RemoveClusterClient(cluster.Kubeconfig)

	return global.GVA_DB.Save(cluster).Error
}

func (s *ClusterService) DeleteCluster(id uint) error {
	var cluster model.K8sCluster
	if err := global.GVA_DB.First(&cluster, id).Error; err != nil {
		return err
	}

	utils.RemoveClusterClient(cluster.Kubeconfig)

	return global.GVA_DB.Delete(&cluster).Error
}

func (s *ClusterService) GetNamespaces(clusterID uint) ([]string, error) {
	cluster, err := s.GetClusterByID(clusterID)
	if err != nil {
		return nil, err
	}

	client, err := utils.GetClusterClientFromKubeconfig(cluster.Kubeconfig)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	nsList, err := client.Clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	namespaces := make([]string, 0, len(nsList.Items))
	for _, ns := range nsList.Items {
		namespaces = append(namespaces, ns.Name)
	}

	return namespaces, nil
}

func (s *ClusterService) GetClusterInfo(clusterID uint) (map[string]string, error) {
	cluster, err := s.GetClusterByID(clusterID)
	if err != nil {
		return nil, err
	}

	info := make(map[string]string)

	server, err := utils.GetClusterServer(cluster.Kubeconfig)
	if err == nil {
		info["server"] = server
	}

	context, err := utils.GetCurrentContext(cluster.Kubeconfig)
	if err == nil {
		info["context"] = context
	}

	info["name"] = cluster.Name
	info["namespace"] = cluster.Namespace

	return info, nil
}
