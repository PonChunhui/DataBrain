package service

import (
	"devops-backend/global"
	"devops-backend/model"
	"devops-backend/model/request"
)

type K8sAuthService struct{}

func (s *K8sAuthService) GetAuthorizations(params request.K8sAuthListRequest) ([]model.K8sAuthorization, int64, error) {
	var auths []model.K8sAuthorization
	var total int64

	db := global.GVA_DB.Model(&model.K8sAuthorization{})

	if params.UserID > 0 {
		db = db.Where("user_id = ?", params.UserID)
	}
	if params.ClusterID > 0 {
		db = db.Where("cluster_id = ?", params.ClusterID)
	}
	if params.Resource != "" {
		db = db.Where("resource = ?", params.Resource)
	}
	if params.Namespace != "" {
		db = db.Where("namespace = ?", params.Namespace)
	}

	err := db.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	offset := (params.Page - 1) * params.PageSize
	err = db.Offset(offset).Limit(params.PageSize).Order("id DESC").Find(&auths).Error
	if err != nil {
		return nil, 0, err
	}

	return auths, total, nil
}

func (s *K8sAuthService) GetAuthorizationByID(id uint) (*model.K8sAuthorization, error) {
	var auth model.K8sAuthorization
	err := global.GVA_DB.First(&auth, id).Error
	if err != nil {
		return nil, err
	}
	return &auth, nil
}

func (s *K8sAuthService) CreateAuthorization(auth *model.K8sAuthorization) error {
	return global.GVA_DB.Create(auth).Error
}

func (s *K8sAuthService) UpdateAuthorization(auth *model.K8sAuthorization) error {
	return global.GVA_DB.Save(auth).Error
}

func (s *K8sAuthService) DeleteAuthorization(id uint) error {
	return global.GVA_DB.Delete(&model.K8sAuthorization{}, id).Error
}

func (s *K8sAuthService) GetUserAuthorizations(userID uint) ([]model.K8sAuthorization, error) {
	var auths []model.K8sAuthorization
	err := global.GVA_DB.Where("user_id = ?", userID).Find(&auths).Error
	return auths, err
}

func (s *K8sAuthService) CheckPermission(userID, clusterID uint, namespace, resource, action string) bool {
	var auths []model.K8sAuthorization
	global.GVA_DB.Where("user_id = ?", userID).Find(&auths)

	for _, auth := range auths {
		if auth.ClusterID != 0 && auth.ClusterID != clusterID {
			continue
		}
		if auth.Namespace != "*" && auth.Namespace != namespace {
			continue
		}
		if auth.Resource != "*" && auth.Resource != resource {
			continue
		}

		switch action {
		case "view":
			return auth.CanView
		case "edit", "update":
			return auth.CanEdit
		case "delete":
			return auth.CanDelete
		case "create":
			return auth.CanCreate
		}
	}
	return false
}

func (s *K8sAuthService) GetUserAuthorizedClusters(userID uint) ([]uint, error) {
	var clusterIDs []uint
	err := global.GVA_DB.Model(&model.K8sAuthorization{}).
		Where("user_id = ? AND cluster_id != 0", userID).
		Distinct("cluster_id").
		Pluck("cluster_id", &clusterIDs).Error

	var allCluster bool
	global.GVA_DB.Model(&model.K8sAuthorization{}).
		Where("user_id = ? AND cluster_id = 0", userID).
		Select("1").Limit(1).Scan(&allCluster)
	if allCluster {
		var clusters []model.K8sCluster
		global.GVA_DB.Model(&model.K8sCluster{}).Find(&clusters)
		clusterIDs = make([]uint, len(clusters))
		for i, c := range clusters {
			clusterIDs[i] = c.ID
		}
	}

	return clusterIDs, err
}

func (s *K8sAuthService) GetUserAuthorizedNamespaces(userID, clusterID uint) ([]string, error) {
	var namespaces []string
	err := global.GVA_DB.Model(&model.K8sAuthorization{}).
		Where("user_id = ? AND (cluster_id = ? OR cluster_id = 0)", userID, clusterID).
		Distinct("namespace").
		Pluck("namespace", &namespaces).Error

	return namespaces, err
}

func (s *K8sAuthService) GetUserPermissions(userID uint) ([]model.K8sAuthorization, error) {
	return s.GetUserAuthorizations(userID)
}

func (s *K8sAuthService) BatchCreateAuthorizations(auths []model.K8sAuthorization) error {
	return global.GVA_DB.Create(&auths).Error
}

func (s *K8sAuthService) GetUserAuthCount(userID uint) (int64, error) {
	var count int64
	err := global.GVA_DB.Model(&model.K8sAuthorization{}).Where("user_id = ?", userID).Count(&count).Error
	return count, err
}

func (s *K8sAuthService) IsSuperAdmin(userID uint) bool {
	var userRoles []model.UserRole
	global.GVA_DB.Where("user_id = ?", userID).Find(&userRoles)

	for _, ur := range userRoles {
		var role model.Role
		if err := global.GVA_DB.First(&role, ur.RoleID).Error; err == nil {
			if role.Name == "admin" || role.Name == "super_admin" {
				return true
			}
		}
	}
	return false
}
