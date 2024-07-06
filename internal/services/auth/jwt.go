package auth

import (
	"fmt"
	"strings"
	"ultraphx-core/internal/models"

	"github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"
)

type JwtPayload struct {
	ClientID string            `json:"id"`
	Type     models.ClientType `json:"type"`
}

type JwtClaims struct {
	jwt.RegisteredClaims
	JwtPayload
}

func CreateJWTToken(payload JwtPayload) (string, error) {
	claims := JwtClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:  "",
			Subject: "token",
		},
		JwtPayload: payload,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(GetJwtPrivateKey()))
}

func ParseJWTToken(rawToken string) (*JwtClaims, error) {
	rawToken = strings.TrimPrefix(rawToken, "Bearer ")
	token, err := jwt.ParseWithClaims(rawToken, &JwtClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(GetJwtPrivateKey()), nil
	})
	if err != nil {
		logrus.WithError(err).Error("Failed to parse token")
		return nil, err
	}

	claims := token.Claims.(*JwtClaims)

	return claims, nil
}
