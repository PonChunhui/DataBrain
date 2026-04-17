package v1

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"devops-backend/global"
	"devops-backend/model"
	"devops-backend/model/request"
	"devops-backend/service/aiops"
)

type AIOPSDiagnosticController struct{}

var diagnosticService = aiops.NewDiagnosticService()
var historyService = aiops.NewDiagnosticHistoryService()

func (c *AIOPSDiagnosticController) Diagnose(ctx *gin.Context) {
	var req request.DiagnosticRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, gin.H{"code": 400, "msg": "参数错误: " + err.Error(), "data": nil})
		return
	}

	var llmConfig *model.LLMConfig
	var err error

	if req.LLMConfigID > 0 {
		llmConfig, err = llmConfigService.GetConfigByID(req.LLMConfigID)
	} else {
		llmConfig, err = llmConfigService.GetDefaultConfig()
	}

	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{"code": 400, "msg": "LLM配置不存在，请先配置LLM模型", "data": nil})
		return
	}

	if !llmConfig.IsEnabled {
		ctx.JSON(http.StatusOK, gin.H{"code": 400, "msg": "LLM配置已禁用", "data": nil})
		return
	}

	var cluster model.K8sCluster
	if err := global.GVA_DB.First(&cluster, req.ClusterID).Error; err != nil {
		ctx.JSON(http.StatusOK, gin.H{"code": 400, "msg": "集群不存在", "data": nil})
		return
	}

	userID := ctx.GetUint("user_id")
	username := ctx.GetString("username")

	record := &model.DiagnosticRecord{
		ClusterID:       req.ClusterID,
		ClusterName:     cluster.Alias,
		Namespace:       req.Namespace,
		ResourceType:    req.ResourceType,
		ResourceName:    req.ResourceName,
		Status:          "pending",
		TriggeredBy:     userID,
		TriggeredByName: username,
		LLMConfigID:     llmConfig.ID,
		LLMProvider:     llmConfig.Provider,
		LLMModel:        llmConfig.Model,
		CreatedAt:       time.Now(),
	}

	if err := global.GVA_DB.Create(record).Error; err != nil {
		ctx.JSON(http.StatusOK, gin.H{"code": 500, "msg": "创建诊断记录失败: " + err.Error(), "data": nil})
		return
	}

	go func() {
		diagnosticCtx := context.Background()
		diagnosticService.Diagnose(diagnosticCtx, record.ID, llmConfig)
	}()

	ctx.JSON(http.StatusOK, gin.H{"code": 0, "msg": "诊断任务已启动", "data": gin.H{"id": record.ID}})
}

func (c *AIOPSDiagnosticController) GetDiagnostic(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{"code": 400, "msg": "无效的ID", "data": nil})
		return
	}

	record, err := historyService.GetByID(uint(id))
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{"code": 400, "msg": "诊断记录不存在", "data": nil})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"code": 0, "msg": "success", "data": record})
}

func (c *AIOPSDiagnosticController) GetHistory(ctx *gin.Context) {
	var query request.DiagnosticHistoryQuery
	if err := ctx.ShouldBindQuery(&query); err != nil {
		ctx.JSON(http.StatusOK, gin.H{"code": 400, "msg": "参数错误", "data": nil})
		return
	}

	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = 20
	}

	var records []model.DiagnosticRecord
	var total int64

	db := global.GVA_DB.Model(&model.DiagnosticRecord{})

	if query.ResourceType != "" {
		db = db.Where("resource_type = ?", query.ResourceType)
	}
	if query.Severity != "" {
		db = db.Where("severity = ?", query.Severity)
	}
	if query.LLMProvider != "" {
		db = db.Where("llm_provider = ?", query.LLMProvider)
	}
	if query.ClusterID > 0 {
		db = db.Where("cluster_id = ?", query.ClusterID)
	}
	if query.Namespace != "" {
		db = db.Where("namespace = ?", query.Namespace)
	}
	if query.Keyword != "" {
		db = db.Where("resource_name LIKE ? OR problem_desc LIKE ?", "%"+query.Keyword+"%", "%"+query.Keyword+"%")
	}
	if query.StartTime != "" {
		db = db.Where("created_at >= ?", query.StartTime)
	}
	if query.EndTime != "" {
		db = db.Where("created_at <= ?", query.EndTime)
	}

	db.Count(&total)

	offset := (query.Page - 1) * query.PageSize
	db.Order("created_at desc").Offset(offset).Limit(query.PageSize).Find(&records)

	ctx.JSON(http.StatusOK, gin.H{"code": 0, "msg": "success", "data": gin.H{
		"list":     records,
		"total":    total,
		"page":     query.Page,
		"pageSize": query.PageSize,
	}})
}

func (c *AIOPSDiagnosticController) GetStats(ctx *gin.Context) {
	stats, err := historyService.GetStats()
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{"code": 400, "msg": "获取统计失败", "data": nil})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"code": 0, "msg": "success", "data": stats})
}
