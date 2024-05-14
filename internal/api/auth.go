package api

import (
	"net/http"
	"ultraphx-core/internal/models"
	"ultraphx-core/internal/services/auth"
	"ultraphx-core/pkg/resp"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

func HandlePluginRegister(c *gin.Context) {
	var req struct {
		Name        string            `json:"name" validate:"required"`
		Description string            `json:"description"`
		Type        models.ClientType `json:"type" validate:"required"`
		Permissions []string          `json:"permissions"`
	}
	if err := c.ShouldBind(&req); err != nil {
		resp.Error(c, "Invalid request")
		return
	}

	clientPermissions := make([]models.Permission, 0, len(req.Permissions))
	for _, p := range req.Permissions {
		permission, err := models.PrasePermission(p)
		if err != nil {
			resp.Error(c, "Permission not found")
			return
		}
		clientPermissions = append(clientPermissions, permission)
	}

	client := models.Client{
		ID:          uuid.New().String(),
		Name:        req.Name,
		Description: req.Description,
		Type:        req.Type,
		Permissions: clientPermissions,
		Status:      models.ClientStatusPending,
	}
	err := client.Query().Create(&client).Error
	if err != nil {
		logrus.WithError(err).Error("Failed to create client")
		resp.Error(c, "Failed to create client")
		return
	}

	token, err := auth.CreateJWEToken(auth.JwtPayload{
		ClientID: client.ID,
		Name:     client.Name,
		Type:     client.Type,
	})

	if err != nil {
		logrus.WithError(err).Error("Failed to create token")
		resp.Error(c, "Failed to create token")
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"token": token,
	})
}

func HandlePluginCheckActive(c *gin.Context) {
	jwtStr := c.GetHeader("Authorization")
	if jwtStr == "" {
		resp.Error(c, "Unauthorized")
		return
	}
	claims, err := auth.ParseJWEToken(jwtStr)
	if err != nil {
		resp.Error(c, err.Error())
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
		resp.Error(c, "Failed to find client")
		return
	}

	client.CheckIsExpired()

	c.JSON(http.StatusOK, gin.H{
		"active": client.Status == models.ClientStatusActive,
		"status": client.Status,
	})
}

func GetConnectedClients(c *gin.Context) {
	var clients []models.Client
	err := models.DB.Find(&clients).Where("status = ?", models.ClientStatusActive).Error
	if err != nil {
		logrus.WithError(err).Error("Failed to get clients")
		resp.Error(c, "Failed to get clients")
		return
	}
	resp.OK(c, clients)
}

func GetPendingClients(c *gin.Context) {
	var clients []models.Client
	err := models.DB.Find(&clients).Where("status = ?", models.ClientStatusPending).Error
	if err != nil {
		logrus.WithError(err).Error("Failed to get clients")
		resp.Error(c, "Failed to get clients")
		return
	}
	resp.OK(c, clients)
}
