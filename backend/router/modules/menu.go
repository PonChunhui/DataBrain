package modules

import (
	v1 "devops-backend/api/v1"
	"github.com/gin-gonic/gin"
)

func InitMenuRouter(ApiGroup *gin.RouterGroup) {
	router := ApiGroup.Group("menu")
	api := &v1.MenuApi{}
	{
		router.GET("", api.GetMenuList)
		router.GET("/tree", api.GetMenuTree)
		router.POST("", api.CreateMenu)
		router.PUT("/:id", api.UpdateMenu)
		router.DELETE("/:id", api.DeleteMenu)
	}

	buttonRouter := ApiGroup.Group("menu-button")
	buttonApi := &v1.MenuButtonApi{}
	{
		buttonRouter.GET("", buttonApi.GetMenuButtonList)
		buttonRouter.GET("/menu/:menu_id", buttonApi.GetButtonsByMenu)
		buttonRouter.POST("", buttonApi.CreateMenuButton)
		buttonRouter.PUT("/:id", buttonApi.UpdateMenuButton)
		buttonRouter.DELETE("/:id", buttonApi.DeleteMenuButton)
	}
}
