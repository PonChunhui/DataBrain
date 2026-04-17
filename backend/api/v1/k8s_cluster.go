package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"devops-backend/model"
	"devops-backend/model/response"
	"devops-backend/service"
	"devops-backend/utils"
)

type ClusterController struct{}

var clusterService = &service.ClusterService{}

func (c *ClusterController) GetClusters(ctx *gin.Context) {
	alias := ctx.Query("alias")
	if alias != "" {
		cluster, err := clusterService.GetClusterByAlias(alias)
		if err != nil {
			ctx.JSON(http.StatusOK, response.Fail(err.Error()))
			return
		}
		ctx.JSON(http.StatusOK, response.Success([]*model.K8sCluster{cluster}))
		return
	}

	clusters, err := clusterService.GetClusters()
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, response.Success(clusters))
}

func (c *ClusterController) GetCluster(ctx *gin.Context) {
	clusterID, err := resolveClusterID(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(err.Error()))
		return
	}

	cluster, err := clusterService.GetClusterByID(clusterID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, response.Success(cluster))
}

func (c *ClusterController) GetClusterSecure(ctx *gin.Context) {
	clusterID, err := resolveClusterID(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(err.Error()))
		return
	}

	cluster, err := clusterService.GetClusterByIDSecure(clusterID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, response.Success(cluster))
}

func (c *ClusterController) CreateCluster(ctx *gin.Context) {
	var cluster model.K8sCluster
	if err := ctx.ShouldBindJSON(&cluster); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(err.Error()))
		return
	}

	if err := clusterService.CreateCluster(&cluster); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(err.Error()))
		return
	}

	result, err := clusterService.GetClusterByIDSecure(cluster.ID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Success(cluster))
	} else {
		ctx.JSON(http.StatusOK, response.Success(result))
	}
}

func (c *ClusterController) UpdateCluster(ctx *gin.Context) {
	clusterID, err := resolveClusterID(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(err.Error()))
		return
	}

	var cluster model.K8sCluster
	if err := ctx.ShouldBindJSON(&cluster); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(err.Error()))
		return
	}

	cluster.ID = clusterID

	if err := clusterService.UpdateCluster(&cluster); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(err.Error()))
		return
	}

	result, err := clusterService.GetClusterByIDSecure(cluster.ID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Success(cluster))
	} else {
		ctx.JSON(http.StatusOK, response.Success(result))
	}
}

func (c *ClusterController) DeleteCluster(ctx *gin.Context) {
	clusterID, err := resolveClusterID(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(err.Error()))
		return
	}

	if err := clusterService.DeleteCluster(clusterID); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, response.Success(nil))
}

func (c *ClusterController) GetNamespaces(ctx *gin.Context) {
	clusterID, err := resolveClusterID(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(err.Error()))
		return
	}

	namespaces, err := clusterService.GetNamespaces(clusterID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, response.Success(namespaces))
}

func (c *ClusterController) GetClusterInfo(ctx *gin.Context) {
	clusterID, err := resolveClusterID(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(err.Error()))
		return
	}

	info, err := clusterService.GetClusterInfo(clusterID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, response.Success(info))
}

func (c *ClusterController) TestKubeconfig(ctx *gin.Context) {
	var req struct {
		Kubeconfig string `json:"kubeconfig" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail("kubeconfig内容不能为空"))
		return
	}

	if err := utils.TestClusterConnectionFromKubeconfig(req.Kubeconfig); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(err.Error()))
		return
	}

	server, err := utils.GetClusterServer(req.Kubeconfig)
	if err != nil {
		server = "未知"
	}

	context, err := utils.GetCurrentContext(req.Kubeconfig)
	if err != nil {
		context = "未知"
	}

	ctx.JSON(http.StatusOK, response.Success(map[string]string{
		"server":  server,
		"context": context,
		"message": "kubeconfig连接测试成功",
	}))
}
