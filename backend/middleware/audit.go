package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"
	"time"

	"devops-backend/global"
	"devops-backend/model"
	"devops-backend/service"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func AuditMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == "OPTIONS" {
			c.Next()
			return
		}

		method := c.Request.Method
		if method != "POST" && method != "PUT" && method != "DELETE" {
			c.Next()
			return
		}

		path := c.Request.URL.Path
		if strings.Contains(path, "/login") || strings.Contains(path, "/register") {
			c.Next()
			return
		}

		startTime := time.Now()

		var requestBody []byte
		if c.Request.Body != nil {
			requestBody, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = blw

		c.Next()

		statusCode := c.Writer.Status()
		status := "success"
		if statusCode >= 400 {
			status = "failed"
		}

		claims, exists := c.Get("claims")
		var userID uint
		var username string

		if exists {
			if customClaims, ok := claims.(*CustomClaims); ok {
				userID = customClaims.UserID
				var user model.User
				if err := global.GVA_DB.First(&user, userID).Error; err == nil {
					username = user.Username
				}
			}
		}

		action := getAction(method)
		resource := getResource(path)
		resourceID := getResourceID(path)

		ip := c.ClientIP()

		detail := map[string]interface{}{
			"method":      method,
			"path":        path,
			"status_code": statusCode,
			"duration":    time.Since(startTime).Milliseconds(),
			"request":     safeRequestBody(requestBody),
		}

		if statusCode >= 400 {
			detail["response"] = blw.body.String()
		}

		auditService := &service.AuditService{}
		if err := auditService.RecordAudit(userID, username, action, resource, resourceID, ip, status, detail); err != nil {
			global.GVA_LOG.Error("记录审计日志失败", zap.Error(err))
		}
	}
}

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func getAction(method string) string {
	switch method {
	case "POST":
		return "create"
	case "PUT":
		return "update"
	case "DELETE":
		return "delete"
	default:
		return method
	}
}

func getResource(path string) string {
	parts := strings.Split(path, "/")
	for i := len(parts) - 1; i >= 0; i-- {
		if parts[i] != "" && !isNumeric(parts[i]) {
			return parts[i]
		}
	}
	return "unknown"
}

func getResourceID(path string) string {
	parts := strings.Split(path, "/")
	for i := len(parts) - 1; i >= 0; i-- {
		if isNumeric(parts[i]) {
			return parts[i]
		}
	}
	return ""
}

func isNumeric(s string) bool {
	if s == "" {
		return false
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

func safeRequestBody(body []byte) string {
	if len(body) == 0 {
		return ""
	}

	var obj map[string]interface{}
	if err := json.Unmarshal(body, &obj); err != nil {
		return string(body)
	}

	if _, ok := obj["password"]; ok {
		obj["password"] = "***"
	}
	if _, ok := obj["old_password"]; ok {
		obj["old_password"] = "***"
	}
	if _, ok := obj["new_password"]; ok {
		obj["new_password"] = "***"
	}

	result, _ := json.Marshal(obj)
	return string(result)
}
