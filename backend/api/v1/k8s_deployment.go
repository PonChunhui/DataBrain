package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"

	appsv1 "k8s.io/api/apps/v1"

	"devops-backend/model/response"
	"devops-backend/service"
)

type DeploymentController struct{}

var deploymentService = &service.DeploymentService{}

func (c *DeploymentController) GetDeployments(ctx *gin.Context) {
	clusterID, err := resolveClusterID(ctx.Query("cluster"))
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(err.Error()))
		return
	}

	namespace := ctx.Query("namespace")
	if namespace == "" {
		namespace = "default"
	}

	deployments, err := deploymentService.GetDeployments(uint(clusterID), namespace)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, response.Success(deployments))
}

func (c *DeploymentController) GetDeployment(ctx *gin.Context) {
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

	deployment, err := deploymentService.GetDeployment(uint(clusterID), namespace, name)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, response.Success(deployment))
}

func (c *DeploymentController) ScaleDeployment(ctx *gin.Context) {
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

	var req struct {
		Replicas int32 `json:"replicas"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(err.Error()))
		return
	}

	if err := deploymentService.ScaleDeployment(uint(clusterID), namespace, name, req.Replicas); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, response.Success(nil))
}

func (c *DeploymentController) RestartDeployment(ctx *gin.Context) {
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

	if err := deploymentService.RestartDeployment(uint(clusterID), namespace, name); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, response.Success(nil))
}

func (c *DeploymentController) DeleteDeployment(ctx *gin.Context) {
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

	if err := deploymentService.DeleteDeployment(uint(clusterID), namespace, name); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, response.Success(nil))
}

func (c *DeploymentController) CreateDeployment(ctx *gin.Context) {
	clusterID, err := resolveClusterID(ctx.Query("cluster"))
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(err.Error()))
		return
	}

	namespace := ctx.Query("namespace")
	if namespace == "" {
		namespace = "default"
	}

	var deployment appsv1.Deployment
	if err := ctx.ShouldBindJSON(&deployment); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(err.Error()))
		return
	}

	if err := deploymentService.CreateDeployment(uint(clusterID), namespace, &deployment); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, response.Success(nil))
}

func (c *DeploymentController) GetDeploymentYAML(ctx *gin.Context) {
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

	yaml, err := deploymentService.GetDeploymentYAML(uint(clusterID), namespace, name)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, response.Success(yaml))
}

func (c *DeploymentController) UpdateDeploymentYAML(ctx *gin.Context) {
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

	var req struct {
		YAML string `json:"yaml" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail("YAML内容不能为空"))
		return
	}

	if err := deploymentService.UpdateDeploymentYAML(uint(clusterID), namespace, name, req.YAML); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, response.Success(nil))
}
