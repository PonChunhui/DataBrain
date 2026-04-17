package v1

import (
	"net/http"
	"strconv"

	"devops-backend/global"
	"devops-backend/middleware"
	"devops-backend/model/response"
	"devops-backend/service"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type DeploymentDetailController struct{}

var deploymentDetailService = &service.DeploymentDetailService{}

func (c *DeploymentDetailController) GetDeploymentDetail(ctx *gin.Context) {
	clusterID, err := resolveClusterID(ctx.Query("cluster"))
	namespace := ctx.Query("namespace")
	name := ctx.Param("name")

	detail, err := deploymentDetailService.GetDeploymentDetail(uint(clusterID), namespace, name)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(err.Error()))
		return
		global.GVA_LOG.Error("获取Deployment详情失败", zap.Error(err))
		ctx.JSON(http.StatusOK, response.FailWithMsg(nil, "获取失败"))
		return
	}

	ctx.JSON(http.StatusOK, response.Success(detail))
}

func (c *DeploymentDetailController) GetDeploymentPods(ctx *gin.Context) {
	clusterID, err := resolveClusterID(ctx.Query("cluster"))
	namespace := ctx.Query("namespace")
	name := ctx.Param("name")

	pods, err := deploymentDetailService.GetDeploymentPods(uint(clusterID), namespace, name)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(err.Error()))
		return
		global.GVA_LOG.Error("获取Deployment Pod列表失败", zap.Error(err))
		ctx.JSON(http.StatusOK, response.FailWithMsg(nil, "获取失败"))
		return
	}

	ctx.JSON(http.StatusOK, response.Success(pods))
}

func (c *DeploymentDetailController) GetDeploymentEvents(ctx *gin.Context) {
	clusterID, err := resolveClusterID(ctx.Query("cluster"))
	namespace := ctx.Query("namespace")
	name := ctx.Param("name")

	events, err := deploymentDetailService.GetDeploymentEvents(uint(clusterID), namespace, name)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(err.Error()))
		return
		global.GVA_LOG.Error("获取Deployment事件失败", zap.Error(err))
		ctx.JSON(http.StatusOK, response.FailWithMsg(nil, "获取失败"))
		return
	}

	ctx.JSON(http.StatusOK, response.Success(events))
}

func (c *DeploymentDetailController) GetDeploymentRevisions(ctx *gin.Context) {
	clusterID, err := resolveClusterID(ctx.Query("cluster"))
	namespace := ctx.Query("namespace")
	name := ctx.Param("name")

	revisions, err := deploymentDetailService.GetDeploymentRevisions(uint(clusterID), namespace, name)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(err.Error()))
		return
		global.GVA_LOG.Error("获取Deployment版本历史失败", zap.Error(err))
		ctx.JSON(http.StatusOK, response.FailWithMsg(nil, "获取失败"))
		return
	}

	ctx.JSON(http.StatusOK, response.Success(revisions))
}

func (c *DeploymentDetailController) CompareRevisions(ctx *gin.Context) {
	clusterID, err := resolveClusterID(ctx.Query("cluster"))
	namespace := ctx.Query("namespace")
	name := ctx.Param("name")
	rev1, _ := strconv.Atoi(ctx.Query("revision1"))
	rev2, _ := strconv.Atoi(ctx.Query("revision2"))

	revisions, err := deploymentDetailService.GetDeploymentRevisions(uint(clusterID), namespace, name)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(err.Error()))
		return
		ctx.JSON(http.StatusOK, response.FailWithMsg(nil, "获取版本失败"))
		return
	}

	var yaml1, yaml2 string
	for _, rev := range revisions {
		if rev.Revision == rev1 {
			yaml1 = rev.YAML
		}
		if rev.Revision == rev2 {
			yaml2 = rev.YAML
		}
	}

	if yaml1 == "" || yaml2 == "" {
		ctx.JSON(http.StatusOK, response.FailWithMsg(nil, "版本不存在"))
		return
	}

	if rev1 > rev2 {
		yaml1, yaml2 = yaml2, yaml1
	}

	diff, err := deploymentDetailService.CompareRevisions(yaml1, yaml2)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(err.Error()))
		return
		ctx.JSON(http.StatusOK, response.FailWithMsg(nil, "对比失败"))
		return
	}

	ctx.JSON(http.StatusOK, response.Success(diff))
}

func (c *DeploymentDetailController) RollbackToRevision(ctx *gin.Context) {
	clusterID, err := resolveClusterID(ctx.Query("cluster"))
	if err != nil {
		ctx.JSON(http.StatusOK, response.FailWithMsg(nil, err.Error()))
		return
	}
	namespace := ctx.Query("namespace")
	name := ctx.Param("name")
	revision, _ := strconv.Atoi(ctx.Param("revision"))

	claims, exists := ctx.Get("claims")
	changedBy := "system"
	if exists {
		if customClaims, ok := claims.(*middleware.CustomClaims); ok {
			changedBy = "user-" + strconv.Itoa(int(customClaims.UserID))
		}
	}

	global.GVA_LOG.Info("回退Deployment版本",
		zap.String("deployment", name),
		zap.Int("revision", revision),
		zap.String("by", changedBy))

	if err := deploymentDetailService.RollbackToRevision(uint(clusterID), namespace, name, revision); err != nil {
		ctx.JSON(http.StatusOK, response.FailWithMsg(nil, err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, response.Success(nil))
}
