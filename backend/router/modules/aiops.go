package modules

import (
	"github.com/gin-gonic/gin"

	"devops-backend/api/v1"
)

func InitAIOPSRouter(ApiGroup *gin.RouterGroup) {
	llmConfigController := &v1.AIOPSLLMConfigController{}
	diagnosticController := &v1.AIOPSDiagnosticController{}
	inspectionController := &v1.AIOPSClusterInspectionController{}

	aiopsRouter := ApiGroup.Group("aiops")
	{
		llmRouter := aiopsRouter.Group("llm-config")
		{
			llmRouter.GET("", llmConfigController.GetConfigs)
			llmRouter.GET("/:id", llmConfigController.GetConfig)
			llmRouter.POST("", llmConfigController.CreateConfig)
			llmRouter.PUT("/:id", llmConfigController.UpdateConfig)
			llmRouter.DELETE("/:id", llmConfigController.DeleteConfig)
			llmRouter.POST("/:id/test", llmConfigController.TestConfig)
			llmRouter.POST("/:id/set-default", llmConfigController.SetDefault)
		}

		diagnosticRouter := aiopsRouter.Group("diagnostic")
		{
			diagnosticRouter.POST("", diagnosticController.Diagnose)
			diagnosticRouter.POST("/stream", diagnosticController.StreamDiagnose)
			diagnosticRouter.POST("/chat", diagnosticController.StreamChat)
			diagnosticRouter.GET("/:id", diagnosticController.GetDiagnostic)
		}

		inspectionRouter := aiopsRouter.Group("inspection")
		{
			inspectionRouter.POST("/stream", inspectionController.StreamInspection)
			inspectionRouter.POST("/chat", inspectionController.StreamInspectionChat)
		}

		historyRouter := aiopsRouter.Group("history")
		{
			historyRouter.GET("", diagnosticController.GetHistory)
			historyRouter.GET("/stats", diagnosticController.GetStats)
		}
	}
}
