package v1

import (
	"net/http"
	"strconv"

	"devops-backend/model/request"
	"devops-backend/model/response"
	"devops-backend/service"
	"github.com/gin-gonic/gin"
)

type UserApi struct{}

var userService = &service.UserService{}

func (api *UserApi) Login(c *gin.Context) {
	var req request.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, response.FailWithMsg(nil, "参数错误"))
		return
	}

	token, user, err := userService.Login(req)
	if err != nil {
		c.JSON(http.StatusOK, response.FailWithMsg(nil, err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(response.LoginResponse{
		Token: token,
		User:  user,
	}))
}

func (api *UserApi) GetUserList(c *gin.Context) {
	users, err := userService.GetUserList()
	if err != nil {
		c.JSON(http.StatusOK, response.FailWithMsg(nil, "获取用户列表失败"))
		return
	}

	c.JSON(http.StatusOK, response.Success(users))
}

func (api *UserApi) CreateUser(c *gin.Context) {
	var req request.UserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, response.FailWithMsg(nil, "参数错误"))
		return
	}

	if err := userService.CreateUser(req); err != nil {
		c.JSON(http.StatusOK, response.FailWithMsg(nil, "创建用户失败"))
		return
	}

	c.JSON(http.StatusOK, response.Success(nil))
}

func (api *UserApi) UpdateUser(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var req request.UserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, response.FailWithMsg(nil, "参数错误"))
		return
	}

	if err := userService.UpdateUser(uint(id), req); err != nil {
		c.JSON(http.StatusOK, response.FailWithMsg(nil, "更新用户失败"))
		return
	}

	c.JSON(http.StatusOK, response.Success(nil))
}

func (api *UserApi) DeleteUser(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := userService.DeleteUser(uint(id)); err != nil {
		c.JSON(http.StatusOK, response.FailWithMsg(nil, "删除用户失败"))
		return
	}

	c.JSON(http.StatusOK, response.Success(nil))
}

func (api *UserApi) AssignRole(c *gin.Context) {
	userID, _ := strconv.Atoi(c.Param("id"))
	var req request.AssignRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, response.FailWithMsg(nil, "参数错误"))
		return
	}

	if err := userService.AssignRole(uint(userID), req.RoleID); err != nil {
		c.JSON(http.StatusOK, response.FailWithMsg(nil, "分配角色失败"))
		return
	}

	c.JSON(http.StatusOK, response.Success(nil))
}

func (api *UserApi) AssignRoles(c *gin.Context) {
	userID, _ := strconv.Atoi(c.Param("id"))
	var req request.AssignRolesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, response.FailWithMsg(nil, "参数错误"))
		return
	}

	if err := userService.AssignRoles(uint(userID), req.RoleIDs); err != nil {
		c.JSON(http.StatusOK, response.FailWithMsg(nil, "分配角色失败"))
		return
	}

	c.JSON(http.StatusOK, response.Success(nil))
}

func (api *UserApi) GetUserByID(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	user, err := userService.GetUserByID(uint(id))
	if err != nil {
		c.JSON(http.StatusOK, response.FailWithMsg(nil, "获取用户失败"))
		return
	}
	c.JSON(http.StatusOK, response.Success(user))
}

func (api *UserApi) GetUserRoles(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	roles, err := userService.GetUserRoles(uint(id))
	if err != nil {
		c.JSON(http.StatusOK, response.FailWithMsg(nil, "获取用户角色失败"))
		return
	}
	c.JSON(http.StatusOK, response.Success(roles))
}

func (api *UserApi) GetUserMenus(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	menus, err := userService.GetUserMenus(uint(id))
	if err != nil {
		c.JSON(http.StatusOK, response.FailWithMsg(nil, "获取用户菜单失败"))
		return
	}
	c.JSON(http.StatusOK, response.Success(menus))
}

func (api *UserApi) GetUserApis(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	apis, err := userService.GetUserApis(uint(id))
	if err != nil {
		c.JSON(http.StatusOK, response.FailWithMsg(nil, "获取用户API失败"))
		return
	}
	c.JSON(http.StatusOK, response.Success(apis))
}

func (api *UserApi) ChangePassword(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var req request.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, response.FailWithMsg(nil, "参数错误"))
		return
	}

	if err := userService.ChangePassword(uint(id), req.OldPassword, req.NewPassword); err != nil {
		c.JSON(http.StatusOK, response.FailWithMsg(nil, err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(nil))
}
