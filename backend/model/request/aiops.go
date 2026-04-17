package request

type LLMConfigCreate struct {
	Name        string  `json:"name" binding:"required"`
	Provider    string  `json:"provider" binding:"required"`
	APIKey      string  `json:"api_key" binding:"required"`
	BaseURL     string  `json:"base_url"`
	Model       string  `json:"model" binding:"required"`
	MaxTokens   int     `json:"max_tokens"`
	Temperature float64 `json:"temperature"`
	IsDefault   bool    `json:"is_default"`
	IsEnabled   bool    `json:"is_enabled"`
}

type LLMConfigUpdate struct {
	Name        string  `json:"name"`
	APIKey      string  `json:"api_key"`
	BaseURL     string  `json:"base_url"`
	Model       string  `json:"model"`
	MaxTokens   int     `json:"max_tokens"`
	Temperature float64 `json:"temperature"`
	IsDefault   bool    `json:"is_default"`
	IsEnabled   bool    `json:"is_enabled"`
}

type DiagnosticRequest struct {
	ClusterID    uint   `json:"cluster_id" binding:"required"`
	Namespace    string `json:"namespace" binding:"required"`
	ResourceType string `json:"resource_type" binding:"required"`
	ResourceName string `json:"resource_name" binding:"required"`
	LLMConfigID  uint   `json:"llm_config_id"`
}

type DiagnosticHistoryQuery struct {
	Page         int    `form:"page"`
	PageSize     int    `form:"pageSize"`
	ResourceType string `form:"resource_type"`
	Severity     string `form:"severity"`
	LLMProvider  string `form:"llm_provider"`
	ClusterID    uint   `form:"cluster_id"`
	Namespace    string `form:"namespace"`
	StartTime    string `form:"start_time"`
	EndTime      string `form:"end_time"`
	Keyword      string `form:"keyword"`
}

type ChatRequest struct {
	RecordID    uint   `json:"record_id" binding:"required"`
	Message     string `json:"message" binding:"required"`
	LLMConfigID uint   `json:"llm_config_id"`
}

type ClusterInspectionRequest struct {
	ClusterID   uint `json:"cluster_id" binding:"required"`
	LLMConfigID uint `json:"llm_config_id"`
}
