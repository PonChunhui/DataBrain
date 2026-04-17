package v1

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tmc/langchaingo/llms"

	"devops-backend/global"
	"devops-backend/model"
	"devops-backend/model/request"
	"devops-backend/service/aiops"
)

type DiagnosticReport struct {
	ProblemDesc    string   `json:"problem_desc"`
	RootCause      string   `json:"root_cause"`
	PossibleCauses []string `json:"possible_causes"`
	Solution       string   `json:"solution"`
	SolutionSteps  []string `json:"solution_steps"`
	ImpactScope    string   `json:"impact_scope"`
	Severity       string   `json:"severity"`
	Confidence     float64  `json:"confidence"`
}

func parseDiagnosticReport(content string) *DiagnosticReport {
	report := &DiagnosticReport{
		Severity:   "medium",
		Confidence: 0.7,
	}

	jsonRegex := regexp.MustCompile("```json\\s*\\n?(\\{[\\s\\S]*?\\})\\s*\\n?```")
	match := jsonRegex.FindStringSubmatch(content)
	if len(match) > 1 {
		var parsed DiagnosticReport
		if err := json.Unmarshal([]byte(match[1]), &parsed); err == nil {
			return &parsed
		}
	}

	problemPatterns := []string{
		"(?i)(?:整体状态概览|状态概览|问题现象|问题描述)[\\s]*\\n+[\\s]*\\n+([\\s\\S]{50,500}?)(?=\\n##|\\n---|$)",
		"(?i)Pod.*?处于.*?状态",
		"(?i)容器.*?CrashLoopBackOff",
	}
	for _, pattern := range problemPatterns {
		re := regexp.MustCompile(pattern)
		if m := re.FindStringSubmatch(content); len(m) > 1 && strings.TrimSpace(m[1]) != "" {
			report.ProblemDesc = strings.TrimSpace(m[1])
			break
		} else if len(m) > 0 && report.ProblemDesc == "" {
			report.ProblemDesc = strings.TrimSpace(m[0])
			break
		}
	}
	if report.ProblemDesc == "" {
		firstPara := regexp.MustCompile("我们来.*?\\n+\\n+([^\\n]+)")
		if m := firstPara.FindStringSubmatch(content); len(m) > 1 {
			report.ProblemDesc = strings.TrimSpace(m[1])
		}
	}

	rootCausePatterns := []string{
		"(?i)(?:根因总结|根本原因|Root Cause)[\\s]*\\n+[\\s]*\\n+([\\s\\S]{50,300}?)(?=\\n##|\\n---)",
		"(?i)根本原因[：:][\\s]*(.+?)(?:\\n|$)",
		"(?i)✅.*?根本原因.*?：(.+?)(?:\\n|$)",
	}
	for _, pattern := range rootCausePatterns {
		re := regexp.MustCompile(pattern)
		if m := re.FindStringSubmatch(content); len(m) > 1 && strings.TrimSpace(m[1]) != "" {
			report.RootCause = strings.TrimSpace(m[1])
			break
		}
	}

	solutionPatterns := []string{
		"(?i)(?:解决方案|Solution)[\\s]*\\n+[\\s]*\\n+([\\s\\S]{50,500}?)(?=\\n##|\\n---)",
		"(?i)解决方案[：:][\\s]*(.+?)(?:\\n|$)",
	}
	for _, pattern := range solutionPatterns {
		re := regexp.MustCompile(pattern)
		if m := re.FindStringSubmatch(content); len(m) > 1 && strings.TrimSpace(m[1]) != "" {
			report.Solution = strings.TrimSpace(m[1])
			break
		}
	}

	impactPatterns := []string{
		"(?i)(?:影响范围|影响评估|影响)[\\s]*\\n+[\\s]*\\n+([\\s\\S]{20,200}?)(?=\\n##|\\n---|\\n\\|)",
		"(?i)影响范围[：:][\\s]*(.+?)(?:\\n|$)",
	}
	for _, pattern := range impactPatterns {
		re := regexp.MustCompile(pattern)
		if m := re.FindStringSubmatch(content); len(m) > 1 && strings.TrimSpace(m[1]) != "" {
			report.ImpactScope = strings.TrimSpace(m[1])
			break
		}
	}

	severityKeywords := map[string]string{
		"critical": "严重|高危|紧急|critical|P0|🔴",
		"high":     "重要|高风险|high|P1|🟡",
		"low":      "低风险|一般|low|P2|🟢",
	}
	for severity, keywords := range severityKeywords {
		if regexp.MustCompile(keywords).MatchString(content) {
			report.Severity = severity
			break
		}
	}

	if report.ProblemDesc == "" {
		report.ProblemDesc = "诊断分析已完成，详见完整报告"
	}
	if report.RootCause == "" {
		report.RootCause = "请查看完整诊断报告中的根因分析部分"
	}
	if report.Solution == "" {
		report.Solution = "请查看完整诊断报告中的解决方案部分"
	}
	if report.ImpactScope == "" {
		report.ImpactScope = "请查看完整诊断报告"
	}

	return report
}

