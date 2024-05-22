package camera

import (
	"ultraphx-core/internal/models"

	"gorm.io/gorm"
)

type Camera struct {
	models.Model
	Name        string         `json:"name" binding:"required"`
	Description string         `json:"description" binding:"required"`
	StreamUrl   string         `json:"streamUrl" binding:"required"`
	Protocol    StreamProtocol `json:"protocol"`
	Enabled     bool           `json:"enabled"`
	Extra       string         `json:"extra"`
}

func (c *Camera) Query() *gorm.DB {
	return models.DB.Model(c)
}

type StreamProtocol string

const (
	RTSP StreamProtocol = "rtsp"
	RTMP StreamProtocol = "rtmp"
	HTTP StreamProtocol = "http"
)
