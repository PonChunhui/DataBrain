package middleware

import (
	"net/http"
	"strings"

	"devops-backend/global"
	"devops-backend/model/response"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == "OPTIONS" {
			c.Next()
			return
		}

		token := c.GetHeader("Authorization")

		if token == "" {
			token = c.Query("token")
		}

		if token == "" {
			c.JSON(http.StatusOK, response.Unauthorized("未登录或非法访问"))
			c.Abort()
			return
		}

		if strings.HasPrefix(token, "Bearer ") {
			token = strings.TrimPrefix(token, "Bearer ")
		}

		j := NewJWT()
		claims, err := j.ParseToken(token)
		if err != nil {
			if err == TokenExpired {
				c.JSON(http.StatusOK, response.Unauthorized("授权已过期"))
				c.Abort()
				return
			}
			c.JSON(http.StatusOK, response.Unauthorized("授权验证失败"))
			c.Abort()
			return
		}

		c.Set("claims", claims)
		c.Next()
	}
}

type JWT struct {
	SigningKey []byte
}

type CustomClaims struct {
	UserID uint
	jwt.RegisteredClaims
}

var (
	TokenExpired     error = jwt.ErrTokenExpired
	TokenNotValidYet error = jwt.ErrTokenNotValidYet
)

func NewJWT() *JWT {
	return &JWT{
		[]byte(global.GVA_CONFIG.Jwt.SigningKey),
	}
}

func (j *JWT) ParseToken(tokenString string) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return j.SigningKey, nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, TokenNotValidYet
}
