package router

import (
	"net/http"
	"ultraphx-core/internal/models"
	"ultraphx-core/internal/services/auth"
	"ultraphx-core/pkg/resp"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func AuthMiddleware(c *gin.Context) {
	jwtStr := c.GetHeader("Authorization")
	if jwtStr == "" {
		resp.ErrorWithCode(c, http.StatusUnauthorized, "Unauthorized")
		return
	}
	claims, err := auth.ParseJWEToken(jwtStr)
	if err != nil {
		resp.ErrorWithCode(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// now we not check the token expiration
	// if claims.Expiry.Time().Before(time.Now()) {
	// 	return false, errors.New("token expired")
	// }

	client := models.Client{
		ID: claims.ClientID,
	}

	if err := client.Query().Find(&client).Error; err != nil {
		logrus.WithError(err).Error("Failed to find client")
		resp.ErrorWithCode(c, http.StatusUnauthorized, "Unauthorized")
		return
	}
	client.CheckIsExpired()

	if client.Status != models.ClientStatusActive {
		resp.ErrorWithCode(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	c.Set("client", &client)
	c.Set("claims", claims)
	c.Next()
}

func CorsMiddleware(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
	c.Header("Access-Control-Expose-Headers", "Authorization")
	c.Header("Access-Control-Max-Age", "86400")
	c.Header("Access-Control-Allow-Credentials", "true")
	if c.Request.Method == "OPTIONS" {
		c.AbortWithStatus(http.StatusOK)
	}
	c.Next()
}
