package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"devops-backend/global"
	"devops-backend/model"
	"go.uber.org/zap"
)

type PrometheusService struct{}

type PrometheusQueryResult struct {
	Status string `json:"status"`
	Data   struct {
		ResultType string `json:"resultType"`
		Result     []struct {
			Metric map[string]string `json:"metric"`
			Values [][]interface{}   `json:"values"`
			Value  []interface{}     `json:"value"`
		} `json:"result"`
	} `json:"data"`
}

type PrometheusInstantResult struct {
	Status string `json:"status"`
	Data   struct {
		ResultType string `json:"resultType"`
		Result     []struct {
			Metric map[string]interface{} `json:"metric"`
			Value  []interface{}          `json:"value"`
		} `json:"result"`
	} `json:"data"`
}

func (s *PrometheusService) QueryRange(prometheusURL, query, start, end, step string, authEnabled bool, authUser, authPass string) (*PrometheusQueryResult, error) {
	u, err := url.Parse(prometheusURL)
	if err != nil {
		return nil, fmt.Errorf("解析Prometheus URL失败: %v", err)
	}

	queryURL := u.Scheme + "://" + u.Host + "/api/v1/query_range"
	params := url.Values{}
	params.Set("query", query)
	params.Set("start", start)
	params.Set("end", end)
	params.Set("step", step)

	fullURL := queryURL + "?" + params.Encode()

	global.GVA_LOG.Info("Prometheus QueryRange", zap.String("url", fullURL), zap.String("query", query))

	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	if authEnabled {
		req.SetBasicAuth(authUser, authPass)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求Prometheus失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Prometheus返回错误: %d, %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	var result PrometheusQueryResult
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	global.GVA_LOG.Info("Prometheus QueryRange结果",
		zap.String("status", result.Status),
		zap.Int("resultCount", len(result.Data.Result)))

	return &result, nil
}

func (s *PrometheusService) QueryInstant(prometheusURL, query string, authEnabled bool, authUser, authPass string) (*PrometheusInstantResult, error) {
	u, err := url.Parse(prometheusURL)
	if err != nil {
		return nil, fmt.Errorf("解析Prometheus URL失败: %v", err)
	}

	queryURL := u.Scheme + "://" + u.Host + "/api/v1/query"
	params := url.Values{}
	params.Set("query", query)

	fullURL := queryURL + "?" + params.Encode()

	global.GVA_LOG.Info("Prometheus QueryInstant", zap.String("url", fullURL), zap.String("query", query))

	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	if authEnabled {
		req.SetBasicAuth(authUser, authPass)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求Prometheus失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Prometheus返回错误: %d, %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	var result PrometheusInstantResult
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	global.GVA_LOG.Info("Prometheus QueryInstant结果",
		zap.String("status", result.Status),
		zap.Int("resultCount", len(result.Data.Result)))

	return &result, nil
}

type AggregatedMetrics struct {
	CPU       *CPUMetrics         `json:"cpu,omitempty"`
	Memory    *MemoryMetrics      `json:"memory,omitempty"`
	Network   *NetworkMetrics     `json:"network,omitempty"`
	Disk      *DiskMetrics        `json:"disk,omitempty"`
	Pod       *PodKubeMetrics     `json:"pod_kube,omitempty"`
	Container []*ContainerMetrics `json:"containers,omitempty"`
}

type CPUMetrics struct {
	AvgCore        float64 `json:"avg_core"`
	MaxCore        float64 `json:"max_core"`
	MinCore        float64 `json:"min_core"`
	LimitCore      float64 `json:"limit_core,omitempty"`
	RequestCore    float64 `json:"request_core,omitempty"`
	UsagePercent   float64 `json:"usage_percent"`
	RequestPercent float64 `json:"request_percent,omitempty"`
	Throttling     bool    `json:"throttling"`
	ThrottledSec   float64 `json:"throttled_seconds,omitempty"`
	Trend          string  `json:"trend"`
}

type MemoryMetrics struct {
	AvgMB        float64 `json:"avg_mb"`
	MaxMB        float64 `json:"max_mb"`
	MinMB        float64 `json:"min_mb"`
	LimitMB      float64 `json:"limit_mb,omitempty"`
	RequestMB    float64 `json:"request_mb,omitempty"`
	UsagePercent float64 `json:"usage_percent"`
	OOMRisk      string  `json:"oom_risk"`
	CacheMB      float64 `json:"cache_mb,omitempty"`
	RssMB        float64 `json:"rss_mb,omitempty"`
	Trend        string  `json:"trend"`
}

type NetworkMetrics struct {
	TxAvgKbS  float64 `json:"tx_avg_kb_s"`
	RxAvgKbS  float64 `json:"rx_avg_kb_s"`
	TxMaxKbS  float64 `json:"tx_max_kb_s"`
	RxMaxKbS  float64 `json:"rx_max_kb_s"`
	TxTotalMb float64 `json:"tx_total_mb,omitempty"`
	RxTotalMb float64 `json:"rx_total_mb,omitempty"`
}

type DiskMetrics struct {
	ReadAvgKbS   float64 `json:"read_avg_kb_s"`
	WriteAvgKbS  float64 `json:"write_avg_kb_s"`
	ReadTotalMb  float64 `json:"read_total_mb,omitempty"`
	WriteTotalMb float64 `json:"write_total_mb,omitempty"`
}

type PodKubeMetrics struct {
	StatusPhase         string `json:"status_phase,omitempty"`
	ContainerWaiting    bool   `json:"container_waiting,omitempty"`
	WaitingReason       string `json:"waiting_reason,omitempty"`
	ContainerTerminated bool   `json:"container_terminated,omitempty"`
	TerminatedReason    string `json:"terminated_reason,omitempty"`
	Restarts            int    `json:"restarts,omitempty"`
	Ready               bool   `json:"ready,omitempty"`
	Scheduled           bool   `json:"scheduled,omitempty"`
}

type ContainerMetrics struct {
	Name         string  `json:"name"`
	CPUAvgCore   float64 `json:"cpu_avg_core"`
	CPUMaxCore   float64 `json:"cpu_max_core"`
	MemoryAvgMB  float64 `json:"memory_avg_mb"`
	MemoryMaxMB  float64 `json:"memory_max_mb"`
	CPUThrottled float64 `json:"cpu_throttled_seconds,omitempty"`
}

func (s *PrometheusService) GetAggregatedPodMetrics(clusterID uint, namespace, podName string, durationMinutes int, cpuLimit, memLimit float64) (*AggregatedMetrics, error) {
	cluster, err := clusterService.GetClusterByID(clusterID)
	if err != nil {
		return nil, err
	}

	if cluster.PrometheusUrl == "" {
		global.GVA_LOG.Info("集群未配置Prometheus，跳过指标采集", zap.Uint("clusterID", clusterID))
		return nil, nil
	}

	end := time.Now().Unix()
	start := end - int64(durationMinutes*60)
	step := "30s"
	rateWindow := "60s"

	metrics := &AggregatedMetrics{}
	containerMetrics := []*ContainerMetrics{}

	cpuQueries := []string{
		fmt.Sprintf(`sum(rate(container_cpu_usage_seconds_total{namespace="%s",pod="%s",container!="",container!="POD"}[%s]))`, namespace, podName, rateWindow),
		fmt.Sprintf(`sum(irate(container_cpu_usage_seconds_total{namespace="%s",pod="%s",container!="",container!="POD"}[30s]))`, namespace, podName),
	}

	for _, cpuQuery := range cpuQueries {
		cpuResult, err := s.QueryRange(cluster.PrometheusUrl, cpuQuery, fmt.Sprintf("%d", start), fmt.Sprintf("%d", end), step, cluster.PrometheusAuthEnabled, cluster.PrometheusBasicAuthUser, cluster.PrometheusBasicAuthPass)
		if err == nil && cpuResult.Status == "success" && len(cpuResult.Data.Result) > 0 {
			values := s.extractAllValues(cpuResult)
			if len(values) > 0 {
				avg, max, min, trend := s.calculateStats(values)
				metrics.CPU = &CPUMetrics{
					AvgCore:      avg,
					MaxCore:      max,
					MinCore:      min,
					LimitCore:    cpuLimit,
					UsagePercent: s.calculateUsagePercent(avg, cpuLimit),
					Throttling:   max > cpuLimit*0.9,
					Trend:        trend,
				}
				break
			}
		}
	}

	cpuThrottleQuery := fmt.Sprintf(`sum(rate(container_cpu_cfs_throttled_seconds_total{namespace="%s",pod="%s",container!="",container!="POD"}[%s]))`, namespace, podName, rateWindow)
	throttleResult, err := s.QueryRange(cluster.PrometheusUrl, cpuThrottleQuery, fmt.Sprintf("%d", start), fmt.Sprintf("%d", end), step, cluster.PrometheusAuthEnabled, cluster.PrometheusBasicAuthUser, cluster.PrometheusBasicAuthPass)
	if err == nil && throttleResult.Status == "success" && len(throttleResult.Data.Result) > 0 {
		values := s.extractAllValues(throttleResult)
		if len(values) > 0 && metrics.CPU != nil {
			avg, _, _, _ := s.calculateStats(values)
			metrics.CPU.ThrottledSec = avg
			if avg > 0 {
				metrics.CPU.Throttling = true
			}
		}
	}

	cpuPerContainerQuery := fmt.Sprintf(`rate(container_cpu_usage_seconds_total{namespace="%s",pod="%s",container!="",container!="POD"}[%s])`, namespace, podName, rateWindow)
	cpuContainerResult, err := s.QueryRange(cluster.PrometheusUrl, cpuPerContainerQuery, fmt.Sprintf("%d", start), fmt.Sprintf("%d", end), step, cluster.PrometheusAuthEnabled, cluster.PrometheusBasicAuthUser, cluster.PrometheusBasicAuthPass)
	if err == nil && cpuContainerResult.Status == "success" && len(cpuContainerResult.Data.Result) > 0 {
		containerCPUData := s.extractContainerValues(cpuContainerResult)
		for name, values := range containerCPUData {
			if len(values) > 0 {
				avg, max, _, _ := s.calculateStats(values)
				containerMetrics = append(containerMetrics, &ContainerMetrics{
					Name:       name,
					CPUAvgCore: avg,
					CPUMaxCore: max,
				})
			}
		}
	}

	memQueries := []string{
		fmt.Sprintf(`sum(container_memory_working_set_bytes{namespace="%s",pod="%s",container!="",container!="POD"})`, namespace, podName),
		fmt.Sprintf(`sum(container_memory_usage_bytes{namespace="%s",pod="%s",container!="",container!="POD"})`, namespace, podName),
	}

	for _, memQuery := range memQueries {
		memResult, err := s.QueryRange(cluster.PrometheusUrl, memQuery, fmt.Sprintf("%d", start), fmt.Sprintf("%d", end), step, cluster.PrometheusAuthEnabled, cluster.PrometheusBasicAuthUser, cluster.PrometheusBasicAuthPass)
		if err == nil && memResult.Status == "success" && len(memResult.Data.Result) > 0 {
			values := s.extractAllValues(memResult)
			if len(values) > 0 {
				avgBytes, maxBytes, minBytes, trend := s.calculateStats(values)
				avgMB := avgBytes / 1024 / 1024
				maxMB := maxBytes / 1024 / 1024
				minMB := minBytes / 1024 / 1024
				memLimitMB := memLimit / 1024 / 1024
				usagePercent := s.calculateUsagePercent(avgBytes, memLimit)
				oomRisk := "low"
				if usagePercent > 80 {
					oomRisk = "medium"
				}
				if usagePercent > 90 {
					oomRisk = "high"
				}
				metrics.Memory = &MemoryMetrics{
					AvgMB:        avgMB,
					MaxMB:        maxMB,
					MinMB:        minMB,
					LimitMB:      memLimitMB,
					UsagePercent: usagePercent,
					OOMRisk:      oomRisk,
					Trend:        trend,
				}
				break
			}
		}
	}

	memRssQuery := fmt.Sprintf(`sum(container_memory_rss{namespace="%s",pod="%s",container!="",container!="POD"})`, namespace, podName)
	rssResult, err := s.QueryRange(cluster.PrometheusUrl, memRssQuery, fmt.Sprintf("%d", start), fmt.Sprintf("%d", end), step, cluster.PrometheusAuthEnabled, cluster.PrometheusBasicAuthUser, cluster.PrometheusBasicAuthPass)
	if err == nil && rssResult.Status == "success" && len(rssResult.Data.Result) > 0 {
		values := s.extractAllValues(rssResult)
		if len(values) > 0 && metrics.Memory != nil {
			avg, _, _, _ := s.calculateStats(values)
			metrics.Memory.RssMB = avg / 1024 / 1024
		}
	}

	memCacheQuery := fmt.Sprintf(`sum(container_memory_cache{namespace="%s",pod="%s",container!="",container!="POD"})`, namespace, podName)
	cacheResult, err := s.QueryRange(cluster.PrometheusUrl, memCacheQuery, fmt.Sprintf("%d", start), fmt.Sprintf("%d", end), step, cluster.PrometheusAuthEnabled, cluster.PrometheusBasicAuthUser, cluster.PrometheusBasicAuthPass)
	if err == nil && cacheResult.Status == "success" && len(cacheResult.Data.Result) > 0 {
		values := s.extractAllValues(cacheResult)
		if len(values) > 0 && metrics.Memory != nil {
			avg, _, _, _ := s.calculateStats(values)
			metrics.Memory.CacheMB = avg / 1024 / 1024
		}
	}

	memPerContainerQuery := fmt.Sprintf(`container_memory_working_set_bytes{namespace="%s",pod="%s",container!="",container!="POD"}`, namespace, podName)
	memContainerResult, err := s.QueryRange(cluster.PrometheusUrl, memPerContainerQuery, fmt.Sprintf("%d", start), fmt.Sprintf("%d", end), step, cluster.PrometheusAuthEnabled, cluster.PrometheusBasicAuthUser, cluster.PrometheusBasicAuthPass)
	if err == nil && memContainerResult.Status == "success" && len(memContainerResult.Data.Result) > 0 {
		containerMemData := s.extractContainerValues(memContainerResult)
		for name, values := range containerMemData {
			if len(values) > 0 {
				avg, max, _, _ := s.calculateStats(values)
				for i, cm := range containerMetrics {
					if cm.Name == name {
						containerMetrics[i].MemoryAvgMB = avg / 1024 / 1024
						containerMetrics[i].MemoryMaxMB = max / 1024 / 1024
						break
					}
				}
			}
		}
	}

	txQuery := fmt.Sprintf(`sum(rate(container_network_transmit_bytes_total{namespace="%s",pod="%s"}[%s]))*8/1024`, namespace, podName, rateWindow)
	txResult, err := s.QueryRange(cluster.PrometheusUrl, txQuery, fmt.Sprintf("%d", start), fmt.Sprintf("%d", end), step, cluster.PrometheusAuthEnabled, cluster.PrometheusBasicAuthUser, cluster.PrometheusBasicAuthPass)
	if err == nil && txResult.Status == "success" && len(txResult.Data.Result) > 0 {
		values := s.extractAllValues(txResult)
		if len(values) > 0 {
			avg, max, _, _ := s.calculateStats(values)
			if metrics.Network == nil {
				metrics.Network = &NetworkMetrics{}
			}
			metrics.Network.TxAvgKbS = avg
			metrics.Network.TxMaxKbS = max
		}
	}

	rxQuery := fmt.Sprintf(`sum(rate(container_network_receive_bytes_total{namespace="%s",pod="%s"}[%s]))*8/1024`, namespace, podName, rateWindow)
	rxResult, err := s.QueryRange(cluster.PrometheusUrl, rxQuery, fmt.Sprintf("%d", start), fmt.Sprintf("%d", end), step, cluster.PrometheusAuthEnabled, cluster.PrometheusBasicAuthUser, cluster.PrometheusBasicAuthPass)
	if err == nil && rxResult.Status == "success" && len(rxResult.Data.Result) > 0 {
		values := s.extractAllValues(rxResult)
		if len(values) > 0 {
			avg, max, _, _ := s.calculateStats(values)
			if metrics.Network == nil {
				metrics.Network = &NetworkMetrics{}
			}
			metrics.Network.RxAvgKbS = avg
			metrics.Network.RxMaxKbS = max
		}
	}

	diskReadQuery := fmt.Sprintf(`sum(rate(container_fs_reads_bytes_total{namespace="%s",pod="%s"}[%s]))/1024`, namespace, podName, rateWindow)
	readResult, err := s.QueryRange(cluster.PrometheusUrl, diskReadQuery, fmt.Sprintf("%d", start), fmt.Sprintf("%d", end), step, cluster.PrometheusAuthEnabled, cluster.PrometheusBasicAuthUser, cluster.PrometheusBasicAuthPass)
	if err == nil && readResult.Status == "success" && len(readResult.Data.Result) > 0 {
		values := s.extractAllValues(readResult)
		if len(values) > 0 {
			avg, _, _, _ := s.calculateStats(values)
			if metrics.Disk == nil {
				metrics.Disk = &DiskMetrics{}
			}
			metrics.Disk.ReadAvgKbS = avg
		}
	}

	diskWriteQuery := fmt.Sprintf(`sum(rate(container_fs_writes_bytes_total{namespace="%s",pod="%s"}[%s]))/1024`, namespace, podName, rateWindow)
	writeResult, err := s.QueryRange(cluster.PrometheusUrl, diskWriteQuery, fmt.Sprintf("%d", start), fmt.Sprintf("%d", end), step, cluster.PrometheusAuthEnabled, cluster.PrometheusBasicAuthUser, cluster.PrometheusBasicAuthPass)
	if err == nil && writeResult.Status == "success" && len(writeResult.Data.Result) > 0 {
		values := s.extractAllValues(writeResult)
		if len(values) > 0 {
			avg, _, _, _ := s.calculateStats(values)
			if metrics.Disk == nil {
				metrics.Disk = &DiskMetrics{}
			}
			metrics.Disk.WriteAvgKbS = avg
		}
	}

	kubePodPhaseQuery := fmt.Sprintf(`kube_pod_status_phase{namespace="%s",pod="%s"}`, namespace, podName)
	phaseResult, err := s.QueryInstant(cluster.PrometheusUrl, kubePodPhaseQuery, cluster.PrometheusAuthEnabled, cluster.PrometheusBasicAuthUser, cluster.PrometheusBasicAuthPass)
	if err == nil && phaseResult.Status == "success" && len(phaseResult.Data.Result) > 0 {
		metrics.Pod = &PodKubeMetrics{}
		for _, r := range phaseResult.Data.Result {
			if phase, ok := r.Metric["phase"].(string); ok {
				if len(r.Value) >= 2 {
					if val, ok := r.Value[1].(string); ok && val == "1" {
						metrics.Pod.StatusPhase = phase
					}
				}
			}
		}
	}

	kubeContainerWaitingQuery := fmt.Sprintf(`kube_pod_container_status_waiting_reason{namespace="%s",pod="%s"}`, namespace, podName)
	waitingResult, err := s.QueryInstant(cluster.PrometheusUrl, kubeContainerWaitingQuery, cluster.PrometheusAuthEnabled, cluster.PrometheusBasicAuthUser, cluster.PrometheusBasicAuthPass)
	if err == nil && waitingResult.Status == "success" && len(waitingResult.Data.Result) > 0 {
		if metrics.Pod == nil {
			metrics.Pod = &PodKubeMetrics{}
		}
		metrics.Pod.ContainerWaiting = true
		for _, r := range waitingResult.Data.Result {
			if reason, ok := r.Metric["reason"].(string); ok {
				if len(r.Value) >= 2 {
					if val, ok := r.Value[1].(string); ok && val == "1" {
						metrics.Pod.WaitingReason = reason
					}
				}
			}
		}
	}

	kubeContainerTerminatedQuery := fmt.Sprintf(`kube_pod_container_status_terminated_reason{namespace="%s",pod="%s"}`, namespace, podName)
	terminatedResult, err := s.QueryInstant(cluster.PrometheusUrl, kubeContainerTerminatedQuery, cluster.PrometheusAuthEnabled, cluster.PrometheusBasicAuthUser, cluster.PrometheusBasicAuthPass)
	if err == nil && terminatedResult.Status == "success" && len(terminatedResult.Data.Result) > 0 {
		if metrics.Pod == nil {
			metrics.Pod = &PodKubeMetrics{}
		}
		metrics.Pod.ContainerTerminated = true
		for _, r := range terminatedResult.Data.Result {
			if reason, ok := r.Metric["reason"].(string); ok {
				if len(r.Value) >= 2 {
					if val, ok := r.Value[1].(string); ok && val == "1" {
						metrics.Pod.TerminatedReason = reason
					}
				}
			}
		}
	}

	kubeContainerRestartsQuery := fmt.Sprintf(`kube_pod_container_status_restarts_total{namespace="%s",pod="%s"}`, namespace, podName)
	restartsResult, err := s.QueryInstant(cluster.PrometheusUrl, kubeContainerRestartsQuery, cluster.PrometheusAuthEnabled, cluster.PrometheusBasicAuthUser, cluster.PrometheusBasicAuthPass)
	if err == nil && restartsResult.Status == "success" && len(restartsResult.Data.Result) > 0 {
		if metrics.Pod == nil {
			metrics.Pod = &PodKubeMetrics{}
		}
		totalRestarts := 0
		for _, r := range restartsResult.Data.Result {
			if len(r.Value) >= 2 {
				if val, ok := r.Value[1].(string); ok {
					totalRestarts += int(parseFloat(val))
				}
			}
		}
		metrics.Pod.Restarts = totalRestarts
	}

	kubePodReadyQuery := fmt.Sprintf(`kube_pod_status_ready{namespace="%s",pod="%s",condition="true"}`, namespace, podName)
	readyResult, err := s.QueryInstant(cluster.PrometheusUrl, kubePodReadyQuery, cluster.PrometheusAuthEnabled, cluster.PrometheusBasicAuthUser, cluster.PrometheusBasicAuthPass)
	if err == nil && readyResult.Status == "success" && len(readyResult.Data.Result) > 0 {
		if metrics.Pod == nil {
			metrics.Pod = &PodKubeMetrics{}
		}
		for _, r := range readyResult.Data.Result {
			if len(r.Value) >= 2 {
				if val, ok := r.Value[1].(string); ok && val == "1" {
					metrics.Pod.Ready = true
				}
			}
		}
	}

	if len(containerMetrics) > 0 {
		metrics.Container = containerMetrics
	}

	return metrics, nil
}

func (s *PrometheusService) GetNodeMetrics(clusterID uint, nodeName string, durationMinutes int) (map[string]interface{}, error) {
	cluster, err := clusterService.GetClusterByID(clusterID)
	if err != nil {
		return nil, err
	}

	if cluster.PrometheusUrl == "" {
		return nil, nil
	}

	end := time.Now().Unix()
	start := end - int64(durationMinutes*60)
	step := "30s"

	result := map[string]interface{}{}

	cpuQueries := []string{
		fmt.Sprintf(`sum(rate(node_cpu_seconds_total{node="%s",mode!="idle"}[60s])) / count(node_cpu_seconds_total{node="%s",mode="idle"})`, nodeName, nodeName),
		fmt.Sprintf(`1 - avg(rate(node_cpu_seconds_total{node="%s",mode="idle"}[60s]))`, nodeName),
		fmt.Sprintf(`avg(rate(node_cpu_seconds_total{node="%s"}[60s])) by (node)`, nodeName),
	}

	for _, cpuQuery := range cpuQueries {
		cpuResult, err := s.QueryRange(cluster.PrometheusUrl, cpuQuery, fmt.Sprintf("%d", start), fmt.Sprintf("%d", end), step, cluster.PrometheusAuthEnabled, cluster.PrometheusBasicAuthUser, cluster.PrometheusBasicAuthPass)
		if err == nil && cpuResult.Status == "success" && len(cpuResult.Data.Result) > 0 {
			values := s.extractAllValues(cpuResult)
			if len(values) > 0 {
				avg, max, _, trend := s.calculateStats(values)
				result["cpu_usage"] = avg * 100
				result["cpu_max"] = max * 100
				result["cpu_trend"] = trend
				break
			}
		}
	}

	memQueries := []string{
		fmt.Sprintf(`(1 - (node_memory_MemAvailable_bytes{node="%s"} / node_memory_MemTotal_bytes{node="%s"})) * 100`, nodeName, nodeName),
		fmt.Sprintf(`(node_memory_MemTotal_bytes{node="%s"} - node_memory_MemFree_bytes{node="%s"} - node_memory_Buffers_bytes{node="%s"} - node_memory_Cached_bytes{node="%s"}) / node_memory_MemTotal_bytes{node="%s"} * 100`, nodeName, nodeName, nodeName, nodeName, nodeName),
	}

	for _, memQuery := range memQueries {
		memResult, err := s.QueryRange(cluster.PrometheusUrl, memQuery, fmt.Sprintf("%d", start), fmt.Sprintf("%d", end), step, cluster.PrometheusAuthEnabled, cluster.PrometheusBasicAuthUser, cluster.PrometheusBasicAuthPass)
		if err == nil && memResult.Status == "success" && len(memResult.Data.Result) > 0 {
			values := s.extractAllValues(memResult)
			if len(values) > 0 {
				avg, max, _, trend := s.calculateStats(values)
				result["memory_usage"] = avg
				result["memory_max"] = max
				result["memory_trend"] = trend
				break
			}
		}
	}

	memTotalQuery := fmt.Sprintf(`node_memory_MemTotal_bytes{node="%s"}`, nodeName)
	memTotalResult, err := s.QueryInstant(cluster.PrometheusUrl, memTotalQuery, cluster.PrometheusAuthEnabled, cluster.PrometheusBasicAuthUser, cluster.PrometheusBasicAuthPass)
	if err == nil && memTotalResult.Status == "success" && len(memTotalResult.Data.Result) > 0 {
		for _, r := range memTotalResult.Data.Result {
			if len(r.Value) >= 2 {
				if val, ok := r.Value[1].(string); ok {
					result["memory_total_gb"] = parseFloat(val) / 1024 / 1024 / 1024
				}
			}
		}
	}

	diskQuery := fmt.Sprintf(`avg((node_filesystem_size_bytes{node="%s",fstype!="tmpfs"} - node_filesystem_free_bytes{node="%s",fstype!="tmpfs"}) / node_filesystem_size_bytes{node="%s",fstype!="tmpfs"}) * 100`, nodeName, nodeName, nodeName)
	diskResult, err := s.QueryRange(cluster.PrometheusUrl, diskQuery, fmt.Sprintf("%d", start), fmt.Sprintf("%d", end), step, cluster.PrometheusAuthEnabled, cluster.PrometheusBasicAuthUser, cluster.PrometheusBasicAuthPass)
	if err == nil && diskResult.Status == "success" && len(diskResult.Data.Result) > 0 {
		values := s.extractAllValues(diskResult)
		if len(values) > 0 {
			avg, _, _, _ := s.calculateStats(values)
			result["disk_usage"] = avg
		}
	}

	kubeNodeReadyQuery := fmt.Sprintf(`kube_node_status_condition{node="%s",condition="Ready",status="true"}`, nodeName)
	nodeReadyResult, err := s.QueryInstant(cluster.PrometheusUrl, kubeNodeReadyQuery, cluster.PrometheusAuthEnabled, cluster.PrometheusBasicAuthUser, cluster.PrometheusBasicAuthPass)
	if err == nil && nodeReadyResult.Status == "success" && len(nodeReadyResult.Data.Result) > 0 {
		for _, r := range nodeReadyResult.Data.Result {
			if len(r.Value) >= 2 {
				if val, ok := r.Value[1].(string); ok && val == "1" {
					result["node_ready"] = true
				}
			}
		}
	}

	kubeNodeConditionsQuery := fmt.Sprintf(`kube_node_status_condition{node="%s"}`, nodeName)
	conditionsResult, err := s.QueryInstant(cluster.PrometheusUrl, kubeNodeConditionsQuery, cluster.PrometheusAuthEnabled, cluster.PrometheusBasicAuthUser, cluster.PrometheusBasicAuthPass)
	if err == nil && conditionsResult.Status == "success" && len(conditionsResult.Data.Result) > 0 {
		conditions := map[string]string{}
		for _, r := range conditionsResult.Data.Result {
			if condition, ok := r.Metric["condition"].(string); ok {
				if status, ok := r.Metric["status"].(string); ok {
					if len(r.Value) >= 2 {
						if val, ok := r.Value[1].(string); ok && val == "1" {
							conditions[condition] = status
						}
					}
				}
			}
		}
		result["node_conditions"] = conditions
	}

	podCountQuery := fmt.Sprintf(`count(kube_pod_info{node="%s"})`, nodeName)
	podCountResult, err := s.QueryInstant(cluster.PrometheusUrl, podCountQuery, cluster.PrometheusAuthEnabled, cluster.PrometheusBasicAuthUser, cluster.PrometheusBasicAuthPass)
	if err == nil && podCountResult.Status == "success" && len(podCountResult.Data.Result) > 0 {
		for _, r := range podCountResult.Data.Result {
			if len(r.Value) >= 2 {
				if val, ok := r.Value[1].(string); ok {
					result["pod_count"] = int(parseFloat(val))
				}
			}
		}
	}

	return result, nil
}

func (s *PrometheusService) GetClusterMetrics(clusterID uint, durationMinutes int) (map[string]interface{}, error) {
	cluster, err := clusterService.GetClusterByID(clusterID)
	if err != nil {
		return nil, err
	}

	if cluster.PrometheusUrl == "" {
		return nil, nil
	}

	end := time.Now().Unix()
	start := end - int64(durationMinutes*60)
	step := "30s"

	result := map[string]interface{}{}

	cpuQueries := []string{
		`sum(rate(container_cpu_usage_seconds_total{container!="",container!="POD"}[60s])) / sum(kube_node_status_capacity_cpu_cores) * 100`,
		`sum(rate(node_cpu_seconds_total{mode!="idle"}[60s])) / sum(kube_node_status_capacity_cpu_cores) * 100`,
	}

	for _, cpuQuery := range cpuQueries {
		cpuResult, err := s.QueryRange(cluster.PrometheusUrl, cpuQuery, fmt.Sprintf("%d", start), fmt.Sprintf("%d", end), step, cluster.PrometheusAuthEnabled, cluster.PrometheusBasicAuthUser, cluster.PrometheusBasicAuthPass)
		if err == nil && cpuResult.Status == "success" && len(cpuResult.Data.Result) > 0 {
			values := s.extractAllValues(cpuResult)
			if len(values) > 0 {
				avg, max, _, _ := s.calculateStats(values)
				result["cluster_cpu_usage"] = avg
				result["cluster_cpu_max"] = max
				break
			}
		}
	}

	memQueries := []string{
		`sum(container_memory_working_set_bytes{container!="",container!="POD"}) / sum(kube_node_status_capacity_memory_bytes) * 100`,
		`sum(node_memory_MemTotal_bytes - node_memory_MemAvailable_bytes) / sum(node_memory_MemTotal_bytes) * 100`,
	}

	for _, memQuery := range memQueries {
		memResult, err := s.QueryRange(cluster.PrometheusUrl, memQuery, fmt.Sprintf("%d", start), fmt.Sprintf("%d", end), step, cluster.PrometheusAuthEnabled, cluster.PrometheusBasicAuthUser, cluster.PrometheusBasicAuthPass)
		if err == nil && memResult.Status == "success" && len(memResult.Data.Result) > 0 {
			values := s.extractAllValues(memResult)
			if len(values) > 0 {
				avg, max, _, _ := s.calculateStats(values)
				result["cluster_memory_usage"] = avg
				result["cluster_memory_max"] = max
				break
			}
		}
	}

	nodeCountQuery := `count(kube_node_info)`
	nodeCountResult, err := s.QueryInstant(cluster.PrometheusUrl, nodeCountQuery, cluster.PrometheusAuthEnabled, cluster.PrometheusBasicAuthUser, cluster.PrometheusBasicAuthPass)
	if err == nil && nodeCountResult.Status == "success" && len(nodeCountResult.Data.Result) > 0 {
		for _, r := range nodeCountResult.Data.Result {
			if len(r.Value) >= 2 {
				if val, ok := r.Value[1].(string); ok {
					result["node_count"] = int(parseFloat(val))
				}
			}
		}
	}

	nodeReadyQuery := `count(kube_node_status_condition{condition="Ready",status="true"})`
	nodeReadyResult, err := s.QueryInstant(cluster.PrometheusUrl, nodeReadyQuery, cluster.PrometheusAuthEnabled, cluster.PrometheusBasicAuthUser, cluster.PrometheusBasicAuthPass)
	if err == nil && nodeReadyResult.Status == "success" && len(nodeReadyResult.Data.Result) > 0 {
		for _, r := range nodeReadyResult.Data.Result {
			if len(r.Value) >= 2 {
				if val, ok := r.Value[1].(string); ok {
					result["node_ready_count"] = int(parseFloat(val))
				}
			}
		}
	}

	podTotalQuery := `count(kube_pod_info)`
	podTotalResult, err := s.QueryInstant(cluster.PrometheusUrl, podTotalQuery, cluster.PrometheusAuthEnabled, cluster.PrometheusBasicAuthUser, cluster.PrometheusBasicAuthPass)
	if err == nil && podTotalResult.Status == "success" && len(podTotalResult.Data.Result) > 0 {
		for _, r := range podTotalResult.Data.Result {
			if len(r.Value) >= 2 {
				if val, ok := r.Value[1].(string); ok {
					result["pod_total"] = int(parseFloat(val))
				}
			}
		}
	}

	podRunningQuery := `count(kube_pod_status_phase{phase="Running"})`
	podRunningResult, err := s.QueryInstant(cluster.PrometheusUrl, podRunningQuery, cluster.PrometheusAuthEnabled, cluster.PrometheusBasicAuthUser, cluster.PrometheusBasicAuthPass)
	if err == nil && podRunningResult.Status == "success" && len(podRunningResult.Data.Result) > 0 {
		for _, r := range podRunningResult.Data.Result {
			if len(r.Value) >= 2 {
				if val, ok := r.Value[1].(string); ok {
					result["pod_running"] = int(parseFloat(val))
				}
			}
		}
	}

	podPendingQuery := `count(kube_pod_status_phase{phase="Pending"})`
	podPendingResult, err := s.QueryInstant(cluster.PrometheusUrl, podPendingQuery, cluster.PrometheusAuthEnabled, cluster.PrometheusBasicAuthUser, cluster.PrometheusBasicAuthPass)
	if err == nil && podPendingResult.Status == "success" && len(podPendingResult.Data.Result) > 0 {
		for _, r := range podPendingResult.Data.Result {
			if len(r.Value) >= 2 {
				if val, ok := r.Value[1].(string); ok {
					result["pod_pending"] = int(parseFloat(val))
				}
			}
		}
	}

	podFailedQuery := `count(kube_pod_status_phase{phase="Failed"})`
	podFailedResult, err := s.QueryInstant(cluster.PrometheusUrl, podFailedQuery, cluster.PrometheusAuthEnabled, cluster.PrometheusBasicAuthUser, cluster.PrometheusBasicAuthPass)
	if err == nil && podFailedResult.Status == "success" && len(podFailedResult.Data.Result) > 0 {
		for _, r := range podFailedResult.Data.Result {
			if len(r.Value) >= 2 {
				if val, ok := r.Value[1].(string); ok {
					result["pod_failed"] = int(parseFloat(val))
				}
			}
		}
	}

	containerRestartQuery := `sum(kube_pod_container_status_restarts_total)`
	restartResult, err := s.QueryInstant(cluster.PrometheusUrl, containerRestartQuery, cluster.PrometheusAuthEnabled, cluster.PrometheusBasicAuthUser, cluster.PrometheusBasicAuthPass)
	if err == nil && restartResult.Status == "success" && len(restartResult.Data.Result) > 0 {
		for _, r := range restartResult.Data.Result {
			if len(r.Value) >= 2 {
				if val, ok := r.Value[1].(string); ok {
					result["container_restarts"] = int(parseFloat(val))
				}
			}
		}
	}

	deploymentUnavailableQuery := `sum(kube_deployment_status_replicas_unavailable)`
	unavailableResult, err := s.QueryInstant(cluster.PrometheusUrl, deploymentUnavailableQuery, cluster.PrometheusAuthEnabled, cluster.PrometheusBasicAuthUser, cluster.PrometheusBasicAuthPass)
	if err == nil && unavailableResult.Status == "success" && len(unavailableResult.Data.Result) > 0 {
		for _, r := range unavailableResult.Data.Result {
			if len(r.Value) >= 2 {
				if val, ok := r.Value[1].(string); ok {
					result["deployment_unavailable"] = int(parseFloat(val))
				}
			}
		}
	}

	nodeMetricsList := []map[string]interface{}{}
	nodeListQuery := `kube_node_info`
	nodeListResult, err := s.QueryInstant(cluster.PrometheusUrl, nodeListQuery, cluster.PrometheusAuthEnabled, cluster.PrometheusBasicAuthUser, cluster.PrometheusBasicAuthPass)
	if err == nil && nodeListResult.Status == "success" && len(nodeListResult.Data.Result) > 0 {
		for _, r := range nodeListResult.Data.Result {
			if node, ok := r.Metric["node"].(string); ok {
				nodeMetric := map[string]interface{}{"name": node}
				nodeMetrics, _ := s.GetNodeMetrics(clusterID, node, durationMinutes)
				if nodeMetrics != nil {
					for k, v := range nodeMetrics {
						nodeMetric[k] = v
					}
				}
				nodeMetricsList = append(nodeMetricsList, nodeMetric)
			}
		}
	}
	result["node_metrics"] = nodeMetricsList

	return result, nil
}

func (s *PrometheusService) ExecuteCustomQuery(clusterID uint, query string, durationMinutes int) (map[string]interface{}, error) {
	cluster, err := clusterService.GetClusterByID(clusterID)
	if err != nil {
		return nil, err
	}

	if cluster.PrometheusUrl == "" {
		return nil, fmt.Errorf("集群未配置Prometheus地址")
	}

	end := time.Now().Unix()
	start := end - int64(durationMinutes*60)
	step := "30s"

	result := map[string]interface{}{}

	rangeResult, err := s.QueryRange(cluster.PrometheusUrl, query, fmt.Sprintf("%d", start), fmt.Sprintf("%d", end), step, cluster.PrometheusAuthEnabled, cluster.PrometheusBasicAuthUser, cluster.PrometheusBasicAuthPass)
	if err != nil {
		return nil, err
	}

	if rangeResult.Status == "success" && len(rangeResult.Data.Result) > 0 {
		result["raw_data"] = s.parsePrometheusResult(rangeResult)
		values := s.extractAllValues(rangeResult)
		if len(values) > 0 {
			avg, max, min, trend := s.calculateStats(values)
			result["avg"] = avg
			result["max"] = max
			result["min"] = min
			result["trend"] = trend
			result["data_points"] = len(values)
		}
	}

	return result, nil
}

func (s *PrometheusService) extractAllValues(result *PrometheusQueryResult) []float64 {
	values := []float64{}
	for _, r := range result.Data.Result {
		for _, v := range r.Values {
			if len(v) >= 2 {
				if strVal, ok := v[1].(string); ok && strVal != "NaN" {
					values = append(values, parseFloat(strVal))
				}
			}
		}
	}
	return values
}

func (s *PrometheusService) extractContainerValues(result *PrometheusQueryResult) map[string][]float64 {
	containerData := map[string][]float64{}
	for _, r := range result.Data.Result {
		containerName := r.Metric["container"]
		if containerName == "" {
			containerName = r.Metric["name"]
		}
		if containerName == "" {
			continue
		}
		values := []float64{}
		for _, v := range r.Values {
			if len(v) >= 2 {
				if strVal, ok := v[1].(string); ok && strVal != "NaN" {
					values = append(values, parseFloat(strVal))
				}
			}
		}
		containerData[containerName] = values
	}
	return containerData
}

func (s *PrometheusService) calculateStats(values []float64) (avg, max, min float64, trend string) {
	if len(values) == 0 {
		return 0, 0, 0, "unknown"
	}

	sum := 0.0
	max = values[0]
	min = values[0]
	for _, v := range values {
		sum += v
		if v > max {
			max = v
		}
		if v < min {
			min = v
		}
	}
	avg = sum / float64(len(values))

	if len(values) >= 10 {
		firstHalf := values[:len(values)/2]
		secondHalf := values[len(values)/2:]
		firstAvg := 0.0
		secondAvg := 0.0
		for _, v := range firstHalf {
			firstAvg += v
		}
		for _, v := range secondHalf {
			secondAvg += v
		}
		firstAvg /= float64(len(firstHalf))
		secondAvg /= float64(len(secondHalf))

		if firstAvg == 0 {
			trend = "stable"
			return avg, max, min, trend
		}

		diff := (secondAvg - firstAvg) / firstAvg * 100
		if diff > 10 {
			trend = "rising"
		} else if diff < -10 {
			trend = "falling"
		} else {
			trend = "stable"
		}
	} else {
		trend = "stable"
	}

	return avg, max, min, trend
}

func (s *PrometheusService) calculateUsagePercent(used, limit float64) float64 {
	if limit <= 0 {
		return 0
	}
	return used / limit * 100
}

func (s *PrometheusService) parsePrometheusResult(result *PrometheusQueryResult) []map[string]interface{} {
	data := make([]map[string]interface{}, 0)

	for _, r := range result.Data.Result {
		metricInfo := map[string]interface{}{}
		for k, v := range r.Metric {
			metricInfo[k] = v
		}

		values := make([]map[string]interface{}, 0)
		for _, v := range r.Values {
			if len(v) >= 2 {
				timestamp := int64(v[0].(float64))
				value := 0.0
				if strVal, ok := v[1].(string); ok {
					if strVal != "NaN" {
						value = parseFloat(strVal)
					}
				}
				values = append(values, map[string]interface{}{
					"time":  timestamp,
					"value": value,
				})
			}
		}

		data = append(data, map[string]interface{}{
			"metric": metricInfo,
			"values": values,
		})
	}

	return data
}

func parseFloat(s string) float64 {
	s = strings.TrimSpace(s)
	var f float64
	var neg bool

	if len(s) > 0 && s[0] == '-' {
		neg = true
		s = s[1:]
	}

	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= '0' && c <= '9' {
			f = f*10 + float64(c-'0')
		} else if c == '.' {
			dec := 0.0
			pos := 0.1
			for j := i + 1; j < len(s); j++ {
				if s[j] >= '0' && s[j] <= '9' {
					dec += float64(s[j]-'0') * pos
					pos *= 0.1
				} else {
					break
				}
			}
			f += dec
			break
		} else {
			break
		}
	}

	if neg {
		f = -f
	}
	return f
}

func (s *PrometheusService) GetPodMetrics(clusterID uint, namespace, podName string, durationMinutes int) (map[string]interface{}, error) {
	cluster, err := clusterService.GetClusterByID(clusterID)
	if err != nil {
		return nil, err
	}

	if cluster.PrometheusUrl == "" {
		return nil, fmt.Errorf("集群未配置Prometheus地址，请在集群管理中配置")
	}

	end := time.Now().Unix()
	start := end - int64(durationMinutes*60)
	step := "15s"
	rateWindow := "60s"

	global.GVA_LOG.Info("查询Prometheus指标",
		zap.String("url", cluster.PrometheusUrl),
		zap.String("namespace", namespace),
		zap.String("pod", podName),
		zap.Int("duration", durationMinutes))

	return s.queryMetrics(cluster, namespace, podName, start, end, step, rateWindow)
}

func (s *PrometheusService) GetPodMetricsWithParams(clusterID uint, namespace, podName string, durationMinutes int, stepSeconds int, customStart, customEnd int64) (map[string]interface{}, error) {
	cluster, err := clusterService.GetClusterByID(clusterID)
	if err != nil {
		return nil, err
	}

	if cluster.PrometheusUrl == "" {
		return nil, fmt.Errorf("集群未配置Prometheus地址，请在集群管理中配置")
	}

	var start, end int64
	if customStart > 0 && customEnd > 0 {
		start = customStart
		end = customEnd
	} else {
		end = time.Now().Unix()
		start = end - int64(durationMinutes*60)
	}

	step := fmt.Sprintf("%ds", stepSeconds)
	rateWindow := stepSeconds * 3
	if rateWindow < 60 {
		rateWindow = 60
	}
	rateWindowStr := fmt.Sprintf("%ds", rateWindow)

	return s.queryMetrics(cluster, namespace, podName, start, end, step, rateWindowStr)
}

func (s *PrometheusService) queryMetrics(cluster *model.K8sCluster, namespace, podName string, start, end int64, step string, rateWindow string) (map[string]interface{}, error) {
	metrics := map[string]interface{}{}

	cpuQueries := []string{
		fmt.Sprintf(`rate(container_cpu_usage_seconds_total{namespace="%s",pod="%s",container!="",container!="POD"}[%s])`, namespace, podName, rateWindow),
		fmt.Sprintf(`irate(container_cpu_usage_seconds_total{namespace="%s",pod="%s",container!="",container!="POD"}[30s])`, namespace, podName),
	}
	for _, cpuQuery := range cpuQueries {
		cpuResult, err := s.QueryRange(cluster.PrometheusUrl, cpuQuery, fmt.Sprintf("%d", start), fmt.Sprintf("%d", end), step, cluster.PrometheusAuthEnabled, cluster.PrometheusBasicAuthUser, cluster.PrometheusBasicAuthPass)
		if err == nil && cpuResult.Status == "success" && len(cpuResult.Data.Result) > 0 {
			metrics["cpu"] = s.parsePrometheusResult(cpuResult)
			break
		}
	}

	memQueries := []string{
		fmt.Sprintf(`container_memory_working_set_bytes{namespace="%s",pod="%s",container!="",container!="POD"}`, namespace, podName),
		fmt.Sprintf(`container_memory_usage_bytes{namespace="%s",pod="%s",container!="",container!="POD"}`, namespace, podName),
	}
	for _, memQuery := range memQueries {
		memResult, err := s.QueryRange(cluster.PrometheusUrl, memQuery, fmt.Sprintf("%d", start), fmt.Sprintf("%d", end), step, cluster.PrometheusAuthEnabled, cluster.PrometheusBasicAuthUser, cluster.PrometheusBasicAuthPass)
		if err == nil && memResult.Status == "success" && len(memResult.Data.Result) > 0 {
			metrics["memory"] = s.parsePrometheusResult(memResult)
			break
		}
	}

	txQueries := []string{
		fmt.Sprintf(`rate(container_network_transmit_bytes_total{namespace="%s",pod="%s"}[%s])*8`, namespace, podName, rateWindow),
		fmt.Sprintf(`irate(container_network_transmit_bytes_total{namespace="%s",pod="%s"}[30s])*8`, namespace, podName),
	}
	for _, txQuery := range txQueries {
		txResult, err := s.QueryRange(cluster.PrometheusUrl, txQuery, fmt.Sprintf("%d", start), fmt.Sprintf("%d", end), step, cluster.PrometheusAuthEnabled, cluster.PrometheusBasicAuthUser, cluster.PrometheusBasicAuthPass)
		if err == nil && txResult.Status == "success" && len(txResult.Data.Result) > 0 {
			metrics["network_tx"] = s.parsePrometheusResult(txResult)
			break
		}
	}

	rxQueries := []string{
		fmt.Sprintf(`rate(container_network_receive_bytes_total{namespace="%s",pod="%s"}[%s])*8`, namespace, podName, rateWindow),
		fmt.Sprintf(`irate(container_network_receive_bytes_total{namespace="%s",pod="%s"}[30s])*8`, namespace, podName),
	}
	for _, rxQuery := range rxQueries {
		rxResult, err := s.QueryRange(cluster.PrometheusUrl, rxQuery, fmt.Sprintf("%d", start), fmt.Sprintf("%d", end), step, cluster.PrometheusAuthEnabled, cluster.PrometheusBasicAuthUser, cluster.PrometheusBasicAuthPass)
		if err == nil && rxResult.Status == "success" && len(rxResult.Data.Result) > 0 {
			metrics["network_rx"] = s.parsePrometheusResult(rxResult)
			break
		}
	}

	return metrics, nil
}
