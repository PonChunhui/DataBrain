package v1

import (
	"net/http"
	"strconv"

	"devops-backend/model/request"
	"devops-backend/model/response"
	"devops-backend/service"
	"github.com/gin-gonic/gin"
)

type AuditApi struct{}

var auditService = &service.AuditService{}

func (api *AuditApi) GetAuditList(c *gin.Context) {
	var req request.AuditSearchRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusOK, response.FailWithMsg(nil, "参数错误"))
		return
	}

	result, err := auditService.GetAuditList(req)
	if err != nil {
		c.JSON(http.StatusOK, response.FailWithMsg(nil, "获取审计日志列表失败"))
		return
	}

	c.JSON(http.StatusOK, response.Success(result))
}

func (api *AuditApi) GetAuditByID(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	auditLog, err := auditService.GetAuditByID(uint(id))
	if err != nil {
		c.JSON(http.StatusOK, response.FailWithMsg(nil, "获取审计日志详情失败"))
		return
	}

	c.JSON(http.StatusOK, response.Success(auditLog))
}
