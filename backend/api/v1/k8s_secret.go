package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"devops-backend/model/response"
	"devops-backend/service"
)

type SecretController struct{}

var k8sSecretService = &service.K8sSecret{}

func (c *SecretController) GetSecrets(ctx *gin.Context) {
	clusterID, err := resolveClusterID(ctx.Query("cluster"))
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(err.Error()))
		return
	}

	namespace := ctx.Query("namespace")
	if namespace == "" {
		namespace = "default"
	}

	secrets, err := k8sSecretService.GetSecrets(uint(clusterID), namespace)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, response.Success(secrets))
}

func (c *SecretController) GetSecret(ctx *gin.Context) {
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

	secret, err := k8sSecretService.GetSecret(uint(clusterID), namespace, name)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, response.Success(secret))
}

func (c *SecretController) CreateSecret(ctx *gin.Context) {
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
		Type   string            `json:"type"`
		Data   map[string]string `json:"data"`
		Labels map[string]string `json:"labels"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(err.Error()))
		return
	}

	secretType := corev1.SecretTypeOpaque
	if req.Type != "" {
		secretType = corev1.SecretType(req.Type)
	}

	data := make(map[string][]byte)
	for k, v := range req.Data {
		data[k] = []byte(v)
	}

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:   req.Name,
			Labels: req.Labels,
		},
		Type: secretType,
		Data: data,
	}

	if err := k8sSecretService.CreateSecret(uint(clusterID), namespace, secret); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, response.Success(nil))
}

func (c *SecretController) UpdateSecret(ctx *gin.Context) {
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

	existingSecret, err := k8sSecretService.GetSecret(uint(clusterID), namespace, name)
	if err != nil {
		ctx.JSON(http.StatusOK, response.Fail(err.Error()))
		return
	}

	data := make(map[string][]byte)
	for k, v := range req.Data {
		data[k] = []byte(v)
	}

	existingSecret.Data = data
	if req.Labels != nil {
		existingSecret.Labels = req.Labels
	}

	if err := k8sSecretService.UpdateSecret(uint(clusterID), namespace, existingSecret); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, response.Success(nil))
}

func (c *SecretController) DeleteSecret(ctx *gin.Context) {
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

	if err := k8sSecretService.DeleteSecret(uint(clusterID), namespace, name); err != nil {
		ctx.JSON(http.StatusOK, response.Fail(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, response.Success(nil))
}
