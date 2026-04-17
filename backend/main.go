package main

import (
	"fmt"

	"devops-backend/core"
	"devops-backend/global"
	"devops-backend/initialize"
	"devops-backend/model"
	"devops-backend/router"
	"devops-backend/source"
	"go.uber.org/zap"
)

func main() {
	core.InitConfig()
	core.InitLog()

	global.GVA_LOG.Info("配置初始化成功", zap.Any("config", global.GVA_CONFIG))

	initialize.InitDB()

	global.GVA_LOG.Info("数据库初始化成功")

	global.GVA_DB.AutoMigrate(
		&model.User{},
		&model.Role{},
		&model.MenuButton{},
		&model.UserRole{},
		&model.RoleMenuButton{},
		&model.RoleMenu{},
		&model.RoleApi{},
		&model.Menu{},
		&model.Api{},
		&model.K8sCluster{},
		&model.K8sAuthorization{},
		&model.AuditLog{},
		&model.Webhook{},
		&model.DeploymentHistory{},
		&model.LLMConfig{},
		&model.DiagnosticRecord{},
	)

	global.GVA_LOG.Info("数据库迁移成功")

	source.InitData()

	Router := router.InitRouter()

	global.GVA_LOG.Info("路由初始化成功")

	s := fmt.Sprintf(":%d", global.GVA_CONFIG.Server.Port)
	global.GVA_LOG.Info("服务器启动", zap.String("port", s))

	if err := Router.Run(s); err != nil {
		global.GVA_LOG.Error("服务器启动失败", zap.Error(err))
	}
}
