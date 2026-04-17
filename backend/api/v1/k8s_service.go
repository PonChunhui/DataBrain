package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"

	corev1 "k8s.io/api/core/v1"

	"devops-backend/model/response"
	"devops-backend/service"
)

type ServiceController struct{}

var k8sServiceService = &service.K8sService{}

func (c *ServiceController) GetServices(ctx *gin.Context) {
	clusterID, err := resolveClusterID(ctx.Query("cluster"))
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(err.Error()))
		return
	}

	namespace := ctx.Query("namespace")
	if namespace == "" {
		namespace = "default"
	}

	services, err := k8sServiceService.GetServices(uint(clusterID), namespace)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, response.Success(services))
}

func (c *ServiceController) GetService(ctx *gin.Context) {
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

	svc, err := k8sServiceService.GetService(uint(clusterID), namespace, name)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, response.Success(svc))
}

func (c *ServiceController) DeleteService(ctx *gin.Context) {
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

	if err := k8sServiceService.DeleteService(uint(clusterID), namespace, name); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, response.Success(nil))
}

func (c *ServiceController) CreateService(ctx *gin.Context) {
	clusterID, err := resolveClusterID(ctx.Query("cluster"))
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(err.Error()))
		return
	}

	namespace := ctx.Query("namespace")
	if namespace == "" {
		namespace = "default"
	}

	var svc corev1.Service
	if err := ctx.ShouldBindJSON(&svc); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(err.Error()))
		return
	}

	if err := k8sServiceService.CreateService(uint(clusterID), namespace, &svc); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, response.Success(nil))
}

func (c *ServiceController) GetServiceDetail(ctx *gin.Context) {
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

	detail, err := k8sServiceService.GetServiceDetail(uint(clusterID), namespace, name)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, response.Success(detail))
}

func (c *ServiceController) GetServicePods(ctx *gin.Context) {
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

	svc, err := k8sServiceService.GetService(uint(clusterID), namespace, name)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(err.Error()))
		return
	}

	pods, err := k8sServiceService.GetServicePods(uint(clusterID), namespace, svc.Spec.Selector)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, response.Success(pods))
}

func (c *ServiceController) GetServiceEvents(ctx *gin.Context) {
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

	events, err := k8sServiceService.GetServiceEvents(uint(clusterID), namespace, name)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, response.Success(events))
}

func (c *ServiceController) GetServiceDeployments(ctx *gin.Context) {
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

	svc, err := k8sServiceService.GetService(uint(clusterID), namespace, name)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(err.Error()))
		return
	}

	deployments, err := k8sServiceService.GetServiceDeployments(uint(clusterID), namespace, svc.Spec.Selector)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, response.Success(deployments))
}

func (c *ServiceController) GetServiceYAML(ctx *gin.Context) {
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

	yamlStr, err := k8sServiceService.GetServiceYAML(uint(clusterID), namespace, name)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, response.Success(yamlStr))
}

func (c *ServiceController) UpdateServiceYAML(ctx *gin.Context) {
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

	var req map[string]string
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(err.Error()))
		return
	}

	yamlStr := req["yaml"]
	if yamlStr == "" {
		ctx.JSON(http.StatusOK, response.Fail("YAML内容不能为空"))
		return
	}

	if err := k8sServiceService.UpdateServiceYAML(uint(clusterID), namespace, name, yamlStr); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, response.Success(nil))
}
