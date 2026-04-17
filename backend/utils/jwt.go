package utils

import (
	"time"

	"devops-backend/global"
	"github.com/golang-jwt/jwt/v4"
)

type JWT struct {
	SigningKey []byte
}

type CustomClaims struct {
	UserID uint
	jwt.RegisteredClaims
}

func NewJWT() *JWT {
	return &JWT{
		[]byte(global.GVA_CONFIG.Jwt.SigningKey),
	}
}

func (j *JWT) CreateToken(userID uint) (string, error) {
	claims := CustomClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(global.GVA_CONFIG.Jwt.ExpiresTime) * time.Second)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    global.GVA_CONFIG.Jwt.Issuer,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.SigningKey)
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
	return nil, jwt.NewValidationError("invalid token", jwt.ValidationErrorClaimsInvalid)
}
