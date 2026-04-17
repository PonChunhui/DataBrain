package v1

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"devops-backend/middleware"
	"devops-backend/model"
	"devops-backend/model/request"
	"devops-backend/model/response"
	"devops-backend/service"
)

type K8sAuthController struct{}

var k8sAuthService = &service.K8sAuthService{}

func (c *K8sAuthController) GetAuthorizations(ctx *gin.Context) {
	var params request.K8sAuthListRequest
	if err := ctx.ShouldBindQuery(&params); err != nil {
		ctx.JSON(http.StatusOK, response.FailWithMsg(nil, "参数错误: "+err.Error()))
		return
	}

	if params.Page <= 0 {
		params.Page = 1
	}
	if params.PageSize <= 0 {
		params.PageSize = 10
	}

	auths, total, err := k8sAuthService.GetAuthorizations(params)
	if err != nil {
		ctx.JSON(http.StatusOK, response.FailWithMsg(nil, "获取授权列表失败"))
		return
	}

	ctx.JSON(http.StatusOK, response.Success(map[string]interface{}{
		"list":     auths,
		"total":    total,
		"page":     params.Page,
		"pageSize": params.PageSize,
	}))
}

func (c *K8sAuthController) GetAuthorization(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, response.FailWithMsg(nil, "参数错误"))
		return
	}

	auth, err := k8sAuthService.GetAuthorizationByID(uint(id))
	if err != nil {
		ctx.JSON(http.StatusOK, response.FailWithMsg(nil, "获取授权详情失败"))
		return
	}

	ctx.JSON(http.StatusOK, response.Success(auth))
}

func (c *K8sAuthController) CreateAuthorization(ctx *gin.Context) {
	var req request.K8sAuthRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, response.FailWithMsg(nil, "参数错误: "+err.Error()))
		return
	}

	auth := &model.K8sAuthorization{
		UserID:      req.UserID,
		ClusterID:   req.ClusterID,
		Namespace:   req.Namespace,
		Resource:    req.Resource,
		CanView:     req.CanView,
		CanEdit:     req.CanEdit,
		CanDelete:   req.CanDelete,
		CanCreate:   req.CanCreate,
		Description: req.Description,
	}

	if err := k8sAuthService.CreateAuthorization(auth); err != nil {
		ctx.JSON(http.StatusOK, response.FailWithMsg(nil, "创建授权失败"))
		return
	}

	ctx.JSON(http.StatusOK, response.Success(auth))
}

func (c *K8sAuthController) UpdateAuthorization(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, response.FailWithMsg(nil, "参数错误"))
		return
	}

	var req request.K8sAuthRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusOK, response.FailWithMsg(nil, "参数错误: "+err.Error()))
		return
	}

	auth, err := k8sAuthService.GetAuthorizationByID(uint(id))
	if err != nil {
		ctx.JSON(http.StatusOK, response.FailWithMsg(nil, "授权不存在"))
		return
	}

	auth.UserID = req.UserID
	auth.ClusterID = req.ClusterID
	auth.Namespace = req.Namespace
	auth.Resource = req.Resource
	auth.CanView = req.CanView
	auth.CanEdit = req.CanEdit
	auth.CanDelete = req.CanDelete
	auth.CanCreate = req.CanCreate
	auth.Description = req.Description

	if err := k8sAuthService.UpdateAuthorization(auth); err != nil {
		ctx.JSON(http.StatusOK, response.FailWithMsg(nil, "更新授权失败"))
		return
	}

	ctx.JSON(http.StatusOK, response.Success(auth))
}

func (c *K8sAuthController) DeleteAuthorization(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, response.FailWithMsg(nil, "参数错误"))
		return
	}

	if err := k8sAuthService.DeleteAuthorization(uint(id)); err != nil {
		ctx.JSON(http.StatusOK, response.FailWithMsg(nil, "删除授权失败"))
		return
	}

	ctx.JSON(http.StatusOK, response.Success(nil))
}

func (c *K8sAuthController) GetUserAuthorizations(ctx *gin.Context) {
	userIDStr := ctx.Query("user_id")
	if userIDStr == "" {
		ctx.JSON(http.StatusOK, response.FailWithMsg(nil, "缺少user_id参数"))
		return
	}

	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, response.FailWithMsg(nil, "参数错误"))
		return
	}

	auths, err := k8sAuthService.GetUserAuthorizations(uint(userID))
	if err != nil {
		ctx.JSON(http.StatusOK, response.FailWithMsg(nil, "获取用户授权失败"))
		return
	}

	ctx.JSON(http.StatusOK, response.Success(auths))
}

func (c *K8sAuthController) GetUserAuthorizedClusters(ctx *gin.Context) {
	userID := getUserIDFromContext(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.FailWithMsg(nil, "未登录"))
		return
	}

	clusterIDs, err := k8sAuthService.GetUserAuthorizedClusters(userID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.FailWithMsg(nil, "获取授权集群失败"))
		return
	}

	ctx.JSON(http.StatusOK, response.Success(clusterIDs))
}

func (c *K8sAuthController) GetUserAuthorizedNamespaces(ctx *gin.Context) {
	userID := getUserIDFromContext(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.FailWithMsg(nil, "未登录"))
		return
	}

	clusterIDStr := ctx.Query("cluster_id")
	if clusterIDStr == "" {
		ctx.JSON(http.StatusOK, response.FailWithMsg(nil, "缺少cluster_id参数"))
		return
	}

	clusterID, err := strconv.ParseUint(clusterIDStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusOK, response.FailWithMsg(nil, "参数错误"))
		return
	}

	namespaces, err := k8sAuthService.GetUserAuthorizedNamespaces(userID, uint(clusterID))
	if err != nil {
		ctx.JSON(http.StatusOK, response.FailWithMsg(nil, "获取授权命名空间失败"))
		return
	}

	ctx.JSON(http.StatusOK, response.Success(namespaces))
}

func (c *K8sAuthController) GetUserPermissions(ctx *gin.Context) {
	userID := getUserIDFromContext(ctx)
	if userID == 0 {
		ctx.JSON(http.StatusOK, response.FailWithMsg(nil, "未登录"))
		return
	}

	auths, err := k8sAuthService.GetUserPermissions(userID)
	if err != nil {
		ctx.JSON(http.StatusOK, response.FailWithMsg(nil, "获取权限失败"))
		return
	}

	ctx.JSON(http.StatusOK, response.Success(auths))
}

func getUserIDFromContext(ctx *gin.Context) uint {
	claims, exists := ctx.Get("claims")
	if !exists {
		return 0
	}
	customClaims, ok := claims.(*middleware.CustomClaims)
	if !ok {
		return 0
	}
	return customClaims.UserID
}
