package camera

import (
	"io"
	"os"
	"os/exec"
	"strconv"
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

func OpenStream(c *gin.Context) {
	req := struct {
		ID     string `binding:"required" form:"id"`
		Fps    int    `form:"fps"`
		Width  int    `form:"width"`
		Height int    `form:"height"`
	}{
		Fps:    25,
		Width:  640,
		Height: 480,
	}
	if err := c.ShouldBindQuery(&req); err != nil {
		resp.Error(c, "Invalid request")
		return
	}
	camera := Camera{}
	if err := camera.Query().Where("id = ?", req.ID).First(&camera).Error; err != nil {
		resp.Error(c, "Camera not found")
		return
	}

	// ffmpeg params
	params := []string{
		"-i", camera.StreamUrl,
		"-f", "mpegts",
		"-codec:v", "mpeg1video",
		"-s", strconv.Itoa(req.Width) + "x" + strconv.Itoa(req.Height),
		"-r", strconv.Itoa(req.Fps),
		"-b:v", "800k",
		"-bf", "0",
		"-q:v", "1",
		"-muxdelay", "0.001",
		"-",
	}

	c.Header("Content-Type", "video/mp2t")
	c.Header("Transfer-Encoding", "chunked")

	cmd := exec.Command("ffmpeg", params...)
	cmd.Stdout = c.Writer

	stderr, err := cmd.StderrPipe()
	if err != nil {
		resp.Error(c, "Failed to capture stderr")
		return
	}

	go func() {
		io.Copy(os.Stderr, stderr) // Log stderr to standard error or a file
	}()

	if err := cmd.Start(); err != nil {
		resp.Error(c, "Failed to start stream")
		return
	}

	err = cmd.Wait()
	if err != nil {
		return
	}
}
