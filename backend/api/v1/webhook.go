package v1

import (
	"net/http"
	"strconv"

	"devops-backend/global"
	"devops-backend/middleware"
	"devops-backend/model/request"
	"devops-backend/model/response"
	"devops-backend/service"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type WebhookApi struct{}

var webhookService = &service.WebhookService{}

func (api *WebhookApi) CreateWebhook(c *gin.Context) {
	claims, exists := c.Get("claims")
	if !exists {
		c.JSON(http.StatusOK, response.FailWithMsg(nil, "未登录"))
		return
	}

	userID := uint(0)
	if customClaims, ok := claims.(*middleware.CustomClaims); ok {
		userID = customClaims.UserID
	}

	var req request.WebhookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, response.FailWithMsg(nil, "参数错误"))
		return
	}

	webhook, err := webhookService.CreateWebhook(userID, req.Name, req.Description, req.ExpiresDays)
	if err != nil {
		global.GVA_LOG.Error("创建Webhook失败", zap.Error(err))
		c.JSON(http.StatusOK, response.FailWithMsg(nil, "创建失败"))
		return
	}

	c.JSON(http.StatusOK, response.Success(webhook))
}

func (api *WebhookApi) GetUserWebhooks(c *gin.Context) {
	claims, exists := c.Get("claims")
	if !exists {
		c.JSON(http.StatusOK, response.FailWithMsg(nil, "未登录"))
		return
	}

	userID := uint(0)
	if customClaims, ok := claims.(*middleware.CustomClaims); ok {
		userID = customClaims.UserID
	}

	webhooks, err := webhookService.GetUserWebhooks(userID)
	if err != nil {
		global.GVA_LOG.Error("获取Webhook列表失败", zap.Error(err))
		c.JSON(http.StatusOK, response.FailWithMsg(nil, "获取失败"))
		return
	}

	c.JSON(http.StatusOK, response.Success(webhooks))
}

func (api *WebhookApi) DeleteWebhook(c *gin.Context) {
	claims, exists := c.Get("claims")
	if !exists {
		c.JSON(http.StatusOK, response.FailWithMsg(nil, "未登录"))
		return
	}

	userID := uint(0)
	if customClaims, ok := claims.(*middleware.CustomClaims); ok {
		userID = customClaims.UserID
	}

	id, _ := strconv.Atoi(c.Param("id"))
	if err := webhookService.DeleteWebhook(userID, uint(id)); err != nil {
		c.JSON(http.StatusOK, response.FailWithMsg(nil, err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(nil))
}

func (api *WebhookApi) RefreshWebhook(c *gin.Context) {
	claims, exists := c.Get("claims")
	if !exists {
		c.JSON(http.StatusOK, response.FailWithMsg(nil, "未登录"))
		return
	}

	userID := uint(0)
	if customClaims, ok := claims.(*middleware.CustomClaims); ok {
		userID = customClaims.UserID
	}

	id, _ := strconv.Atoi(c.Param("id"))

	var req request.WebhookRefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		req.ExpiresDays = 30
	}

	webhook, err := webhookService.RefreshWebhook(userID, uint(id), req.ExpiresDays)
	if err != nil {
		c.JSON(http.StatusOK, response.FailWithMsg(nil, err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(webhook))
}
