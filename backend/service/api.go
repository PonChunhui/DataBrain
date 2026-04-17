package service

import (
	"devops-backend/global"
	"devops-backend/model"
	"devops-backend/model/request"
)

type ApiService struct{}

func (s *ApiService) GetApiList(req request.ApiSearchRequest) ([]model.Api, int64, error) {
	var apis []model.Api
	var total int64

	db := global.GVA_DB.Model(&model.Api{})

	if req.Path != "" {
		db = db.Where("path LIKE ?", "%"+req.Path+"%")
	}
	if req.Method != "" {
		db = db.Where("method = ?", req.Method)
	}
	if req.Group != "" {
		db = db.Where("group LIKE ?", "%"+req.Group+"%")
	}
	if req.Status != nil {
		db = db.Where("status = ?", *req.Status)
	}

	db.Count(&total)

	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	offset := (req.Page - 1) * req.PageSize
	if err := db.Order("id asc").Offset(offset).Limit(req.PageSize).Find(&apis).Error; err != nil {
		return nil, 0, err
	}

	return apis, total, nil
}

func (s *ApiService) CreateApi(req request.ApiRequest) error {
	api := model.Api{
		Path:        req.Path,
		Method:      req.Method,
		Description: req.Description,
		Group:       req.Group,
		Status:      req.Status,
	}
	return global.GVA_DB.Create(&api).Error
}

func (s *ApiService) UpdateApi(id uint, req request.ApiRequest) error {
	var api model.Api
	if err := global.GVA_DB.First(&api, id).Error; err != nil {
		return err
	}

	api.Path = req.Path
	api.Method = req.Method
	api.Description = req.Description
	api.Group = req.Group
	api.Status = req.Status

	return global.GVA_DB.Save(&api).Error
}

func (s *ApiService) DeleteApi(id uint) error {
	return global.GVA_DB.Delete(&model.Api{}, id).Error
}
