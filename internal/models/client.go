package models

import "gorm.io/gorm"

type Client struct {
	gorm.Model
	ID          string `gorm:"primarykey"`
	Name        string
	Type        ClientType
	Permissions []Permission `gorm:"foreignKey:ClientID"`
}

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

func (c *Client) Query() *gorm.DB {
	return DB.Model(c)
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
