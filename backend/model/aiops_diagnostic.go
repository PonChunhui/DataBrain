package model

import "time"

type DiagnosticRecord struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	ClusterID       uint      `gorm:"not null;index" json:"cluster_id"`
	ClusterName     string    `gorm:"size:100" json:"cluster_name"`
	Namespace       string    `gorm:"size:100;index" json:"namespace"`
	ResourceType    string    `gorm:"size:50;not null;index" json:"resource_type"`
	ResourceName    string    `gorm:"size:200;not null;index" json:"resource_name"`
	LLMConfigID     uint      `gorm:"not null;index" json:"llm_config_id"`
	LLMProvider     string    `gorm:"size:50" json:"llm_provider"`
	LLMModel        string    `gorm:"size:100" json:"llm_model"`
	InputData       string    `gorm:"type:text" json:"input_data"`
	ProblemDesc     string    `gorm:"type:text" json:"problem_desc"`
	RootCause       string    `gorm:"type:text" json:"root_cause"`
	PossibleCauses  string    `gorm:"type:text" json:"possible_causes"`
	Solution        string    `gorm:"type:text" json:"solution"`
	SolutionSteps   string    `gorm:"type:text" json:"solution_steps"`
	ImpactScope     string    `gorm:"type:text" json:"impact_scope"`
	RelatedRes      string    `gorm:"type:text" json:"related_resources"`
	Severity        string    `gorm:"size:20;index" json:"severity"`
	Confidence      float64   `json:"confidence"`
	RawResponse     string    `gorm:"type:text" json:"raw_response"`
	Status          string    `gorm:"size:20;default:'pending';index" json:"status"`
	ErrorMessage    string    `gorm:"type:text" json:"error_message"`
	Duration        int       `json:"duration"`
	TriggeredBy     uint      `json:"triggered_by"`
	TriggeredByName string    `gorm:"size:100" json:"triggered_by_name"`
	CreatedAt       time.Time `gorm:"autoCreateTime;index" json:"created_at"`
	UpdatedAt       time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (DiagnosticRecord) TableName() string {
	return "aiops_diagnostic_records"
}
