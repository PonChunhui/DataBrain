package v1

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tmc/langchaingo/llms"
	"go.uber.org/zap"

	"devops-backend/global"
	"devops-backend/model"
	"devops-backend/model/request"
	"devops-backend/service"
	"devops-backend/service/aiops"
)

type AIOPSClusterInspectionController struct{}

func (c *AIOPSClusterInspectionController) StreamInspection(ctx *gin.Context) {
	var req request.ClusterInspectionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, gin.H{"code": 400, "msg": "参数错误: " + err.Error(), "data": nil})
		return
	}

	llmConfig, err := aiops.NewLLMConfigService().GetDefaultConfig()
	if req.LLMConfigID > 0 {
		llmConfig, err = aiops.NewLLMConfigService().GetConfigByID(req.LLMConfigID)
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

	clusterData, err := collectClusterInspectionData(req.ClusterID)
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{"code": 400, "msg": "集群数据采集失败: " + err.Error(), "data": nil})
		return
	}

	dataJSON, err := json.Marshal(clusterData)
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{"code": 400, "msg": "数据序列化失败", "data": nil})
		return
	}

	userID := ctx.GetUint("user_id")
	username := ctx.GetString("username")

	record := &model.DiagnosticRecord{
		ClusterID:       req.ClusterID,
		ClusterName:     cluster.Alias,
		Namespace:       "cluster-wide",
		ResourceType:    "cluster",
		ResourceName:    cluster.Name,
		Status:          "running",
		TriggeredBy:     userID,
		TriggeredByName: username,
		LLMConfigID:     llmConfig.ID,
		LLMProvider:     llmConfig.Provider,
		LLMModel:        llmConfig.Model,
		InputData:       string(dataJSON),
		CreatedAt:       time.Now(),
	}

	if err := global.GVA_DB.Create(record).Error; err != nil {
		ctx.JSON(http.StatusOK, gin.H{"code": 500, "msg": "创建巡检记录失败: " + err.Error(), "data": nil})
		return
	}

	ctx.Header("Content-Type", "text/event-stream")
	ctx.Header("Cache-Control", "no-cache")
	ctx.Header("Connection", "keep-alive")
	ctx.Header("X-Accel-Buffering", "no")

	streamCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	fullContent := ""

	systemPrompt := `你是一个 Kubernetes 集群巡检专家。请根据提供的集群信息和性能指标进行全面巡检分析。

巡检要点：
1. 节点健康状态：检查节点是否正常，资源使用情况
2. Pod状态：检查异常Pod
3. 资源配额：检查资源限制
4. 性能优化：识别性能瓶颈

指标分析指南：
- 节点 CPU usage > 70%：负载较高
- 节点 Memory usage > 80%：内存压力大
- 集群整体 CPU/memory usage > 60%：资源紧张

输出要求：
- 使用Markdown格式
- 按优先级排序问题
- 给出具体解决方案
- 健康评分（0-100分）`

	userPrompt := fmt.Sprintf(`请对以下集群进行巡检分析：

**集群**: %s

**数据**:
%s`, cluster.Alias, string(dataJSON))

	messages := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem, systemPrompt),
		llms.TextParts(llms.ChatMessageTypeHuman, userPrompt),
	}

	sendEvent := func(eventType string, data interface{}) {
		jsonData, _ := json.Marshal(data)
		ctx.SSEvent(eventType, string(jsonData))
		ctx.Writer.Flush()
	}

	sendEvent("start", map[string]interface{}{
		"id":      record.ID,
		"cluster": cluster.Alias,
	})

	err = aiops.StreamLLM(streamCtx, llmConfig, messages, func(chunk string) {
		if chunk != "" {
			fullContent += chunk
			sendEvent("message", map[string]interface{}{
				"content": chunk,
			})
		}
	})

	if err != nil {
		record.Status = "failed"
		record.ErrorMessage = err.Error()
		global.GVA_DB.Save(record)
		sendEvent("error", map[string]interface{}{
			"message": err.Error(),
		})
		return
	}

	record.RawResponse = fullContent
	record.Status = "completed"
	record.Duration = int(time.Since(record.CreatedAt).Milliseconds())
	global.GVA_DB.Save(record)

	sendEvent("done", map[string]interface{}{
		"id":       record.ID,
		"duration": record.Duration,
	})
}

