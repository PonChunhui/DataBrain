package service

import (
	"devops-backend/global"
	"devops-backend/model"
	"devops-backend/model/request"
)

type MenuService struct{}

func (s *MenuService) GetMenuList() ([]model.Menu, error) {
	var menus []model.Menu
	if err := global.GVA_DB.Order("sort asc, id asc").Find(&menus).Error; err != nil {
		return nil, err
	}
	return menus, nil
}

func (s *MenuService) GetMenuTree() ([]model.Menu, error) {
	var menus []model.Menu
	if err := global.GVA_DB.Where("parent_id = ?", 0).Order("sort asc, id asc").Find(&menus).Error; err != nil {
		return nil, err
	}

	for i := range menus {
		s.getChildren(&menus[i])
	}

	return menus, nil
}

func (s *MenuService) getChildren(menu *model.Menu) {
	var children []model.Menu
	global.GVA_DB.Where("parent_id = ?", menu.ID).Order("sort asc, id asc").Find(&children)
	if len(children) > 0 {
		for i := range children {
			s.getChildren(&children[i])
		}
		menu.Children = children
	}
	s.getButtons(menu)
}

func (s *MenuService) getButtons(menu *model.Menu) {
	var buttons []model.MenuButton
	global.GVA_DB.Where("menu_id = ?", menu.ID).Order("id asc").Find(&buttons)
	if len(buttons) > 0 {
		menu.Buttons = buttons
	}
}

func (s *MenuService) CreateMenu(req request.MenuRequest) error {
	menu := model.Menu{
		Name:      req.Name,
		Path:      req.Path,
		Icon:      req.Icon,
		Sort:      req.Sort,
		ParentID:  req.ParentID,
		IsShow:    req.IsShow,
		Component: req.Component,
	}
	return global.GVA_DB.Create(&menu).Error
}

func (s *MenuService) UpdateMenu(id uint, req request.MenuRequest) error {
	var menu model.Menu
	if err := global.GVA_DB.First(&menu, id).Error; err != nil {
		return err
	}

	menu.Name = req.Name
	menu.Path = req.Path
	menu.Icon = req.Icon
	menu.Sort = req.Sort
	menu.ParentID = req.ParentID
	menu.IsShow = req.IsShow
	menu.Component = req.Component

	return global.GVA_DB.Save(&menu).Error
}

func (s *MenuService) DeleteMenu(id uint) error {
	return global.GVA_DB.Delete(&model.Menu{}, id).Error
}
