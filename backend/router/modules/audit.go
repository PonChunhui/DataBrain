package modules

import (
	v1 "devops-backend/api/v1"
	"github.com/gin-gonic/gin"
)

func InitAuditRouter(ApiGroup *gin.RouterGroup) {
	router := ApiGroup.Group("audit")
	api := &v1.AuditApi{}
	{
		router.GET("", api.GetAuditList)
		router.GET("/:id", api.GetAuditByID)
	}
}
