package service

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"devops-backend/global"
	"devops-backend/model"
	"go.uber.org/zap"
)

type WebhookService struct{}

func (s *WebhookService) generateToken() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func (s *WebhookService) CreateWebhook(userID uint, name, description string, expiresDays int) (*model.Webhook, error) {
	token := s.generateToken()

	var expiresAt time.Time
	if expiresDays > 0 {
		expiresAt = time.Now().AddDate(0, 0, expiresDays)
	} else {
		expiresAt = time.Now().AddDate(0, 0, 30)
	}

	webhook := model.Webhook{
		UserID:      userID,
		Token:       token,
		Name:        name,
		Description: description,
		ExpiresDays: expiresDays,
		ExpiresAt:   expiresAt,
		IsExpired:   false,
	}

	if err := global.GVA_DB.Create(&webhook).Error; err != nil {
		global.GVA_LOG.Error("创建Webhook失败", zap.Error(err))
		return nil, err
	}

	return &webhook, nil
}

func (s *WebhookService) GetUserWebhooks(userID uint) ([]model.Webhook, error) {
	var webhooks []model.Webhook
	if err := global.GVA_DB.Where("user_id = ?", userID).Order("created_at desc").Find(&webhooks).Error; err != nil {
		return nil, err
	}

	for i := range webhooks {
		if webhooks[i].ExpiresAt.Before(time.Now()) && !webhooks[i].IsExpired {
			webhooks[i].IsExpired = true
			global.GVA_DB.Save(&webhooks[i])
		}
	}

	return webhooks, nil
}

func (s *WebhookService) DeleteWebhook(userID uint, webhookID uint) error {
	result := global.GVA_DB.Where("id = ? AND user_id = ?", webhookID, userID).Delete(&model.Webhook{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("Webhook不存在或无权限删除")
	}
	return nil
}

func (s *WebhookService) RefreshWebhook(userID uint, webhookID uint, expiresDays int) (*model.Webhook, error) {
	var webhook model.Webhook
	if err := global.GVA_DB.Where("id = ? AND user_id = ?", webhookID, userID).First(&webhook).Error; err != nil {
		return nil, errors.New("Webhook不存在或无权限")
	}

	webhook.Token = s.generateToken()
	if expiresDays > 0 {
		webhook.ExpiresDays = expiresDays
	}
	webhook.ExpiresAt = time.Now().AddDate(0, 0, webhook.ExpiresDays)
	webhook.IsExpired = false

	if err := global.GVA_DB.Save(&webhook).Error; err != nil {
		return nil, err
	}

	return &webhook, nil
}

func (s *WebhookService) ValidateWebhookToken(token string) (uint, error) {
	var webhook model.Webhook
	if err := global.GVA_DB.Where("token = ? AND is_expired = ?", token, false).First(&webhook).Error; err != nil {
		return 0, errors.New("无效的Webhook Token")
	}

	if webhook.ExpiresAt.Before(time.Now()) {
		webhook.IsExpired = true
		global.GVA_DB.Save(&webhook)
		return 0, errors.New("Webhook Token已过期")
	}

	return webhook.UserID, nil
}
