package global

import (
	"devops-backend/config"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var (
	GVA_DB     *gorm.DB
	GVA_CONFIG config.Config
	GVA_LOG    *zap.Logger
)
