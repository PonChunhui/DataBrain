package middleware

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"devops-backend/global"
	"devops-backend/model"
	"devops-backend/model/response"
	"devops-backend/service"
)

var k8sAuthService = &service.K8sAuthService{}

func K8sAuthMiddleware(resource, action string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		claims, exists := ctx.Get("claims")
		if !exists {
			ctx.JSON(http.StatusOK, response.FailWithMsg(nil, "未登录，请先登录"))
			ctx.Abort()
			return
		}

		customClaims, ok := claims.(*CustomClaims)
		if !ok {
			ctx.JSON(http.StatusOK, response.FailWithMsg(nil, "无效的认证信息"))
			ctx.Abort()
			return
		}

		uid := customClaims.UserID

		if k8sAuthService.IsSuperAdmin(uid) {
			ctx.Next()
			return
		}

		clusterID := getClusterIDFromContext(ctx)
		namespace := ctx.Query("namespace")
		if namespace == "" {
			namespace = ctx.Param("namespace")
		}

		if !k8sAuthService.CheckPermission(uid, clusterID, namespace, resource, action) {
			actionText := getActionText(action)
			resourceText := getResourceText(resource)
			clusterName := getClusterName(clusterID)
			nsText := namespace
			if namespace == "" {
				nsText = "默认命名空间"
			}
			msg := fmt.Sprintf("无权访问: 您没有集群[%s]命名空间[%s]的%s%s权限，请联系管理员授权", clusterName, nsText, resourceText, actionText)
			ctx.JSON(http.StatusOK, response.FailWithMsg(nil, msg))
			ctx.Abort()
			return
		}

		ctx.Next()
	}
}

func getActionText(action string) string {
	actionMap := map[string]string{
		"view":   "查看",
		"edit":   "编辑",
		"delete": "删除",
		"create": "创建",
	}
	if text, ok := actionMap[action]; ok {
		return text
	}
	return action
}

func getResourceText(resource string) string {
	resourceMap := map[string]string{
		"cluster":    "集群",
		"deployment": "Deployment",
		"pod":        "Pod",
		"service":    "Service",
		"configmap":  "ConfigMap",
		"secret":     "Secret",
		"ingress":    "Ingress",
		"node":       "Node",
	}
	if text, ok := resourceMap[resource]; ok {
		return text
	}
	return resource
}

func getClusterName(clusterID uint) string {
	if clusterID == 0 {
		return "所有集群"
	}
	var cluster model.K8sCluster
	if err := global.GVA_DB.First(&cluster, clusterID).Error; err != nil {
		return "未知集群"
	}
	return cluster.Name
}

func getClusterIDFromContext(ctx *gin.Context) uint {
	clusterAlias := ctx.Query("cluster")
	if clusterAlias == "" {
		clusterAlias = ctx.Param("cluster")
	}

	if clusterAlias == "" {
		return 0
	}

	var cluster model.K8sCluster
	if err := global.GVA_DB.Where("alias = ?", clusterAlias).First(&cluster).Error; err != nil {
		return 0
	}
	return cluster.ID
}
