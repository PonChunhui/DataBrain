package v1

import (
	"errors"
	"strconv"

	"devops-backend/model"
)

func resolveClusterID(clusterAliasOrID string) (uint, error) {
	if clusterAliasOrID == "" {
		return 0, errors.New("缺少集群参数")
	}

	id, err := strconv.Atoi(clusterAliasOrID)
	if err == nil && id > 0 {
		return uint(id), nil
	}

	cluster, err := clusterService.GetClusterByAlias(clusterAliasOrID)
	if err != nil {
		return 0, errors.New("集群不存在: " + clusterAliasOrID)
	}

	return cluster.ID, nil
}

func resolveCluster(clusterAliasOrID string) (*model.K8sCluster, error) {
	if clusterAliasOrID == "" {
		return nil, errors.New("缺少集群参数")
	}

	id, err := strconv.Atoi(clusterAliasOrID)
	if err == nil && id > 0 {
		return clusterService.GetClusterByID(uint(id))
	}

	return clusterService.GetClusterByAlias(clusterAliasOrID)
}
