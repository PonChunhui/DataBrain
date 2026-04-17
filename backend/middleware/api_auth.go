package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"devops-backend/global"
	"devops-backend/model"
	"devops-backend/model/response"
)

func ApiAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusOK, response.Unauthorized("未登录"))
			c.Abort()
			return
		}

		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}

		method := c.Request.Method

		var api model.Api
		err := global.GVA_DB.Where("path = ? AND method = ?", path, method).First(&api).Error
		if err != nil {
			c.Next()
			return
		}

		var userRoles []model.UserRole
		if err := global.GVA_DB.Where("user_id = ?", userID).Find(&userRoles).Error; err != nil {
			c.JSON(http.StatusOK, response.Unauthorized("获取用户角色失败"))
			c.Abort()
			return
		}

		if len(userRoles) == 0 {
			c.JSON(http.StatusOK, response.Unauthorized("无权限访问"))
			c.Abort()
			return
		}

		roleIDs := make([]uint, len(userRoles))
		for i, ur := range userRoles {
			roleIDs[i] = ur.RoleID
		}

		var roleApis []model.RoleApi
		if err := global.GVA_DB.Where("role_id IN ? AND api_id = ?", roleIDs, api.ID).First(&roleApis).Error; err != nil {
			c.JSON(http.StatusOK, response.Unauthorized("无权限访问该API"))
			c.Abort()
			return
		}

		c.Next()
	}
}
