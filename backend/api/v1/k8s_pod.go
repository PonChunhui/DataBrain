package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"devops-backend/model/response"
	"devops-backend/service"
	"strconv"
)

type PodController struct{}

var podService = &service.PodService{}

func (c *PodController) GetPods(ctx *gin.Context) {
	clusterID, err := resolveClusterID(ctx.Query("cluster"))
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(err.Error()))
		return
	}

	namespace := ctx.Query("namespace")
	if namespace == "" {
		namespace = "default"
	}

	pods, err := podService.GetPods(uint(clusterID), namespace)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, response.Success(pods))
}

func (c *PodController) GetPod(ctx *gin.Context) {
	clusterID, err := resolveClusterID(ctx.Query("cluster"))
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(err.Error()))
		return
	}

	namespace := ctx.Query("namespace")
	if namespace == "" {
		namespace = "default"
	}

	name := ctx.Param("name")

	pod, err := podService.GetPod(uint(clusterID), namespace, name)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, response.Success(pod))
}

func (c *PodController) GetPodLogs(ctx *gin.Context) {
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
	containerName := ctx.Query("container")

	tailLines := int64(100)
	if tl := ctx.Query("tail_lines"); tl != "" {
		if val, err := strconv.ParseInt(tl, 10, 64); err == nil {
			tailLines = val
		}
	}

	logs, err := podService.GetPodLogs(uint(clusterID), namespace, podName, containerName, tailLines)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, response.Success(logs))
}

func (c *PodController) GetPodEvents(ctx *gin.Context) {
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

	events, err := podService.GetPodEvents(uint(clusterID), namespace, podName)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, response.Success(events))
}

func (c *PodController) DeletePod(ctx *gin.Context) {
	clusterID, err := resolveClusterID(ctx.Query("cluster"))
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(err.Error()))
		return
	}

	namespace := ctx.Query("namespace")
	if namespace == "" {
		namespace = "default"
	}

	name := ctx.Param("name")

	if err := podService.DeletePod(uint(clusterID), namespace, name); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, response.Success(nil))
}

func (c *PodController) GetPodDetail(ctx *gin.Context) {
	clusterID, err := resolveClusterID(ctx.Query("cluster"))
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(err.Error()))
		return
	}

	namespace := ctx.Query("namespace")
	if namespace == "" {
		namespace = "default"
	}

	name := ctx.Param("name")

	detail, err := podService.GetPodDetail(uint(clusterID), namespace, name)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, response.Success(detail))
}

func (c *PodController) GetPodMetrics(ctx *gin.Context) {
	clusterID, err := resolveClusterID(ctx.Query("cluster"))
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(err.Error()))
		return
	}

	namespace := ctx.Query("namespace")
	if namespace == "" {
		namespace = "default"
	}

	name := ctx.Param("name")

	metrics, err := podService.GetPodMetrics(uint(clusterID), namespace, name)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, response.Success(metrics))
}
