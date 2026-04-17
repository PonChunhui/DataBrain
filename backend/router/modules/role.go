package modules

import (
	v1 "devops-backend/api/v1"
	"github.com/gin-gonic/gin"
)

func InitRoleRouter(ApiGroup *gin.RouterGroup) {
	router := ApiGroup.Group("role")
	api := &v1.RoleApi{}
	{
		router.GET("", api.GetRoleList)
		router.POST("", api.CreateRole)
		router.PUT("/:id", api.UpdateRole)
		router.DELETE("/:id", api.DeleteRole)
		router.POST("/:id/button", api.AssignMenuButton)
		router.POST("/:id/menus", api.AssignMenus)
		router.GET("/:id/menus", api.GetRoleMenus)
		router.POST("/:id/apis", api.AssignApis)
		router.GET("/:id/apis", api.GetRoleApis)
	}
}
