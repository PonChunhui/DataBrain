package aiops

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"devops-backend/global"
	"devops-backend/model"
	"go.uber.org/zap"
)

type DiagnosticService struct {
	dataCollector *DataCollector
}

func NewDiagnosticService() *DiagnosticService {
	return &DiagnosticService{
		dataCollector: NewDataCollector(),
	}
}

func (s *DiagnosticService) Diagnose(ctx context.Context, recordID uint, llmConfig *model.LLMConfig) error {
	var record model.DiagnosticRecord
	if err := global.GVA_DB.First(&record, recordID).Error; err != nil {
		return err
	}

	record.Status = "running"
	global.GVA_DB.Save(&record)

	startTime := time.Now()

	defer func() {
		record.Duration = int(time.Since(startTime).Milliseconds())
		global.GVA_DB.Save(&record)
	}()

	input, err := s.dataCollector.Collect(ctx, record.ClusterID, record.Namespace, record.ResourceType, record.ResourceName)
	if err != nil {
		record.Status = "failed"
		record.ErrorMessage = fmt.Sprintf("数据采集失败: %v", err)
		return err
	}

	inputJSON, err := s.dataCollector.ToJSON(input)
	if err != nil {
		record.Status = "failed"
		record.ErrorMessage = fmt.Sprintf("数据序列化失败: %v", err)
		return err
	}
	record.InputData = inputJSON

	prompt := s.buildPrompt(inputJSON)

	response, err := CallLLM(ctx, llmConfig, prompt)
	if err != nil {
		record.Status = "failed"
		record.ErrorMessage = fmt.Sprintf("LLM调用失败: %v", err)
		return err
	}

	record.RawResponse = response.Content

	report, err := s.parseResponse(response.Content)
	if err != nil {
		record.Status = "failed"
		record.ErrorMessage = fmt.Sprintf("响应解析失败: %v", err)
		return err
	}

	record.ProblemDesc = report.ProblemDesc
	record.RootCause = report.RootCause
	if len(report.PossibleCauses) > 0 {
		causesJSON, _ := json.Marshal(report.PossibleCauses)
		record.PossibleCauses = string(causesJSON)
	}
	record.Solution = report.Solution
	if len(report.SolutionSteps) > 0 {
		stepsJSON, _ := json.Marshal(report.SolutionSteps)
		record.SolutionSteps = string(stepsJSON)
	}
	record.ImpactScope = report.ImpactScope
	if len(report.RelatedRes) > 0 {
		resJSON, _ := json.Marshal(report.RelatedRes)
		record.RelatedRes = string(resJSON)
	}
	record.Severity = report.Severity
	record.Confidence = report.Confidence

	record.Status = "completed"

	return nil
}

func (s *DiagnosticService) buildPrompt(inputJSON string) string {
	return fmt.Sprintf(`你是一个 Kubernetes 故障诊断专家。请根据以下资源信息分析问题并提供诊断报告。

资源信息:
%s

请以JSON格式输出诊断报告，包含以下字段:
{
    "problem_desc": "问题描述(简明扼要描述当前问题现象)",
    "root_cause": "根本原因分析(分析导致问题的根本原因)",
    "possible_causes": ["可能原因1", "可能原因2", ...],
    "solution": "解决方案概述",
    "solution_steps": ["步骤1: xxx", "步骤2: xxx", ...],
    "impact_scope": "影响范围(受影响的资源和服务)",
    "related_resources": [{"type": "xxx", "name": "xxx", "reason": "xxx"}, ...],
    "severity": "high/medium/low",
    "confidence": 0.85
}

注意:
1. 分析要基于实际采集的数据，不要臆测
2. 解决方案要具体可操作，包含具体的kubectl命令或配置修改
3. 如果数据不足以诊断，请明确说明需要哪些额外信息
4. 置信度根据数据完整性和分析确定性评估(0-1之间的数值)
5. 只输出JSON，不要输出其他内容`, inputJSON)
}

type DiagnosticReport struct {
	ProblemDesc    string                   `json:"problem_desc"`
	RootCause      string                   `json:"root_cause"`
	PossibleCauses []string                 `json:"possible_causes"`
	Solution       string                   `json:"solution"`
	SolutionSteps  []string                 `json:"solution_steps"`
	ImpactScope    string                   `json:"impact_scope"`
	RelatedRes     []map[string]interface{} `json:"related_resources"`
	Severity       string                   `json:"severity"`
	Confidence     float64                  `json:"confidence"`
}

func (s *DiagnosticService) parseResponse(content string) (*DiagnosticReport, error) {
	jsonStart := strings.Index(content, "{")
	jsonEnd := strings.LastIndex(content, "}")
	if jsonStart == -1 || jsonEnd == -1 {
		return nil, fmt.Errorf("response does not contain valid JSON")
	}

	jsonStr := content[jsonStart : jsonEnd+1]

	var report DiagnosticReport
	if err := json.Unmarshal([]byte(jsonStr), &report); err != nil {
		return nil, fmt.Errorf("parse JSON failed: %v", err)
	}

	if report.Confidence == 0 {
		report.Confidence = 0.75
	}

	if report.Severity == "" {
		report.Severity = "medium"
	}

	return &report, nil
}

