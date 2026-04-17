package router

import (
	"devops-backend/middleware"
	"devops-backend/router/modules"
	"github.com/gin-gonic/gin"
)

func InitRouter() *gin.Engine {
	Router := gin.Default()

	Router.Use(middleware.Cors())

	ApiGroup := Router.Group("/api")

	// 公开路由（无需认证）
	modules.InitUserRouterPublic(ApiGroup)

	// 需认证路由
	ApiGroup.Use(middleware.JWTAuth())
	ApiGroup.Use(middleware.AuditMiddleware())
	{
		modules.InitUserRouter(ApiGroup)
		modules.InitRoleRouter(ApiGroup)
		modules.InitMenuRouter(ApiGroup)
		modules.InitApiRouter(ApiGroup)
		modules.InitK8sRouter(ApiGroup)
		modules.InitK8sAuthRouter(ApiGroup)
		modules.InitAuditRouter(ApiGroup)
		modules.InitWebhookRouter(ApiGroup)
		modules.InitAIOPSRouter(ApiGroup)
	}

	return Router
}