func (c *AIOPSClusterInspectionController) StreamInspectionChat(ctx *gin.Context) {
	var req request.ChatRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, gin.H{"code": 400, "msg": "参数错误: " + err.Error(), "data": nil})
		return
	}

	llmConfig, err := aiops.NewLLMConfigService().GetDefaultConfig()
	if req.LLMConfigID > 0 {
		llmConfig, err = aiops.NewLLMConfigService().GetConfigByID(req.LLMConfigID)
	}

	if err != nil || !llmConfig.IsEnabled {
		ctx.JSON(http.StatusOK, gin.H{"code": 400, "msg": "LLM配置不可用", "data": nil})
		return
	}

	var record model.DiagnosticRecord
	if err := global.GVA_DB.First(&record, req.RecordID).Error; err != nil {
		ctx.JSON(http.StatusOK, gin.H{"code": 400, "msg": "巡检记录不存在", "data": nil})
		return
	}

	ctx.Header("Content-Type", "text/event-stream")
	ctx.Header("Cache-Control", "no-cache")
	ctx.Header("Connection", "keep-alive")
	ctx.Header("X-Accel-Buffering", "no")

	streamCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	messages := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem, `你是Kubernetes集群巡检专家。继续对话，使用Markdown格式。`),
	}

	if record.RawResponse != "" {
		messages = append(messages, llms.TextParts(llms.ChatMessageTypeAI, record.RawResponse))
	}

	messages = append(messages, llms.TextParts(llms.ChatMessageTypeHuman, req.Message))

	sendEvent := func(eventType string, data interface{}) {
		jsonData, _ := json.Marshal(data)
		ctx.SSEvent(eventType, string(jsonData))
		ctx.Writer.Flush()
	}

	sendEvent("start", map[string]interface{}{})

	fullContent := ""
	err = aiops.StreamLLM(streamCtx, llmConfig, messages, func(chunk string) {
		if chunk != "" {
			fullContent += chunk
			sendEvent("message", map[string]interface{}{
				"content": chunk,
			})
		}
	})

	if err != nil {
		sendEvent("error", map[string]interface{}{
			"message": err.Error(),
		})
		return
	}

	record.RawResponse = fullContent
	global.GVA_DB.Save(&record)

	sendEvent("done", map[string]interface{}{
		"id": record.ID,
	})
}

func collectClusterInspectionData(clusterID uint) (map[string]interface{}, error) {
	statsService := &service.StatsService{}
	clusterService := &service.ClusterService{}
	podService := &service.PodService{}
	promService := &service.PrometheusService{}

	stats, err := statsService.GetClusterStats(clusterID)
	if err != nil {
		return nil, fmt.Errorf("获取集群统计失败: %v", err)
	}

	namespaces, err := clusterService.GetNamespaces(clusterID)
	if err != nil {
		return nil, fmt.Errorf("获取命名空间失败: %v", err)
	}

	podStats := map[string]interface{}{
		"total":   stats.PodCount,
		"running": stats.RunningPods,
		"pending": stats.PendingPods,
		"failed":  stats.FailedPods,
	}

	problemPods := []map[string]interface{}{}
	for _, ns := range namespaces {
		pods, err := podService.GetPods(clusterID, ns)
		if err != nil {
			continue
		}
		for _, pod := range pods {
			if pod.Status != "Running" && pod.Status != "Succeeded" {
				problemPods = append(problemPods, map[string]interface{}{
					"name":      pod.Name,
					"namespace": pod.Namespace,
					"status":    pod.Status,
					"restarts":  pod.Restarts,
				})
			}
		}
	}
	podStats["problem_pods"] = problemPods

	nodeList := []map[string]interface{}{}
	nodeMetricsList := []map[string]interface{}{}
	for _, node := range stats.Nodes {
		nodeData := map[string]interface{}{
			"name":         node.Name,
			"ip":           node.IP,
			"status":       node.Status,
			"role":         node.Role,
			"cpu_capacity": node.CpuCapacity,
			"mem_capacity": node.MemCapacity,
			"pod_count":    node.PodCount,
		}

		nodeMetrics, err := promService.GetNodeMetrics(clusterID, node.Name, 30)
		if err == nil && nodeMetrics != nil {
			nodeData["cpu_usage"] = nodeMetrics["cpu_usage"]
			nodeData["memory_usage"] = nodeMetrics["memory_usage"]
			nodeMetricsList = append(nodeMetricsList, map[string]interface{}{
				"name":         node.Name,
				"cpu_usage":    nodeMetrics["cpu_usage"],
				"memory_usage": nodeMetrics["memory_usage"],
			})
		} else if err != nil {
			global.GVA_LOG.Warn("获取节点指标失败", zap.String("node", node.Name), zap.Error(err))
		}

		nodeList = append(nodeList, nodeData)
	}

	clusterMetrics, err := promService.GetClusterMetrics(clusterID, 30)
	metricsData := map[string]interface{}{}
	if err == nil && clusterMetrics != nil {
		metricsData = clusterMetrics
	} else if err != nil {
		global.GVA_LOG.Warn("获取集群指标失败", zap.Error(err))
	}
	metricsData["node_metrics"] = nodeMetricsList

	return map[string]interface{}{
		"nodes":           nodeList,
		"node_count":      len(stats.Nodes),
		"namespaces":      namespaces,
		"namespace_count": stats.NamespaceCount,
		"pod_stats":       podStats,
		"metrics":         metricsData,
		"inspection_time": time.Now().Format(time.RFC3339),
	}, nil
}
