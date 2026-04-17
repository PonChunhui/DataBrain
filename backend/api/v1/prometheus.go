package v1

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"devops-backend/model/response"
	"devops-backend/service"
)

type PrometheusController struct{}

var prometheusService = &service.PrometheusService{}

func (c *PrometheusController) GetPodMetrics(ctx *gin.Context) {
	clusterID, err := resolveClusterID(ctx.Query("cluster"))
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(err.Error()))
		return
	}

	namespace := ctx.Query("namespace")
	if namespace == "" {
		namespace = "default"
	}

	podName := ctx.Param("name")

	step := 600
	if s := ctx.Query("step"); s != "" {
		if val, err := strconv.Atoi(s); err == nil && val > 0 {
			step = val
		}
	}

	var start, end int64
	var duration int

	if startStr := ctx.Query("start"); startStr != "" {
		if startVal, err := strconv.ParseInt(startStr, 10, 64); err == nil {
			start = startVal
			if endStr := ctx.Query("end"); endStr != "" {
				if endVal, err := strconv.ParseInt(endStr, 10, 64); err == nil {
					end = endVal
				}
			}
			if end == 0 {
				end = time.Now().Unix()
			}
			duration = int((end - start) / 60)
		}
	} else {
		duration = 480
		if d := ctx.Query("duration"); d != "" {
			if val, err := strconv.Atoi(d); err == nil && val > 0 {
				duration = val
			}
		}
	}

	metrics, err := prometheusService.GetPodMetricsWithParams(uint(clusterID), namespace, podName, duration, step, start, end)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(err.Error()))
		return
	}

	if len(metrics) == 0 {
		ctx.JSON(http.StatusOK, response.Fail("未获取到Prometheus指标数据，请检查：1. 集群是否配置了Prometheus地址；2. Prometheus是否正常采集Pod指标；3. Pod是否有对应容器"))
		return
	}

	ctx.JSON(http.StatusOK, response.Success(metrics))
}
