package api

import (
	"net/http"
	"ultraphx-core/internal/models"
	"ultraphx-core/internal/services/auth"
	"ultraphx-core/internal/services/sensor"
	"ultraphx-core/pkg/resp"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
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

	token, err := auth.CreateJWTToken(auth.JwtPayload{
		ClientID: client.ID,
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
	claims, err := auth.ParseJWTToken(jwtStr)
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
	err := models.DB.Where("status = ?", models.ClientStatusActive).Find(&clients).Error
	if err != nil {
		logrus.WithError(err).Error("Failed to get clients")
		resp.Error(c, "Failed to get clients")
		return
	}
	resp.OK(c, clients)
}

func GetPendingClients(c *gin.Context) {
	var clients []models.Client
	err := models.DB.Where("status = ?", models.ClientStatusPending).Find(&clients).Error
	if err != nil {
		logrus.WithError(err).Error("Failed to get clients")
		resp.Error(c, "Failed to get clients")
		return
	}
	resp.OK(c, clients)
}

// 新增主动传感器
func AddActiveSensor(c *gin.Context) {
	var req struct {
		Name           string `json:"name" validate:"required"`
		Description    string `json:"description"`
		CollectionInfo struct {
			DataType           models.CollectionDataType `json:"dataType" validate:"required"`
			CollectionPeriod   int                       `json:"collectionPeriod" validate:"required"`
			IPAddress          string                    `json:"ipAddress" validate:"required"`
			CollectionEndpoint string                    `json:"collectionEndpoint" validate:"required"`
			AuthToken          string                    `json:"authToken"`
			CustomLabels       string                    `json:"customLabels"`
		}
	}

	if err := c.ShouldBind(&req); err != nil {
		resp.Error(c, "Invalid request")
		return
	}

	client := models.Client{
		ID:          uuid.New().String(),
		Name:        req.Name,
		Description: req.Description,
		Type:        models.ClientTypeSensorActive,
		Status:      models.ClientStatusActive,
	}

	collectionInfo := models.CollectionInfo{
		ClientID:           client.ID,
		DataType:           req.CollectionInfo.DataType,
		CollectionPeriod:   req.CollectionInfo.CollectionPeriod,
		IPAddress:          req.CollectionInfo.IPAddress,
		CollectionEndpoint: req.CollectionInfo.CollectionEndpoint,
		AuthToken:          req.CollectionInfo.AuthToken,
		CustomLabels:       req.CollectionInfo.CustomLabels,
	}

	err := models.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&client).Error; err != nil {
			return err
		}
		if err := tx.Create(&collectionInfo).Error; err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		logrus.WithError(err).Error("Failed to create client")
		resp.Error(c, "Failed to create client")
		return
	}

	resp.OK(c, client)
}

func RemoveClient(c *gin.Context) {
	var req struct {
		ClientID string `json:"clientID" validate:"required"`
	}
	if err := c.ShouldBind(&req); err != nil {
		resp.Error(c, "Invalid request")
		return
	}

	client := models.Client{
		ID: req.ClientID,
	}
	if err := client.Query().Find(&client).Error; err != nil {
		logrus.WithError(err).Error("Failed to find client")
		resp.Error(c, "Failed to find client")
		return
	}

	err := models.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Delete(&client).Error; err != nil {
			return err
		}
		if err := tx.Where("client_id = ?", client.ID).Delete(&models.CollectionInfo{}).Error; err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		logrus.WithError(err).Error("Failed to remove client")
		resp.Error(c, "Failed to remove client")
		return
	}

	resp.OK(c, client)
}

func SetClientStatus(c *gin.Context) {
	var req struct {
		ClientID string `json:"clientID" validate:"required"`
		Status   string `json:"status" validate:"required"`
	}
	if err := c.ShouldBind(&req); err != nil {
		resp.Error(c, "Invalid request")
		return
	}

	client := models.Client{
		ID: req.ClientID,
	}
	if err := client.Query().Find(&client).Error; err != nil {
		logrus.WithError(err).Error("Failed to find client")
		resp.Error(c, "Failed to find client")
		return
	}

	client.Status = models.ClientStatus(req.Status)
	if err := client.Query().Save(&client).Error; err != nil {
		logrus.WithError(err).Error("Failed to update client")
		resp.Error(c, "Failed to update client")
		return
	}

	resp.OK(c, client)
}

// 扫描主动传感器
func ScanActiveSensor(c *gin.Context) {
	// 在本地网络中搜索
	devices := sensor.ScanSensors()
	resp.OK(c, devices)
}

// 初始化本地客户端
func SetupLocalClient(c *gin.Context) {
	req := struct {
		SystemPassword string `json:"systemPassword" validate:"required"`
	}{}
	if err := c.ShouldBind(&req); err != nil {
		resp.Error(c, "Invalid request")
		return
	}

	passWdBcrypt, err := auth.HashPassword(req.SystemPassword)
	if err != nil {
		logrus.WithError(err).Error("Failed to hash password")
		resp.Error(c, "Failed to hash password")
		return
	}

	client := models.Client{
		ID:          uuid.New().String(),
		Name:        "核心客户端",
		Description: "UltraPhoenix 核心客户端",
		Type:        models.ClientTypeLocal,
		Status:      models.ClientStatusActive,
		Payload:     passWdBcrypt,
	}

	err = client.Query().Create(&client).Error
	if err != nil {
		logrus.WithError(err).Error("Failed to create client")
		resp.Error(c, "Failed to create client")
		return
	}
	resp.OK(c, client)
}

// 本地客户端登录
func LoginLocalClient(c *gin.Context) {
	req := struct {
		SystemPassword string `json:"systemPassword" validate:"required"`
	}{}
	if err := c.ShouldBind(&req); err != nil {
		resp.Error(c, "Invalid request")
		return
	}

	client := models.Client{
		Type: models.ClientTypeLocal,
	}
	if err := client.Query().Where("type = ?", models.ClientTypeLocal).First(&client).Error; err != nil {
		logrus.WithError(err).Error("Failed to find client")
		resp.Error(c, "Failed to find client")
		return
	}

	if err := auth.ComparePassword(req.SystemPassword, client.Payload); err != nil {
		resp.Error(c, "Invalid password")
		return
	}

	token, err := auth.CreateJWTToken(auth.JwtPayload{
		ClientID: client.ID,
		Type:     client.Type,
	})

	if err != nil {
		logrus.WithError(err).Error("Failed to create token")
		resp.Error(c, "Failed to create token")
		return
	}

	resp.OK(c, gin.H{
		"token": token,
	})
}

// 检查本地客户端是否存在
func IsLocalClientExist(c *gin.Context) {
	client := models.Client{
		Type: models.ClientTypeLocal,
	}
	if err := client.Query().Where("type = ?", models.ClientTypeLocal).First(&client).Error; err != nil {
		resp.OK(c, gin.H{
			"exist": false,
		})
		return
	}
	resp.OK(c, gin.H{
		"exist": true,
	})
}
