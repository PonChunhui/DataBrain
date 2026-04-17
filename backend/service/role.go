package service

import (
	"devops-backend/global"
	"devops-backend/model"
	"devops-backend/model/request"
)

type RoleService struct{}

func (s *RoleService) GetRoleList() ([]model.Role, error) {
	var roles []model.Role
	if err := global.GVA_DB.Find(&roles).Error; err != nil {
		return nil, err
	}
	return roles, nil
}

func (s *RoleService) CreateRole(req request.RoleRequest) error {
	role := model.Role{
		Name:        req.Name,
		Description: req.Description,
	}
	return global.GVA_DB.Create(&role).Error
}

func (s *RoleService) UpdateRole(id uint, req request.RoleRequest) error {
	var role model.Role
	if err := global.GVA_DB.First(&role, id).Error; err != nil {
		return err
	}

	role.Name = req.Name
	role.Description = req.Description

	return global.GVA_DB.Save(&role).Error
}

func (s *RoleService) DeleteRole(id uint) error {
	return global.GVA_DB.Delete(&model.Role{}, id).Error
}

func (s *RoleService) AssignMenuButton(roleID uint, menuButtonID uint) error {
	roleMenuButton := model.RoleMenuButton{
		RoleID:       roleID,
		MenuButtonID: menuButtonID,
	}
	return global.GVA_DB.Create(&roleMenuButton).Error
}

func (s *RoleService) AssignMenus(roleID uint, menuIDs []uint) error {
	global.GVA_DB.Where("role_id = ?", roleID).Delete(&model.RoleMenu{})

	for _, menuID := range menuIDs {
		roleMenu := model.RoleMenu{
			RoleID: roleID,
			MenuID: menuID,
		}
		if err := global.GVA_DB.Create(&roleMenu).Error; err != nil {
			return err
		}
	}
	return nil
}

func (s *RoleService) GetRoleMenus(roleID uint) ([]uint, error) {
	var roleMenus []model.RoleMenu
	if err := global.GVA_DB.Where("role_id = ?", roleID).Find(&roleMenus).Error; err != nil {
		return nil, err
	}

	menuIDs := make([]uint, len(roleMenus))
	for i, rm := range roleMenus {
		menuIDs[i] = rm.MenuID
	}
	return menuIDs, nil
}

func (s *RoleService) AssignApis(roleID uint, apiIDs []uint) error {
	global.GVA_DB.Where("role_id = ?", roleID).Delete(&model.RoleApi{})

	for _, apiID := range apiIDs {
		roleApi := model.RoleApi{
			RoleID: roleID,
			ApiID:  apiID,
		}
		if err := global.GVA_DB.Create(&roleApi).Error; err != nil {
			return err
		}
	}
	return nil
}

func (s *RoleService) GetRoleApis(roleID uint) ([]uint, error) {
	var roleApis []model.RoleApi
	if err := global.GVA_DB.Where("role_id = ?", roleID).Find(&roleApis).Error; err != nil {
		return nil, err
	}

	apiIDs := make([]uint, len(roleApis))
	for i, ra := range roleApis {
		apiIDs[i] = ra.ApiID
	}
	return apiIDs, nil
}
