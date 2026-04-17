package modules

import (
	v1 "devops-backend/api/v1"
	"github.com/gin-gonic/gin"
)

func InitApiRouter(ApiGroup *gin.RouterGroup) {
	router := ApiGroup.Group("api")
	api := &v1.ApiApi{}
	{
		router.GET("", api.GetApiList)
		router.POST("", api.CreateApi)
		router.PUT("/:id", api.UpdateApi)
		router.DELETE("/:id", api.DeleteApi)
	}
}
