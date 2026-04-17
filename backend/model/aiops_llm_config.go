package model

import "time"

type LLMConfig struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"size:100;not null;unique" json:"name"`
	Provider    string    `gorm:"size:50;not null" json:"provider"`
	APIKey      string    `gorm:"size:500" json:"api_key"`
	BaseURL     string    `gorm:"size:200" json:"base_url"`
	Model       string    `gorm:"size:100;not null" json:"model"`
	MaxTokens   int       `gorm:"default:4000" json:"max_tokens"`
	Temperature float64   `gorm:"default:0.7" json:"temperature"`
	IsDefault   bool      `gorm:"default:false" json:"is_default"`
	IsEnabled   bool      `gorm:"default:true" json:"is_enabled"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (LLMConfig) TableName() string {
	return "aiops_llm_configs"
}
