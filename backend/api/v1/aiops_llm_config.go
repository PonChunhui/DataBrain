package v1

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"devops-backend/model"
	"devops-backend/model/request"
	"devops-backend/service/aiops"
)

type AIOPSLLMConfigController struct{}

var llmConfigService = aiops.NewLLMConfigService()

func (c *AIOPSLLMConfigController) GetConfigs(ctx *gin.Context) {
	configs, err := llmConfigService.GetConfigs()
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{"code": 400, "msg": "获取配置列表失败", "data": nil})
		return
	}

	for i := range configs {
		if len(configs[i].APIKey) > 10 {
			configs[i].APIKey = configs[i].APIKey[:6] + "***..." + configs[i].APIKey[len(configs[i].APIKey)-4:]
		}
	}

	ctx.JSON(http.StatusOK, gin.H{"code": 0, "msg": "success", "data": configs})
}

func (c *AIOPSLLMConfigController) GetConfig(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{"code": 400, "msg": "无效的ID", "data": nil})
		return
	}

	config, err := llmConfigService.GetConfigByID(uint(id))
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{"code": 400, "msg": "配置不存在", "data": nil})
		return
	}

	if len(config.APIKey) > 10 {
		config.APIKey = config.APIKey[:6] + "***..." + config.APIKey[len(config.APIKey)-4:]
	}

	ctx.JSON(http.StatusOK, gin.H{"code": 0, "msg": "success", "data": config})
}

func (c *AIOPSLLMConfigController) CreateConfig(ctx *gin.Context) {
	var req request.LLMConfigCreate
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, gin.H{"code": 400, "msg": "参数错误: " + err.Error(), "data": nil})
		return
	}

	config := &model.LLMConfig{
		Name:        req.Name,
		Provider:    req.Provider,
		APIKey:      req.APIKey,
		BaseURL:     req.BaseURL,
		Model:       req.Model,
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
		IsDefault:   req.IsDefault,
		IsEnabled:   req.IsEnabled,
	}

	if err := llmConfigService.CreateConfig(config); err != nil {
		ctx.JSON(http.StatusOK, gin.H{"code": 400, "msg": "创建失败: " + err.Error(), "data": nil})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"code": 0, "msg": "创建成功", "data": gin.H{"id": config.ID}})
}

func (c *AIOPSLLMConfigController) UpdateConfig(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{"code": 400, "msg": "无效的ID", "data": nil})
		return
	}

	var req request.LLMConfigUpdate
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, gin.H{"code": 400, "msg": "参数错误: " + err.Error(), "data": nil})
		return
	}

	config, err := llmConfigService.GetConfigByID(uint(id))
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{"code": 400, "msg": "配置不存在", "data": nil})
		return
	}

	if req.Name != "" {
		config.Name = req.Name
	}
	if req.APIKey != "" {
		config.APIKey = req.APIKey
	}
	if req.BaseURL != "" {
		config.BaseURL = req.BaseURL
	}
	if req.Model != "" {
		config.Model = req.Model
	}
	if req.MaxTokens > 0 {
		config.MaxTokens = req.MaxTokens
	}
	if req.Temperature > 0 {
		config.Temperature = req.Temperature
	}
	config.IsDefault = req.IsDefault
	config.IsEnabled = req.IsEnabled

	if err := llmConfigService.UpdateConfig(config); err != nil {
		ctx.JSON(http.StatusOK, gin.H{"code": 400, "msg": "更新失败: " + err.Error(), "data": nil})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"code": 0, "msg": "更新成功", "data": nil})
}

func (c *AIOPSLLMConfigController) DeleteConfig(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{"code": 400, "msg": "无效的ID", "data": nil})
		return
	}

	if err := llmConfigService.DeleteConfig(uint(id)); err != nil {
		ctx.JSON(http.StatusOK, gin.H{"code": 400, "msg": "删除失败: " + err.Error(), "data": nil})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"code": 0, "msg": "删除成功", "data": nil})
}

func (c *AIOPSLLMConfigController) TestConfig(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{"code": 400, "msg": "无效的ID", "data": nil})
		return
	}

	result, err := llmConfigService.TestConfig(ctx.Request.Context(), uint(id))
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{"code": 400, "msg": "测试失败: " + err.Error(), "data": gin.H{
			"success": false,
			"error":   err.Error(),
		}})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"code": 0, "msg": "测试成功", "data": result})
}

func (c *AIOPSLLMConfigController) SetDefault(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{"code": 400, "msg": "无效的ID", "data": nil})
		return
	}

	if err := llmConfigService.SetDefault(uint(id)); err != nil {
		ctx.JSON(http.StatusOK, gin.H{"code": 400, "msg": "设置失败: " + err.Error(), "data": nil})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"code": 0, "msg": "设置成功", "data": nil})
}
