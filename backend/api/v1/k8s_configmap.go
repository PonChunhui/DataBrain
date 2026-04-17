package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"devops-backend/model/response"
	"devops-backend/service"
)

type ConfigMapController struct{}

var k8sConfigMapService = &service.K8sConfigMap{}

func (c *ConfigMapController) GetConfigMaps(ctx *gin.Context) {
	clusterID, err := resolveClusterID(ctx.Query("cluster"))
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(err.Error()))
		return
	}

	namespace := ctx.Query("namespace")
	if namespace == "" {
		namespace = "default"
	}

	configMaps, err := k8sConfigMapService.GetConfigMaps(uint(clusterID), namespace)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, response.Success(configMaps))
}

func (c *ConfigMapController) GetConfigMap(ctx *gin.Context) {
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

	cm, err := k8sConfigMapService.GetConfigMap(uint(clusterID), namespace, name)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, response.Success(cm))
}

func (c *ConfigMapController) CreateConfigMap(ctx *gin.Context) {
	clusterID, err := resolveClusterID(ctx.Query("cluster"))
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(err.Error()))
		return
	}

	namespace := ctx.Query("namespace")
	if namespace == "" {
		namespace = "default"
	}

	var req struct {
		Name   string            `json:"name" binding:"required"`
		Data   map[string]string `json:"data"`
		Labels map[string]string `json:"labels"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(err.Error()))
		return
	}

	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:   req.Name,
			Labels: req.Labels,
		},
		Data: req.Data,
	}

	if err := k8sConfigMapService.CreateConfigMap(uint(clusterID), namespace, cm); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, response.Success(nil))
}

func (c *ConfigMapController) UpdateConfigMap(ctx *gin.Context) {
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
		Data   map[string]string `json:"data"`
		Labels map[string]string `json:"labels"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(err.Error()))
		return
	}

	existingCM, err := k8sConfigMapService.GetConfigMap(uint(clusterID), namespace, name)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(err.Error()))
		return
	}

	existingCM.Data = req.Data
	if req.Labels != nil {
		existingCM.Labels = req.Labels
	}

	if err := k8sConfigMapService.UpdateConfigMap(uint(clusterID), namespace, existingCM); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, response.Success(nil))
}

func (c *ConfigMapController) DeleteConfigMap(ctx *gin.Context) {
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

	if err := k8sConfigMapService.DeleteConfigMap(uint(clusterID), namespace, name); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, response.Success(nil))
}
