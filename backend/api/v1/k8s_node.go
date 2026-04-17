package v1

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"devops-backend/model/response"
	"devops-backend/service"
)

type NodeController struct{}

var k8sNodeService = &service.K8sNodeService{}

func (c *NodeController) GetNodeDetail(ctx *gin.Context) {
	clusterID, err := resolveClusterID(ctx.Query("cluster"))
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(err.Error()))
		return
	}

	nodeName := ctx.Param("name")

	detail, err := k8sNodeService.GetNodeDetail(uint(clusterID), nodeName)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, response.Success(detail))
}

func (c *NodeController) GetNodePods(ctx *gin.Context) {
	clusterID, err := resolveClusterID(ctx.Query("cluster"))
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(err.Error()))
		return
	}

	nodeName := ctx.Param("name")
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(ctx.DefaultQuery("pageSize", "10"))

	pods, total, err := k8sNodeService.GetNodePods(uint(clusterID), nodeName, page, pageSize)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, response.Success(map[string]interface{}{
		"list":     pods,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	}))
}

func (c *NodeController) GetNodeEvents(ctx *gin.Context) {
	clusterID, err := resolveClusterID(ctx.Query("cluster"))
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(err.Error()))
		return
	}

	nodeName := ctx.Param("name")

	events, err := k8sNodeService.GetNodeEvents(uint(clusterID), nodeName)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, response.Success(events))
}