type LLMConfigService struct{}

func NewLLMConfigService() *LLMConfigService {
	return &LLMConfigService{}
}

func (s *LLMConfigService) GetConfigs() ([]model.LLMConfig, error) {
	var configs []model.LLMConfig
	err := global.GVA_DB.Order("is_default desc, created_at desc").Find(&configs).Error
	return configs, err
}

func (s *LLMConfigService) GetConfigByID(id uint) (*model.LLMConfig, error) {
	var config model.LLMConfig
	err := global.GVA_DB.First(&config, id).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func (s *LLMConfigService) GetDefaultConfig() (*model.LLMConfig, error) {
	var config model.LLMConfig
	err := global.GVA_DB.Where("is_default = ? AND is_enabled = ?", true, true).First(&config).Error
	if err != nil {
		return nil, fmt.Errorf("no default LLM config found")
	}
	return &config, nil
}

func (s *LLMConfigService) CreateConfig(config *model.LLMConfig) error {
	if config.MaxTokens == 0 {
		config.MaxTokens = 4000
	}
	if config.Temperature == 0 {
		config.Temperature = 0.7
	}

	if config.IsDefault {
		global.GVA_DB.Model(&model.LLMConfig{}).Where("is_default = ?", true).Update("is_default", false)
	}

	return global.GVA_DB.Create(config).Error
}

func (s *LLMConfigService) UpdateConfig(config *model.LLMConfig) error {
	if config.IsDefault {
		global.GVA_DB.Model(&model.LLMConfig{}).Where("is_default = ? AND id != ?", true, config.ID).Update("is_default", false)
	}

	return global.GVA_DB.Save(config).Error
}

func (s *LLMConfigService) DeleteConfig(id uint) error {
	var config model.LLMConfig
	if err := global.GVA_DB.First(&config, id).Error; err != nil {
		return err
	}

	if config.IsDefault {
		var otherConfig model.LLMConfig
		if global.GVA_DB.Where("id != ? AND is_enabled = ?", id, true).First(&otherConfig).Error == nil {
			global.GVA_DB.Model(&otherConfig).Update("is_default", true)
		}
	}

	return global.GVA_DB.Delete(&config).Error
}

func (s *LLMConfigService) TestConfig(ctx context.Context, id uint) (map[string]interface{}, error) {
	config, err := s.GetConfigByID(id)
	if err != nil {
		return nil, err
	}

	global.GVA_LOG.Info("TestConfig", zap.String("provider", config.Provider), zap.String("baseURL", config.BaseURL))

	startTime := time.Now()
	response, err := CallLLM(ctx, config, "你好，请回复'连接成功'。")
	if err != nil {
		global.GVA_LOG.Error("LLM call failed", zap.Error(err))
		return nil, err
	}

	return map[string]interface{}{
		"success":       true,
		"model":         response.Model,
		"response_time": int(time.Since(startTime).Milliseconds()),
		"test_response": response.Content,
		"tokens_used":   response.TokensUsed,
	}, nil
}

func (s *LLMConfigService) SetDefault(id uint) error {
	var config model.LLMConfig
	if err := global.GVA_DB.First(&config, id).Error; err != nil {
		return err
	}

	global.GVA_DB.Model(&model.LLMConfig{}).Where("is_default = ?", true).Update("is_default", false)

	return global.GVA_DB.Model(&config).Update("is_default", true).Error
}

type DiagnosticHistoryService struct{}

func NewDiagnosticHistoryService() *DiagnosticHistoryService {
	return &DiagnosticHistoryService{}
}

func (s *DiagnosticHistoryService) GetHistory(query interface{}) ([]model.DiagnosticRecord, int64, error) {
	return nil, 0, nil
}

func (s *DiagnosticHistoryService) GetByID(id uint) (*model.DiagnosticRecord, error) {
	var record model.DiagnosticRecord
	err := global.GVA_DB.First(&record, id).Error
	if err != nil {
		return nil, err
	}
	return &record, nil
}

func (s *DiagnosticHistoryService) GetStats() (map[string]interface{}, error) {
	var totalCount int64
	global.GVA_DB.Model(&model.DiagnosticRecord{}).Count(&totalCount)

	var weekCount int64
	weekStart := time.Now().AddDate(0, 0, -7)
	global.GVA_DB.Model(&model.DiagnosticRecord{}).Where("created_at >= ?", weekStart).Count(&weekCount)

	var highCount int64
	global.GVA_DB.Model(&model.DiagnosticRecord{}).Where("severity = ?", "high").Count(&highCount)

	var avgConfidence float64
	global.GVA_DB.Model(&model.DiagnosticRecord{}).Where("status = ?", "completed").Select("AVG(confidence)").Scan(&avgConfidence)

	return map[string]interface{}{
		"total_count":    totalCount,
		"week_count":     weekCount,
		"high_count":     highCount,
		"avg_confidence": strconv.FormatFloat(avgConfidence*100, 'f', 0, 64) + "%",
	}, nil
}
