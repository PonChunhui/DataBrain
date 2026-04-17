package service

import (
	"devops-backend/global"
	"devops-backend/model"
	"devops-backend/model/request"
)

type MenuButtonService struct{}

func (s *MenuButtonService) GetMenuButtonList() ([]model.MenuButton, error) {
	var buttons []model.MenuButton
	if err := global.GVA_DB.Order("menu_id, id asc").Find(&buttons).Error; err != nil {
		return nil, err
	}
	return buttons, nil
}

func (s *MenuButtonService) GetMenuButtons(menuID uint) ([]model.MenuButton, error) {
	var buttons []model.MenuButton
	if err := global.GVA_DB.Where("menu_id = ?", menuID).Order("id asc").Find(&buttons).Error; err != nil {
		return nil, err
	}
	return buttons, nil
}

func (s *MenuButtonService) CreateMenuButton(req request.MenuButtonRequest) error {
	button := model.MenuButton{
		MenuID:      req.MenuID,
		Code:        req.Code,
		Name:        req.Name,
		Description: req.Description,
	}
	return global.GVA_DB.Create(&button).Error
}

func (s *MenuButtonService) UpdateMenuButton(id uint, req request.MenuButtonRequest) error {
	var button model.MenuButton
	if err := global.GVA_DB.First(&button, id).Error; err != nil {
		return err
	}

	button.MenuID = req.MenuID
	button.Code = req.Code
	button.Name = req.Name
	button.Description = req.Description

	return global.GVA_DB.Save(&button).Error
}

func (s *MenuButtonService) DeleteMenuButton(id uint) error {
	return global.GVA_DB.Delete(&model.MenuButton{}, id).Error
}
