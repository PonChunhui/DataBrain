package modules

import (
	v1 "devops-backend/api/v1"
	"github.com/gin-gonic/gin"
)

func InitWebhookRouter(ApiGroup *gin.RouterGroup) {
	router := ApiGroup.Group("webhook")
	api := &v1.WebhookApi{}
	{
		router.GET("", api.GetUserWebhooks)
		router.POST("", api.CreateWebhook)
		router.DELETE("/:id", api.DeleteWebhook)
		router.PUT("/:id/refresh", api.RefreshWebhook)
	}
}
