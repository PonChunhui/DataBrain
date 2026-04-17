package v1

import (
	"net/http"
	"strconv"

	"devops-backend/model/request"
	"devops-backend/model/response"
	"devops-backend/service"
	"github.com/gin-gonic/gin"
)

type RoleApi struct{}

var roleService = &service.RoleService{}

func (api *RoleApi) GetRoleList(c *gin.Context) {
	roles, err := roleService.GetRoleList()
	if err != nil {
		c.JSON(http.StatusOK, response.FailWithMsg(nil, "获取角色列表失败"))
		return
	}

	c.JSON(http.StatusOK, response.Success(roles))
}

func (api *RoleApi) CreateRole(c *gin.Context) {
	var req request.RoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, response.FailWithMsg(nil, "参数错误"))
		return
	}

	if err := roleService.CreateRole(req); err != nil {
		c.JSON(http.StatusOK, response.FailWithMsg(nil, "创建角色失败"))
		return
	}

	c.JSON(http.StatusOK, response.Success(nil))
}

func (api *RoleApi) UpdateRole(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var req request.RoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, response.FailWithMsg(nil, "参数错误"))
		return
	}

	if err := roleService.UpdateRole(uint(id), req); err != nil {
		c.JSON(http.StatusOK, response.FailWithMsg(nil, "更新角色失败"))
		return
	}

	c.JSON(http.StatusOK, response.Success(nil))
}

func (api *RoleApi) DeleteRole(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := roleService.DeleteRole(uint(id)); err != nil {
		c.JSON(http.StatusOK, response.FailWithMsg(nil, "删除角色失败"))
		return
	}

	c.JSON(http.StatusOK, response.Success(nil))
}

func (api *RoleApi) AssignMenuButton(c *gin.Context) {
	roleID, _ := strconv.Atoi(c.Param("id"))
	var req request.AssignMenuButtonRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, response.FailWithMsg(nil, "参数错误"))
		return
	}

	if err := roleService.AssignMenuButton(uint(roleID), req.MenuButtonID); err != nil {
		c.JSON(http.StatusOK, response.FailWithMsg(nil, "分配菜单按钮失败"))
		return
	}

	c.JSON(http.StatusOK, response.Success(nil))
}

func (api *RoleApi) AssignMenus(c *gin.Context) {
	roleID, _ := strconv.Atoi(c.Param("id"))
	var req request.AssignMenusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, response.FailWithMsg(nil, "参数错误"))
		return
	}

	if err := roleService.AssignMenus(uint(roleID), req.MenuIDs); err != nil {
		c.JSON(http.StatusOK, response.FailWithMsg(nil, "分配菜单失败"))
		return
	}

	c.JSON(http.StatusOK, response.Success(nil))
}

func (api *RoleApi) GetRoleMenus(c *gin.Context) {
	roleID, _ := strconv.Atoi(c.Param("id"))
	menuIDs, err := roleService.GetRoleMenus(uint(roleID))
	if err != nil {
		c.JSON(http.StatusOK, response.FailWithMsg(nil, "获取角色菜单失败"))
		return
	}

	c.JSON(http.StatusOK, response.Success(menuIDs))
}

func (api *RoleApi) AssignApis(c *gin.Context) {
	roleID, _ := strconv.Atoi(c.Param("id"))
	var req request.AssignApisRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, response.FailWithMsg(nil, "参数错误"))
		return
	}

	if err := roleService.AssignApis(uint(roleID), req.ApiIDs); err != nil {
		c.JSON(http.StatusOK, response.FailWithMsg(nil, "分配API失败"))
		return
	}

	c.JSON(http.StatusOK, response.Success(nil))
}

func (api *RoleApi) GetRoleApis(c *gin.Context) {
	roleID, _ := strconv.Atoi(c.Param("id"))
	apiIDs, err := roleService.GetRoleApis(uint(roleID))
	if err != nil {
		c.JSON(http.StatusOK, response.FailWithMsg(nil, "获取角色API失败"))
		return
	}

	c.JSON(http.StatusOK, response.Success(apiIDs))
}
