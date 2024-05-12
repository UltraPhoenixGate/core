package models

import (
	"errors"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Client struct {
	gorm.Model
	ID          string `gorm:"primarykey"`
	Name        string
	Description string
	Status      ClientStatus
	Type        ClientType
	Permissions []Permission `gorm:"foreignKey:ClientID"`
}

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
	ClientTypeSensor ClientType = "sensor"
)

type Permission struct {
	gorm.Model
	ClientID uint `gorm:"index"`
	Topic    string
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
