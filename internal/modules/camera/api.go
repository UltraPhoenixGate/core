package camera

import (
	"ultraphx-core/pkg/resp"

	"github.com/gin-gonic/gin"
)

// 新增摄像头
func AddCamera(c *gin.Context) {
	var camera Camera
	if err := c.ShouldBindJSON(&camera); err != nil {
		resp.Error(c, "Invalid request")
		return
	}

	if err := camera.Query().Create(&camera).Error; err != nil {
		resp.Error(c, err.Error())
		return
	}

	resp.OK(c, resp.H{
		"camera": camera,
	})
}

// 获取摄像头列表
func GetCameras(c *gin.Context) {
	var cameras []Camera
	if err := (&Camera{}).Query().Find(&cameras).Error; err != nil {
		resp.Error(c, err.Error())
		return
	}

	resp.OK(c, resp.H{
		"cameras": cameras,
	})
}

func DeleteCamera(c *gin.Context) {
	var camera Camera
	if err := c.ShouldBindJSON(&camera); err != nil {
		resp.Error(c, "Invalid request")
		return
	}

	if err := camera.Query().Delete(&camera).Error; err != nil {
		resp.Error(c, err.Error())
		return
	}

	resp.OK(c, resp.H{
		"camera": camera,
	})
}

func UpdateCamera(c *gin.Context) {
	var camera Camera
	if err := c.ShouldBindJSON(&camera); err != nil {
		resp.Error(c, "Invalid request")
		return
	}

	if err := camera.Query().Save(&camera).Error; err != nil {
		resp.Error(c, err.Error())
		return
	}

	resp.OK(c, resp.H{
		"camera": camera,
	})
}

// 获取当前帧
func GetCurrentFrame(c *gin.Context) {
	var req struct {
		StreamURL string `json:"streamUrl" binding:"required" query:"streamUrl" form:"streamUrl"`
	}
	if err := c.ShouldBind(&req); err != nil {
		resp.Error(c, "Invalid request")
		return
	}

	imagePath, err := captureSnapshot(req.StreamURL)
	if err != nil {
		resp.Error(c, err.Error())
		return
	}

	imageBase64, err := genImageBase64(imagePath)
	if err != nil {
		resp.Error(c, err.Error())
		return
	}

	resp.OK(c, resp.H{
		"image": imageBase64,
	})
}
