package camera

import (
	"io"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"ultraphx-core/pkg/resp"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/use-go/onvif"
	"github.com/use-go/onvif/device"
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

type OnvifDevice struct {
	Name  string `json:"name"`
	Xaddr string `json:"xaddr"`
}

func ScanOnvifDevices(c *gin.Context) {
	// 获取全部接口
	interfaces, err := net.Interfaces()
	if err != nil {
		resp.Error(c, err.Error())
		return
	}

	// 获取所有可用的onvif设备
	allDevices := make([]onvif.Device, 0)
	for _, i := range interfaces {
		// 跳过 docker br
		if strings.HasPrefix(i.Name, "br") {
			continue
		}
		// 跳过 veth
		if strings.HasPrefix(i.Name, "veth") {
			continue
		}
		logrus.Infof("Scanning onvif devices at interface: %s", i.Name)
		devices, err := onvif.GetAvailableDevicesAtSpecificEthernetInterface(i.Name)
		if err != nil {
			logrus.Errorf("Failed to get onvif devices at interface %s: %v", i.Name, err)
			continue
		}
		allDevices = append(allDevices, devices...)
	}
	allOnvifDevices := make([]OnvifDevice, 0)

	for _, dev := range allDevices {
		getHostnameRes := device.GetHostnameResponse{}
		err := CallDeviceMethod(&dev, device.GetHostname{}, "GetHostnameResponse", &getHostnameRes)
		if err != nil {
			logrus.Errorf("Failed to get hostname: %v", err)
			continue
		}
		// logrus.WithField("hostname", getHostnameRes.HostnameInformation.Name).Info("Got hostname")
		allOnvifDevices = append(allOnvifDevices, OnvifDevice{
			Name:  string(getHostnameRes.HostnameInformation.Name),
			Xaddr: GetDeviceXAddr(&dev),
		})
	}

	resp.OK(c, resp.H{
		"devices": allOnvifDevices,
	})
}
