package v1

import (
	"net/http"
	"strconv"

	"devops-backend/model/request"
	"devops-backend/model/response"
	"devops-backend/service"
	"github.com/gin-gonic/gin"
)

type ApiApi struct{}

var apiService = &service.ApiService{}

func (api *ApiApi) GetApiList(c *gin.Context) {
	var req request.ApiSearchRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusOK, response.FailWithMsg(nil, "参数错误"))
		return
	}

	apis, total, err := apiService.GetApiList(req)
	if err != nil {
		c.JSON(http.StatusOK, response.FailWithMsg(nil, "获取API列表失败"))
		return
	}

	c.JSON(http.StatusOK, response.Success(response.PageResult{
		List:     apis,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}))
}

func (api *ApiApi) CreateApi(c *gin.Context) {
	var req request.ApiRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, response.FailWithMsg(nil, "参数错误"))
		return
	}

	if err := apiService.CreateApi(req); err != nil {
		c.JSON(http.StatusOK, response.FailWithMsg(nil, "创建API失败"))
		return
	}

	c.JSON(http.StatusOK, response.Success(nil))
}

func (api *ApiApi) UpdateApi(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var req request.ApiRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, response.FailWithMsg(nil, "参数错误"))
		return
	}

	if err := apiService.UpdateApi(uint(id), req); err != nil {
		c.JSON(http.StatusOK, response.FailWithMsg(nil, "更新API失败"))
		return
	}

	c.JSON(http.StatusOK, response.Success(nil))
}

func (api *ApiApi) DeleteApi(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := apiService.DeleteApi(uint(id)); err != nil {
		c.JSON(http.StatusOK, response.FailWithMsg(nil, "删除API失败"))
		return
	}

	c.JSON(http.StatusOK, response.Success(nil))
}
