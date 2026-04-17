package request

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type UserRequest struct {
	Username string `json:"username"`
	RealName string `json:"real_name"`
	Phone    string `json:"phone"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RoleRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

type MenuButtonRequest struct {
	MenuID      uint   `json:"menu_id" binding:"required"`
	Code        string `json:"code" binding:"required"`
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

type AssignRoleRequest struct {
	RoleID uint `json:"role_id" binding:"required"`
}

type AssignRolesRequest struct {
	RoleIDs []uint `json:"role_ids"`
}

type AssignMenuButtonRequest struct {
	MenuButtonID uint `json:"menu_button_id" binding:"required"`
}

type AssignMenusRequest struct {
	MenuIDs []uint `json:"menu_ids" binding:"required"`
}

type AssignApisRequest struct {
	ApiIDs []uint `json:"api_ids" binding:"required"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

type MenuRequest struct {
	Name      string `json:"name" binding:"required"`
	Path      string `json:"path" binding:"required"`
	Icon      string `json:"icon"`
	Sort      int    `json:"sort"`
	ParentID  uint   `json:"parent_id"`
	IsShow    bool   `json:"is_show"`
	Component string `json:"component"`
}

type ApiRequest struct {
	Path        string `json:"path" binding:"required"`
	Method      string `json:"method" binding:"required"`
	Description string `json:"description"`
	Group       string `json:"group"`
	Status      bool   `json:"status"`
}

type ApiSearchRequest struct {
	Path     string `form:"path"`
	Method   string `form:"method"`
	Group    string `form:"group"`
	Status   *bool  `form:"status"`
	Page     int    `form:"page"`
	PageSize int    `form:"pageSize"`
}

type AuditSearchRequest struct {
	Username  string `form:"username"`
	Action    string `form:"action"`
	Resource  string `form:"resource"`
	Status    string `form:"status"`
	StartTime string `form:"start_time"`
	EndTime   string `form:"end_time"`
	Page      int    `form:"page"`
	PageSize  int    `form:"pageSize"`
}

type WebhookRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	ExpiresDays int    `json:"expires_days"`
}

type WebhookRefreshRequest struct {
	ExpiresDays int `json:"expires_days"`
}

type K8sAuthListRequest struct {
	UserID    uint   `form:"user_id"`
	ClusterID uint   `form:"cluster_id"`
	Namespace string `form:"namespace"`
	Resource  string `form:"resource"`
	Page      int    `form:"page"`
	PageSize  int    `form:"pageSize"`
}

type K8sAuthRequest struct {
	UserID      uint   `json:"user_id" binding:"required"`
	ClusterID   uint   `json:"cluster_id"`
	Namespace   string `json:"namespace" binding:"required"`
	Resource    string `json:"resource" binding:"required"`
	CanView     bool   `json:"can_view"`
	CanEdit     bool   `json:"can_edit"`
	CanDelete   bool   `json:"can_delete"`
	CanCreate   bool   `json:"can_create"`
	Description string `json:"description"`
}
