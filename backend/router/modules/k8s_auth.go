package modules

import (
	v1 "devops-backend/api/v1"
	"github.com/gin-gonic/gin"
)

func InitK8sAuthRouter(ApiGroup *gin.RouterGroup) {
	k8sAuthRouter := ApiGroup.Group("k8s/auth")
	k8sAuthController := &v1.K8sAuthController{}
	{
		k8sAuthRouter.GET("", k8sAuthController.GetAuthorizations)
		k8sAuthRouter.GET("/:id", k8sAuthController.GetAuthorization)
		k8sAuthRouter.POST("", k8sAuthController.CreateAuthorization)
		k8sAuthRouter.PUT("/:id", k8sAuthController.UpdateAuthorization)
		k8sAuthRouter.DELETE("/:id", k8sAuthController.DeleteAuthorization)
		k8sAuthRouter.GET("/user", k8sAuthController.GetUserAuthorizations)
	}

	userK8sRouter := ApiGroup.Group("k8s/user")
	{
		userK8sRouter.GET("/clusters", k8sAuthController.GetUserAuthorizedClusters)
		userK8sRouter.GET("/namespaces", k8sAuthController.GetUserAuthorizedNamespaces)
		userK8sRouter.GET("/permissions", k8sAuthController.GetUserPermissions)
	}
}
