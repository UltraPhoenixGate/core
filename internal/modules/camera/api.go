package camera

import (
	"io"
	"net"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"ultraphx-core/pkg/resp"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/use-go/onvif"
	"github.com/use-go/onvif/device"
	"github.com/use-go/onvif/media"
	onvifDef "github.com/use-go/onvif/xsd/onvif"
)

// 新增摄像头
func AddCamera(c *gin.Context) {
	var camera Camera
	if err := c.ShouldBindJSON(&camera); err != nil {
		resp.Error(c, "Invalid request")
		return
	}
	camera.ID = uuid.New().String()

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
	id := c.Query("id")
	if err := (&Camera{}).Query().Where("id = ?", id).Delete(&Camera{}).Error; err != nil {
		resp.Error(c, err.Error())
		return
	}
	resp.OK(c, nil)
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

	// ffmpeg params - mpeg1video
	// params := []string{
	// 	"-rtsp_transport", "tcp",
	// 	"-i", camera.StreamUrl,
	// 	"-f", "mpegts",
	// 	"-codec:v", "mpeg1video",
	// 	"-s", strconv.Itoa(req.Width) + "x" + strconv.Itoa(req.Height),
	// 	"-r", strconv.Itoa(req.Fps),
	// 	"-b:v", "800k",
	// 	"-bf", "0",
	// 	"-q:v", "1",
	// 	"-muxdelay", "0.001",
	// 	"-",
	// }
	// c.Header("Content-Type", "video/mp2t")
	// ffmpeg params - h264
	params := []string{
		"-rtsp_transport", "tcp",
		"-i", camera.StreamUrl,
		"-f", "flv",
		"-flags", "low_delay",
		"-codec:v", "libx264", // 使用H.264编码
		// 音频编码 - aac
		"-codec:a", "aac",
		"-s", strconv.Itoa(req.Width) + "x" + strconv.Itoa(req.Height),
		"-r", strconv.Itoa(req.Fps),
		"-b:v", "800k",
		"-bf", "0",
		"-q:v", "1",
		"-tune", "zerolatency", // 0延迟
		"-g", strconv.Itoa(req.Fps * 2), // 关键帧间隔
		"-muxdelay", "0.001",
		"-preset", "ultrafast",
		"-",
	}
	c.Header("Content-Type", "video/flv")

	c.Header("Transfer-Encoding", "chunked")
	c.Header("Access-Control-Allow-Origin", "*") // 添加CORS头
	c.Header("Connection", "keep-alive")

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
	Name         string `json:"name"`
	Xaddr        string `json:"xAddr"`
	Manufacturer string `json:"manufacturer"`
	Model        string `json:"model"`
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
		devInfoRes := device.GetDeviceInformationResponse{}
		err = CallDeviceMethod(&dev, device.GetDeviceInformation{}, "GetDeviceInformationResponse", &devInfoRes)
		if err != nil {
			logrus.Errorf("Failed to get device information: %v", err)
			continue
		}

		// logrus.WithField("hostname", getHostnameRes.HostnameInformation.Name).Info("Got hostname")
		allOnvifDevices = append(allOnvifDevices, OnvifDevice{
			Name:         string(getHostnameRes.HostnameInformation.Name),
			Xaddr:        GetDeviceXAddr(&dev),
			Manufacturer: devInfoRes.Manufacturer,
			Model:        devInfoRes.Model,
		})
	}

	resp.OK(c, resp.H{
		"devices": allOnvifDevices,
	})
}

func GetOnvifDeviceInfo(c *gin.Context) {
	var req struct {
		Xaddr    string `json:"xAddr" binding:"required" form:"xAddr"`
		User     string `json:"user" form:"user"`
		Password string `json:"password" form:"password"`
	}
	if err := c.ShouldBind(&req); err != nil {
		resp.Error(c, "Invalid request")
		return
	}

	dev, err := onvif.NewDevice(onvif.DeviceParams{
		Xaddr:    req.Xaddr,
		Username: req.User,
		Password: req.Password,
	})
	if err != nil {
		resp.Error(c, err.Error())
		return
	}

	devInfo := device.GetDeviceInformationResponse{}
	err = CallDeviceMethod(dev, device.GetDeviceInformation{}, "GetDeviceInformationResponse", &devInfo)
	if err != nil {
		resp.Error(c, err.Error())
		return
	}

	mediaProfiles := media.GetProfilesResponse{}
	err = CallDeviceMethod(dev, media.GetProfiles{}, "GetProfilesResponse", &mediaProfiles)
	if err != nil {
		resp.Error(c, err.Error())
		return
	}
	defaultProfile := mediaProfiles.Profiles[0]
	streamUrlRes := media.GetStreamUriResponse{}
	err = CallDeviceMethod(dev, media.GetStreamUri{
		ProfileToken: defaultProfile.Token,
		StreamSetup: onvifDef.StreamSetup{
			Stream: "RTP-Unicast",
			Transport: onvifDef.Transport{
				Protocol: "RTSP",
			},
		},
	}, "GetStreamUriResponse", &streamUrlRes)
	if err != nil {
		resp.Error(c, err.Error())
		return
	}

	streamUrl, err := url.Parse(string(streamUrlRes.MediaUri.Uri))
	if err != nil {
		resp.Error(c, err.Error())
		return
	}

	if req.User != "" && req.Password != "" {
		streamUrl.User = url.UserPassword(req.User, req.Password)
	}

	resp.OK(c, resp.H{
		"info":      devInfo,
		"streamUrl": streamUrl.String(),
	})
}