func (c *AIOPSDiagnosticController) StreamDiagnose(ctx *gin.Context) {
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

	dataCollector := aiops.NewDataCollector()
	input, err := dataCollector.Collect(context.Background(), req.ClusterID, req.Namespace, req.ResourceType, req.ResourceName)
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{"code": 400, "msg": "数据采集失败: " + err.Error(), "data": nil})
		return
	}

	inputJSON, err := dataCollector.ToJSON(input)
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{"code": 400, "msg": "数据序列化失败: " + err.Error(), "data": nil})
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
		Status:          "running",
		TriggeredBy:     userID,
		TriggeredByName: username,
		LLMConfigID:     llmConfig.ID,
		LLMProvider:     llmConfig.Provider,
		LLMModel:        llmConfig.Model,
		InputData:       inputJSON,
		CreatedAt:       time.Now(),
	}

	if err := global.GVA_DB.Create(record).Error; err != nil {
		ctx.JSON(http.StatusOK, gin.H{"code": 500, "msg": "创建诊断记录失败: " + err.Error(), "data": nil})
		return
	}

	ctx.Header("Content-Type", "text/event-stream")
	ctx.Header("Cache-Control", "no-cache")
	ctx.Header("Connection", "keep-alive")
	ctx.Header("X-Accel-Buffering", "no")

	streamCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	fullContent := ""

	systemPrompt := "你是一个 Kubernetes 故障诊断专家。请根据用户提供的资源信息和性能指标分析问题，给出专业的诊断分析。\n\n" +
		"回答要求：\n" +
		"1. 首先简要描述当前资源的状态和问题现象\n" +
		"2. 分析 Prometheus 性能指标（如有）：CPU、内存、网络、throttling等\n" +
		"3. 分析可能的根本原因（结合状态、事件、日志、指标综合判断）\n" +
		"4. 提供具体的解决方案和操作步骤\n" +
		"5. 使用 Markdown 格式，结构清晰\n\n" +
		"**重要**：在回答末尾，请提供JSON格式的诊断摘要（代码块包裹），格式如下：\n" +
		"{\n" +
		"  \"problem_desc\": \"问题现象描述\",\n" +
		"  \"root_cause\": \"根本原因\",\n" +
		"  \"possible_causes\": [\"原因1\", \"原因2\"],\n" +
		"  \"solution\": \"解决方案\",\n" +
		"  \"solution_steps\": [\"步骤1\", \"步骤2\"],\n" +
		"  \"impact_scope\": \"影响范围\",\n" +
		"  \"severity\": \"critical/high/medium/low\",\n" +
		"  \"confidence\": 0.8\n" +
		"}"

	userPrompt := fmt.Sprintf(`请分析以下 Kubernetes 资源的问题：

**资源类型**: %s
**资源名称**: %s  
**命名空间**: %s
**集群**: %s

**资源详情**:
%s

请给出详细的诊断分析。`, req.ResourceType, req.ResourceName, req.Namespace, cluster.Alias, inputJSON)

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
		"id":        record.ID,
		"resource":  req.ResourceName,
		"namespace": req.Namespace,
		"cluster":   cluster.Alias,
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

	report := parseDiagnosticReport(fullContent)
	record.ProblemDesc = report.ProblemDesc
	record.RootCause = report.RootCause
	record.Solution = report.Solution
	record.ImpactScope = report.ImpactScope
	record.Severity = report.Severity
	record.Confidence = report.Confidence
	if len(report.PossibleCauses) > 0 {
		record.PossibleCauses = strings.Join(report.PossibleCauses, "\n")
	}
	if len(report.SolutionSteps) > 0 {
		record.SolutionSteps = strings.Join(report.SolutionSteps, "\n")
	}

	global.GVA_DB.Save(record)

	sendEvent("done", map[string]interface{}{
		"id":       record.ID,
		"duration": record.Duration,
	})
}

func (c *AIOPSDiagnosticController) StreamChat(ctx *gin.Context) {
	var req request.ChatRequest
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
		ctx.JSON(http.StatusOK, gin.H{"code": 400, "msg": "LLM配置不存在", "data": nil})
		return
	}

	if !llmConfig.IsEnabled {
		ctx.JSON(http.StatusOK, gin.H{"code": 400, "msg": "LLM配置已禁用", "data": nil})
		return
	}

	var record model.DiagnosticRecord
	if err := global.GVA_DB.First(&record, req.RecordID).Error; err != nil {
		ctx.JSON(http.StatusOK, gin.H{"code": 400, "msg": "诊断记录不存在", "data": nil})
		return
	}

	ctx.Header("Content-Type", "text/event-stream")
	ctx.Header("Cache-Control", "no-cache")
	ctx.Header("Connection", "keep-alive")
	ctx.Header("X-Accel-Buffering", "no")

	streamCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	messages := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem, `你是一个 Kubernetes 故障诊断专家。继续与用户对话，解答关于 Kubernetes 故障诊断的问题。回答要专业、具体，使用 Markdown 格式。`),
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
