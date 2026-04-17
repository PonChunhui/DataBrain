package v1

import (
	"net/http"
	"strconv"

	"devops-backend/model/request"
	"devops-backend/model/response"
	"devops-backend/service"
	"github.com/gin-gonic/gin"
)

type MenuApi struct{}

var menuService = &service.MenuService{}

func (api *MenuApi) GetMenuList(c *gin.Context) {
	menus, err := menuService.GetMenuList()
	if err != nil {
		c.JSON(http.StatusOK, response.FailWithMsg(nil, "获取菜单列表失败"))
		return
	}

	c.JSON(http.StatusOK, response.Success(menus))
}

func (api *MenuApi) GetMenuTree(c *gin.Context) {
	menus, err := menuService.GetMenuTree()
	if err != nil {
		c.JSON(http.StatusOK, response.FailWithMsg(nil, "获取菜单树失败"))
		return
	}

	c.JSON(http.StatusOK, response.Success(menus))
}

func (api *MenuApi) CreateMenu(c *gin.Context) {
	var req request.MenuRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, response.FailWithMsg(nil, "参数错误"))
		return
	}

	if err := menuService.CreateMenu(req); err != nil {
		c.JSON(http.StatusOK, response.FailWithMsg(nil, "创建菜单失败"))
		return
	}

	c.JSON(http.StatusOK, response.Success(nil))
}

func (api *MenuApi) UpdateMenu(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var req request.MenuRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, response.FailWithMsg(nil, "参数错误"))
		return
	}

	if err := menuService.UpdateMenu(uint(id), req); err != nil {
		c.JSON(http.StatusOK, response.FailWithMsg(nil, "更新菜单失败"))
		return
	}

	c.JSON(http.StatusOK, response.Success(nil))
}

func (api *MenuApi) DeleteMenu(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := menuService.DeleteMenu(uint(id)); err != nil {
		c.JSON(http.StatusOK, response.FailWithMsg(nil, "删除菜单失败"))
		return
	}

	c.JSON(http.StatusOK, response.Success(nil))
}
