package model

import (
	"time"

	"gorm.io/gorm"
)

type DeploymentHistory struct {
	ID             uint           `json:"id" gorm:"primaryKey"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `json:"deleted_at" gorm:"index"`
	ClusterID      uint           `json:"cluster_id" gorm:"not null;index"`
	Namespace      string         `json:"namespace" gorm:"not null;index"`
	DeploymentName string         `json:"deployment_name" gorm:"not null;index"`
	YAMLContent    string         `json:"yaml_content" gorm:"type:text;not null"`
	ChangeType     string         `json:"change_type" gorm:"not null"`
	ChangeReason   string         `json:"change_reason"`
	ChangedBy      string         `json:"changed_by"`
	Replicas       int32          `json:"replicas"`
	Version        int            `json:"version" gorm:"not null;index"`
}

func (DeploymentHistory) TableName() string {
	return "deployment_histories"
}
