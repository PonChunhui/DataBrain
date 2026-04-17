package modules

import (
	v1 "devops-backend/api/v1"
	"github.com/gin-gonic/gin"
)

func InitUserRouterPublic(ApiGroup *gin.RouterGroup) {
	router := ApiGroup.Group("user")
	api := &v1.UserApi{}
	{
		router.POST("login", api.Login)
	}
}

func InitUserRouter(ApiGroup *gin.RouterGroup) {
	router := ApiGroup.Group("user")
	api := &v1.UserApi{}
	{
		router.GET("", api.GetUserList)
		router.POST("", api.CreateUser)
		router.GET("/:id", api.GetUserByID)
		router.PUT("/:id", api.UpdateUser)
		router.DELETE("/:id", api.DeleteUser)
		router.POST("/:id/roles", api.AssignRoles)
		router.GET("/:id/roles", api.GetUserRoles)
		router.GET("/:id/menus", api.GetUserMenus)
		router.GET("/:id/apis", api.GetUserApis)
		router.PUT("/:id/password", api.ChangePassword)
	}
}
