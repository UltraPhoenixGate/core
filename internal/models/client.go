package models

import (
	"errors"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Client struct {
	Model
	ID          string          `gorm:"primarykey" json:"id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Status      ClientStatus    `json:"status"`
	Type        ClientType      `json:"type"`
	Payload     string          `json:"payload"`
	Permissions []Permission    `gorm:"foreignKey:ClientID" json:"permissions"`
	Collection  *CollectionInfo `gorm:"foreignKey:ClientID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"collection"`
}

type CollectionInfo struct {
	ClientID           string             `gorm:"primarykey" json:"clientId"`
	DataType           CollectionDataType `json:"dataType"`           // 采集数据类型
	CollectionPeriod   int                `json:"collectionPeriod"`   // 采集周期，单位为秒
	LastCollectionTime time.Time          `json:"lastCollectionTime"` // 上次采集时间
	IPAddress          string             `json:"ipAddress"`          // 客户端 IP 地址
	CollectionEndpoint string             `json:"collectionEndpoint"` // 采集地址（URL）
	AuthToken          string             `json:"authToken"`          // 鉴权信息，例如 token
	CustomLabels       string             `json:"customLabels"`       // 自定义标签
}

type CollectionDataType string

const (
	CollectionDataTypeJSON    CollectionDataType = "json"
	CollectionDataTypeMetrics CollectionDataType = "metrics"
)

type ClientStatus string

const (
	ClientStatusPending  ClientStatus = "pending"
	ClientStatusActive   ClientStatus = "active"
	ClientStatusExpired  ClientStatus = "expired"
	ClientStatusDisabled ClientStatus = "disabled"
)

type ClientType string

const (
	ClientTypePlugin ClientType = "plugin"
	// 传感器（被动）
	ClientTypeSensor ClientType = "sensor"
	// 传感器（主动）
	ClientTypeSensorActive ClientType = "sensor_active"
	// 本地客户端
	ClientTypeLocal ClientType = "local"
)

type Permission struct {
	Model
	ClientID uint   `gorm:"index" json:"clientId"`
	Topic    string `json:"topic"`
	Type     PermissionType
}

type PermissionType int

const (
	PermissionTypeRead PermissionType = iota + 1
	PermissionTypeWrite
)

func (p PermissionType) String() string {
	switch p {
	case PermissionTypeRead:
		return "r"
	case PermissionTypeWrite:
		return "w"
	default:
		return ""
	}
}

func (p Permission) String() string {
	// ex: data-rw
	return p.Topic + "-" + p.Type.String()
}

func PrasePermission(s string) (Permission, error) {
	if strings.Count(s, "-") != 1 {
		return Permission{}, errors.New("invalid permission string")
	}
	topic := strings.Split(s, "-")[0]
	pType := strings.Split(s, "-")[1]
	switch pType {
	case "r":
		return Permission{Topic: topic, Type: PermissionTypeRead}, nil
	case "w":
		return Permission{Topic: topic, Type: PermissionTypeWrite}, nil
	}
	return Permission{}, errors.New("invalid permission type")
}

func (c *Client) Query() *gorm.DB {
	return DB.Model(c)
}

func (c *CollectionInfo) Query() *gorm.DB {
	return DB.Model(c)
}

func (c *Client) CheckIsExpired() {
	// if one minute later, the client is still pending, then set it to expired
	if c.Status == ClientStatusPending && time.Now().After(c.CreatedAt.Add(1*time.Minute)) {
		logrus.WithField("client", c.ID).Info("Client is expired")
		c.Status = ClientStatusExpired
		c.Query().Save(c)
	}
}

// DefaultPluginPermissions 插件默认权限
var DefaultPluginPermissions = map[string][]PermissionType{
	"init": {PermissionTypeWrite},
}

// DefaultSensorPermissions 传感器默认权限
var DefaultSensorPermissions = map[string][]PermissionType{
	"init": {PermissionTypeWrite},
}

// DefaultSensorFullPermissions 传感器默认完全权限
var DefaultSensorFullPermissions = map[string][]PermissionType{
	"init":  {PermissionTypeWrite},
	"data":  {PermissionTypeWrite},
	"alert": {PermissionTypeWrite},
}
