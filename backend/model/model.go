package model

import (
	"gorm.io/gorm"
	"time"
)

type User struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`
	Username  string         `json:"username" gorm:"unique;not null"`
	RealName  string         `json:"real_name"`
	Phone     string         `json:"phone"`
	Email     string         `json:"email" gorm:"not null"`
	Password  string         `json:"-"`
	Roles     []Role         `json:"roles" gorm:"-"`
}

type Role struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at" gorm:"index"`
	Name        string         `json:"name" gorm:"not null"`
	Description string         `json:"description"`
}

type MenuButton struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at" gorm:"index"`
	MenuID      uint           `json:"menu_id" gorm:"index"`
	Code        string         `json:"code" gorm:"not null"`
	Name        string         `json:"name" gorm:"not null"`
	Description string         `json:"description"`
}

func (MenuButton) TableName() string {
	return "menu_buttons"
}

type UserRole struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`
	UserID    uint           `json:"user_id" gorm:"not null"`
	RoleID    uint           `json:"role_id" gorm:"not null"`
}

type RoleMenuButton struct {
	ID           uint           `json:"id" gorm:"primaryKey"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"deleted_at" gorm:"index"`
	RoleID       uint           `json:"role_id" gorm:"not null"`
	MenuButtonID uint           `json:"menu_button_id" gorm:"not null"`
}

func (RoleMenuButton) TableName() string {
	return "role_menu_buttons"
}

type RoleMenu struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`
	RoleID    uint           `json:"role_id" gorm:"not null;index"`
	MenuID    uint           `json:"menu_id" gorm:"not null;index"`
}

func (RoleMenu) TableName() string {
	return "role_menus"
}

type RoleApi struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`
	RoleID    uint           `json:"role_id" gorm:"not null;index"`
	ApiID     uint           `json:"api_id" gorm:"not null;index"`
}

func (RoleApi) TableName() string {
	return "role_apis"
}

type Menu struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`
	Name      string         `json:"name" gorm:"not null"`
	Path      string         `json:"path" gorm:"not null"`
	Icon      string         `json:"icon"`
	Sort      int            `json:"sort" gorm:"default:0"`
	ParentID  uint           `json:"parent_id" gorm:"default:0"`
	IsShow    bool           `json:"is_show" gorm:"default:true"`
	Component string         `json:"component"`
	Children  []Menu         `json:"children" gorm:"-"`
	Buttons   []MenuButton   `json:"buttons" gorm:"-"`
}

type Api struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at" gorm:"index"`
	Path        string         `json:"path" gorm:"not null"`
	Method      string         `json:"method" gorm:"not null"`
	Description string         `json:"description"`
	Group       string         `json:"group"`
	Status      bool           `json:"status" gorm:"default:true"`
}

type K8sCluster struct {
	ID                      uint           `json:"id" gorm:"primaryKey"`
	CreatedAt               time.Time      `json:"created_at"`
	UpdatedAt               time.Time      `json:"updated_at"`
	DeletedAt               gorm.DeletedAt `json:"deleted_at" gorm:"index"`
	Name                    string         `json:"name" gorm:"unique;not null"`
	Alias                   string         `json:"alias" gorm:"size:50"`
	Kubeconfig              string         `json:"kubeconfig" gorm:"type:text;not null"`
	Namespace               string         `json:"namespace" gorm:"default:'default'"`
	Description             string         `json:"description"`
	Status                  bool           `json:"status" gorm:"default:true"`
	PrometheusUrl           string         `json:"prometheus_url" gorm:"size:255"`
	PrometheusAuthEnabled   bool           `json:"prometheus_auth_enabled" gorm:"default:false"`
	PrometheusBasicAuthUser string         `json:"prometheus_basic_auth_user" gorm:"size:100"`
	PrometheusBasicAuthPass string         `json:"prometheus_basic_auth_pass" gorm:"size:100"`
}

func (K8sCluster) TableName() string {
	return "k8s_clusters"
}

type AuditLog struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	CreatedAt  time.Time `json:"created_at"`
	UserID     uint      `json:"user_id" gorm:"index"`
	Username   string    `json:"username"`
	Action     string    `json:"action" gorm:"index"`
	Resource   string    `json:"resource" gorm:"index"`
	ResourceID string    `json:"resource_id"`
	Detail     string    `json:"detail" gorm:"type:text"`
	IP         string    `json:"ip"`
	Status     string    `json:"status" gorm:"index"`
}

func (AuditLog) TableName() string {
	return "audit_logs"
}

type Webhook struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at" gorm:"index"`
	UserID      uint           `json:"user_id" gorm:"not null;index"`
	Token       string         `json:"token" gorm:"unique;not null"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	ExpiresDays int            `json:"expires_days" gorm:"default:30"`
	ExpiresAt   time.Time      `json:"expires_at"`
	IsExpired   bool           `json:"is_expired" gorm:"default:false"`
}

func (Webhook) TableName() string {
	return "webhooks"
}

type K8sAuthorization struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at" gorm:"index"`
	UserID      uint           `json:"user_id" gorm:"not null;index"`
	ClusterID   uint           `json:"cluster_id" gorm:"index"`        // 0表示所有集群
	Namespace   string         `json:"namespace" gorm:"size:64;index"` // "*"表示所有命名空间
	Resource    string         `json:"resource" gorm:"size:32;index"`  // deployment/pod/service/configmap/secret/ingress/node/"*"
	CanView     bool           `json:"can_view" gorm:"default:true"`
	CanEdit     bool           `json:"can_edit" gorm:"default:false"`
	CanDelete   bool           `json:"can_delete" gorm:"default:false"`
	CanCreate   bool           `json:"can_create" gorm:"default:false"`
	Description string         `json:"description"`
}

func (K8sAuthorization) TableName() string {
	return "k8s_authorizations"
}
