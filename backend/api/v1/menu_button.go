package v1

import (
	"net/http"
	"strconv"

	"devops-backend/model/request"
	"devops-backend/model/response"
	"devops-backend/service"
	"github.com/gin-gonic/gin"
)

type MenuButtonApi struct{}

var menuButtonService = &service.MenuButtonService{}

func (api *MenuButtonApi) GetMenuButtonList(c *gin.Context) {
	buttons, err := menuButtonService.GetMenuButtonList()
	if err != nil {
		c.JSON(http.StatusOK, response.FailWithMsg(nil, "获取菜单按钮列表失败"))
		return
	}

	c.JSON(http.StatusOK, response.Success(buttons))
}

func (api *MenuButtonApi) GetButtonsByMenu(c *gin.Context) {
	menuID, _ := strconv.Atoi(c.Param("menu_id"))
	buttons, err := menuButtonService.GetMenuButtons(uint(menuID))
	if err != nil {
		c.JSON(http.StatusOK, response.FailWithMsg(nil, "获取菜单按钮失败"))
		return
	}

	c.JSON(http.StatusOK, response.Success(buttons))
}

func (api *MenuButtonApi) CreateMenuButton(c *gin.Context) {
	var req request.MenuButtonRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, response.FailWithMsg(nil, "参数错误"))
		return
	}

	if err := menuButtonService.CreateMenuButton(req); err != nil {
		c.JSON(http.StatusOK, response.FailWithMsg(nil, "创建菜单按钮失败"))
		return
	}

	c.JSON(http.StatusOK, response.Success(nil))
}

func (api *MenuButtonApi) UpdateMenuButton(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var req request.MenuButtonRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, response.FailWithMsg(nil, "参数错误"))
		return
	}

	if err := menuButtonService.UpdateMenuButton(uint(id), req); err != nil {
		c.JSON(http.StatusOK, response.FailWithMsg(nil, "更新菜单按钮失败"))
		return
	}

	c.JSON(http.StatusOK, response.Success(nil))
}

func (api *MenuButtonApi) DeleteMenuButton(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := menuButtonService.DeleteMenuButton(uint(id)); err != nil {
		c.JSON(http.StatusOK, response.FailWithMsg(nil, "删除菜单按钮失败"))
		return
	}

	c.JSON(http.StatusOK, response.Success(nil))
}
