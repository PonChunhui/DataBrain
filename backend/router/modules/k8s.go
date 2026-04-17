package modules

import (
	v1 "devops-backend/api/v1"
	"devops-backend/middleware"
	"github.com/gin-gonic/gin"
)

func InitK8sRouter(ApiGroup *gin.RouterGroup) {
	// 集群路由 - 需要对集群资源的权限
	clusterRouter := ApiGroup.Group("k8s/cluster")
	clusterController := &v1.ClusterController{}
	{
		clusterRouter.GET("", middleware.K8sAuthMiddleware("cluster", "view"), clusterController.GetClusters)
		clusterRouter.GET("/:id", middleware.K8sAuthMiddleware("cluster", "view"), clusterController.GetCluster)
		clusterRouter.GET("/:id/info", middleware.K8sAuthMiddleware("cluster", "view"), clusterController.GetClusterInfo)
		clusterRouter.POST("", middleware.K8sAuthMiddleware("cluster", "create"), clusterController.CreateCluster)
		clusterRouter.PUT("/:id", middleware.K8sAuthMiddleware("cluster", "edit"), clusterController.UpdateCluster)
		clusterRouter.DELETE("/:id", middleware.K8sAuthMiddleware("cluster", "delete"), clusterController.DeleteCluster)
		clusterRouter.GET("/:id/namespaces", middleware.K8sAuthMiddleware("cluster", "view"), clusterController.GetNamespaces)
		clusterRouter.POST("/test", middleware.K8sAuthMiddleware("cluster", "view"), clusterController.TestKubeconfig)
	}

	// Deployment路由
	deploymentRouter := ApiGroup.Group("k8s/deployment")
	deploymentController := &v1.DeploymentController{}
	deploymentDetailController := &v1.DeploymentDetailController{}
	{
		deploymentRouter.GET("", middleware.K8sAuthMiddleware("deployment", "view"), deploymentController.GetDeployments)
		deploymentRouter.GET("/:name/detail", middleware.K8sAuthMiddleware("deployment", "view"), deploymentDetailController.GetDeploymentDetail)
		deploymentRouter.GET("/:name/pods", middleware.K8sAuthMiddleware("deployment", "view"), deploymentDetailController.GetDeploymentPods)
		deploymentRouter.GET("/:name/events", middleware.K8sAuthMiddleware("deployment", "view"), deploymentDetailController.GetDeploymentEvents)
		deploymentRouter.GET("/:name/revisions", middleware.K8sAuthMiddleware("deployment", "view"), deploymentDetailController.GetDeploymentRevisions)
		deploymentRouter.GET("/:name/diff", middleware.K8sAuthMiddleware("deployment", "view"), deploymentDetailController.CompareRevisions)
		deploymentRouter.GET("/:name/yaml", middleware.K8sAuthMiddleware("deployment", "view"), deploymentController.GetDeploymentYAML)
		deploymentRouter.PUT("/:name/yaml", middleware.K8sAuthMiddleware("deployment", "edit"), deploymentController.UpdateDeploymentYAML)
		deploymentRouter.POST("/:name/rollback/:revision", middleware.K8sAuthMiddleware("deployment", "edit"), deploymentDetailController.RollbackToRevision)
		deploymentRouter.POST("/:name/scale", middleware.K8sAuthMiddleware("deployment", "edit"), deploymentController.ScaleDeployment)
		deploymentRouter.POST("/:name/restart", middleware.K8sAuthMiddleware("deployment", "edit"), deploymentController.RestartDeployment)
		deploymentRouter.DELETE("/:name", middleware.K8sAuthMiddleware("deployment", "delete"), deploymentController.DeleteDeployment)
		deploymentRouter.GET("/:name", middleware.K8sAuthMiddleware("deployment", "view"), deploymentController.GetDeployment)
		deploymentRouter.POST("", middleware.K8sAuthMiddleware("deployment", "create"), deploymentController.CreateDeployment)
	}

	// Pod路由
	podRouter := ApiGroup.Group("k8s/pod")
	podController := &v1.PodController{}
	prometheusController := &v1.PrometheusController{}
	{
		podRouter.GET("", middleware.K8sAuthMiddleware("pod", "view"), podController.GetPods)
		podRouter.GET("/:name/detail", middleware.K8sAuthMiddleware("pod", "view"), podController.GetPodDetail)
		podRouter.GET("/:name/prometheus", middleware.K8sAuthMiddleware("pod", "view"), prometheusController.GetPodMetrics)
		podRouter.GET("/:name/metrics", middleware.K8sAuthMiddleware("pod", "view"), podController.GetPodMetrics)
		podRouter.GET("/:name/logs", middleware.K8sAuthMiddleware("pod", "view"), podController.GetPodLogs)
		podRouter.GET("/:name/events", middleware.K8sAuthMiddleware("pod", "view"), podController.GetPodEvents)
		podRouter.GET("/:name", middleware.K8sAuthMiddleware("pod", "view"), podController.GetPod)
		podRouter.DELETE("/:name", middleware.K8sAuthMiddleware("pod", "delete"), podController.DeletePod)
	}

	// Pod日志WebSocket路由
	podLogRouter := ApiGroup.Group("k8s/pod/log")
	podLogController := &v1.PodLogController{}
	{
		podLogRouter.GET("/stream", middleware.K8sAuthMiddleware("pod", "view"), podLogController.StreamLogs)
		podLogRouter.POST("/stop", middleware.K8sAuthMiddleware("pod", "view"), podLogController.StopStream)
	}

	// Pod终端WebSocket路由
	podExecRouter := ApiGroup.Group("k8s/pod/exec")
	podExecController := &v1.PodExecController{}
	{
		podExecRouter.GET("/stream", middleware.K8sAuthMiddleware("pod", "edit"), podExecController.ExecPod)
		podExecRouter.POST("/stop", middleware.K8sAuthMiddleware("pod", "edit"), podExecController.StopExec)
	}

	// Service路由
	serviceRouter := ApiGroup.Group("k8s/service")
	serviceController := &v1.ServiceController{}
	{
		serviceRouter.GET("", middleware.K8sAuthMiddleware("service", "view"), serviceController.GetServices)
		serviceRouter.GET("/:name", middleware.K8sAuthMiddleware("service", "view"), serviceController.GetService)
		serviceRouter.GET("/:name/detail", middleware.K8sAuthMiddleware("service", "view"), serviceController.GetServiceDetail)
		serviceRouter.GET("/:name/pods", middleware.K8sAuthMiddleware("service", "view"), serviceController.GetServicePods)
		serviceRouter.GET("/:name/events", middleware.K8sAuthMiddleware("service", "view"), serviceController.GetServiceEvents)
		serviceRouter.GET("/:name/deployments", middleware.K8sAuthMiddleware("service", "view"), serviceController.GetServiceDeployments)
		serviceRouter.GET("/:name/yaml", middleware.K8sAuthMiddleware("service", "view"), serviceController.GetServiceYAML)
		serviceRouter.PUT("/:name/yaml", middleware.K8sAuthMiddleware("service", "edit"), serviceController.UpdateServiceYAML)
		serviceRouter.DELETE("/:name", middleware.K8sAuthMiddleware("service", "delete"), serviceController.DeleteService)
		serviceRouter.POST("", middleware.K8sAuthMiddleware("service", "create"), serviceController.CreateService)
	}

	// ConfigMap路由
	configMapRouter := ApiGroup.Group("k8s/configmap")
	configMapController := &v1.ConfigMapController{}
	{
		configMapRouter.GET("", middleware.K8sAuthMiddleware("configmap", "view"), configMapController.GetConfigMaps)
		configMapRouter.GET("/:name", middleware.K8sAuthMiddleware("configmap", "view"), configMapController.GetConfigMap)
		configMapRouter.POST("", middleware.K8sAuthMiddleware("configmap", "create"), configMapController.CreateConfigMap)
		configMapRouter.PUT("/:name", middleware.K8sAuthMiddleware("configmap", "edit"), configMapController.UpdateConfigMap)
		configMapRouter.DELETE("/:name", middleware.K8sAuthMiddleware("configmap", "delete"), configMapController.DeleteConfigMap)
	}

	// Secret路由
	secretRouter := ApiGroup.Group("k8s/secret")
	secretController := &v1.SecretController{}
	{
		secretRouter.GET("", middleware.K8sAuthMiddleware("secret", "view"), secretController.GetSecrets)
		secretRouter.GET("/:name", middleware.K8sAuthMiddleware("secret", "view"), secretController.GetSecret)
		secretRouter.POST("", middleware.K8sAuthMiddleware("secret", "create"), secretController.CreateSecret)
		secretRouter.PUT("/:name", middleware.K8sAuthMiddleware("secret", "edit"), secretController.UpdateSecret)
		secretRouter.DELETE("/:name", middleware.K8sAuthMiddleware("secret", "delete"), secretController.DeleteSecret)
	}

	// Ingress路由
	ingressRouter := ApiGroup.Group("k8s/ingress")
	ingressController := &v1.IngressController{}
	{
		ingressRouter.GET("", middleware.K8sAuthMiddleware("ingress", "view"), ingressController.GetIngresses)
		ingressRouter.GET("/:name/yaml", middleware.K8sAuthMiddleware("ingress", "view"), ingressController.GetIngressYAML)
		ingressRouter.PUT("/:name/yaml", middleware.K8sAuthMiddleware("ingress", "edit"), ingressController.UpdateIngressYAML)
		ingressRouter.GET("/:name", middleware.K8sAuthMiddleware("ingress", "view"), ingressController.GetIngress)
		ingressRouter.GET("/:name/events", middleware.K8sAuthMiddleware("ingress", "view"), ingressController.GetIngressEvents)
		ingressRouter.POST("", middleware.K8sAuthMiddleware("ingress", "create"), ingressController.CreateIngress)
		ingressRouter.PUT("/:name", middleware.K8sAuthMiddleware("ingress", "edit"), ingressController.UpdateIngress)
		ingressRouter.DELETE("/:name", middleware.K8sAuthMiddleware("ingress", "delete"), ingressController.DeleteIngress)
	}

	// 集群统计路由
	statsRouter := ApiGroup.Group("k8s/stats")
	statsController := &v1.StatsController{}
	{
		statsRouter.GET("/:id", middleware.K8sAuthMiddleware("cluster", "view"), statsController.GetClusterStats)
	}

	// Node路由
	nodeRouter := ApiGroup.Group("k8s/node")
	nodeController := &v1.NodeController{}
	{
		nodeRouter.GET("/:name/detail", middleware.K8sAuthMiddleware("node", "view"), nodeController.GetNodeDetail)
		nodeRouter.GET("/:name/pods", middleware.K8sAuthMiddleware("node", "view"), nodeController.GetNodePods)
		nodeRouter.GET("/:name/events", middleware.K8sAuthMiddleware("node", "view"), nodeController.GetNodeEvents)
	}
}
