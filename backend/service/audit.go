package service

import (
	"devops-backend/global"
	"devops-backend/model"
	"devops-backend/model/request"
	"devops-backend/model/response"
	"encoding/json"
	"go.uber.org/zap"
)

type AuditService struct{}

func (s *AuditService) RecordAudit(userID uint, username, action, resource, resourceID, ip, status string, detail interface{}) error {
	var detailStr string
	if detail != nil {
		bytes, err := json.Marshal(detail)
		if err != nil {
			global.GVA_LOG.Error("审计日志详情序列化失败", zap.Error(err))
			detailStr = ""
		} else {
			detailStr = string(bytes)
		}
	}

	auditLog := model.AuditLog{
		UserID:     userID,
		Username:   username,
		Action:     action,
		Resource:   resource,
		ResourceID: resourceID,
		Detail:     detailStr,
		IP:         ip,
		Status:     status,
	}

	if err := global.GVA_DB.Create(&auditLog).Error; err != nil {
		global.GVA_LOG.Error("记录审计日志失败", zap.Error(err))
		return err
	}

	return nil
}

func (s *AuditService) GetAuditList(req request.AuditSearchRequest) (*response.PageResult, error) {
	var auditLogs []model.AuditLog
	var total int64

	db := global.GVA_DB.Model(&model.AuditLog{})

	if req.Username != "" {
		db = db.Where("username LIKE ?", "%"+req.Username+"%")
	}
	if req.Action != "" {
		db = db.Where("action = ?", req.Action)
	}
	if req.Resource != "" {
		db = db.Where("resource = ?", req.Resource)
	}
	if req.Status != "" {
		db = db.Where("status = ?", req.Status)
	}
	if req.StartTime != "" {
		db = db.Where("created_at >= ?", req.StartTime)
	}
	if req.EndTime != "" {
		db = db.Where("created_at <= ?", req.EndTime)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, err
	}

	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	offset := (req.Page - 1) * req.PageSize
	if err := db.Order("created_at DESC").Offset(offset).Limit(req.PageSize).Find(&auditLogs).Error; err != nil {
		return nil, err
	}

	return &response.PageResult{
		List:     auditLogs,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

func (s *AuditService) GetAuditByID(id uint) (*model.AuditLog, error) {
	var auditLog model.AuditLog
	if err := global.GVA_DB.First(&auditLog, id).Error; err != nil {
		return nil, err
	}
	return &auditLog, nil
}
