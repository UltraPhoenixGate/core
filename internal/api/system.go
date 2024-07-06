package api

import (
	"fmt"
	"net"
	"net/http"
	"os/exec"
	"strings"
	"time"
	"ultraphx-core/pkg/resp"

	"github.com/gin-gonic/gin"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/load"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/sirupsen/logrus"
	"github.com/vcraescu/go-xrandr"
)

type SystemInfo struct {
	Version string `json:"version"`
	Uptime  uint64 `json:"uptime"`
	Load    string `json:"load"`
	Memory  struct {
		Total     uint64 `json:"total"`
		Used      uint64 `json:"used"`
		Available uint64 `json:"available"`
	} `json:"memory"`
	Disk struct {
		Total     uint64 `json:"total"`
		Used      uint64 `json:"used"`
		Available uint64 `json:"available"`
	} `json:"disk"`
}

func GetSystemInfo(c *gin.Context) {
	sysinfo := SystemInfo{}
	// 获取系统版本号
	sysinfo.Version = "1.0.0"
	// 获取系统运行时间
	uptime, _ := host.Uptime()
	sysinfo.Uptime = uptime
	// 获取系统负载
	load, _ := load.Avg()
	sysinfo.Load = fmt.Sprintf("%.2f %.2f %.2f", load.Load1, load.Load5, load.Load15)
	// 获取系统内存使用情况
	mem, _ := mem.VirtualMemory()
	sysinfo.Memory.Total = mem.Total
	sysinfo.Memory.Used = mem.Used
	sysinfo.Memory.Available = mem.Available
	// 获取系统磁盘使用情况
	disk, _ := disk.Usage("/")
	sysinfo.Disk.Total = disk.Total
	sysinfo.Disk.Used = disk.Used
	sysinfo.Disk.Available = disk.Free
	resp.OK(c, sysinfo)
}

// 设置系统分辨率
func SetResolution(c *gin.Context) {
	req := struct {
		Width     int    `json:"width" binding:"required"`
		Height    int    `json:"height" binding:"required"`
		MonitorID string `json:"monitorId"`
		Freq      int    `json:"freq"`
	}{
		Freq: 30,
	}

	if err := c.ShouldBind(&req); err != nil {
		resp.Error(c, "Invalid request")
		return
	}

	// 设置分辨率
	cmd := exec.Command("xrandr", "--output", req.MonitorID, "--mode", fmt.Sprintf("%dx%d", req.Width, req.Height), "--rate", fmt.Sprintf("%d", req.Freq))
	if err := cmd.Run(); err != nil {
		resp.Error(c, "Failed to set resolution")
		return
	}

	resp.OK(c, nil)
}

// 获取全部可用分辨率
func GetMonitorResolutions(c *gin.Context) {
	screens, err := xrandr.GetScreens()
	if err != nil {
		resp.Error(c, "Failed to get screens")
		return
	}
	resp.OK(c, screens)
}

// 检查网络连接
func CheckNetwork(c *gin.Context) {
	url := "https://www.baidu.com"
	client := http.Client{
		Timeout: 5 * time.Second,
	}
	res, err := client.Get(url)
	if err != nil {
		resp.OK(c, gin.H{
			"status": false,
		})
		return
	}

	if res.StatusCode != http.StatusOK {
		resp.OK(c, gin.H{
			"status": false,
		})
		return
	}
	resp.OK(c, gin.H{
		"status": true,
	})
}

type NetworkInfo struct {
	IP         string `json:"ip"`
	Connected  bool   `json:"connected"`
	Device     string `json:"device"`
	DeviceType string `json:"deviceType"` // ethernet, wifi
}

// 获取网络信息
func GetNetworkInfos(c *gin.Context) {
	networkInfos := getNetworkInfo()
	resp.OK(c, networkInfos)
}

func getNetworkInfo() []NetworkInfo {
	var networkInfos []NetworkInfo

	interfaces, err := net.Interfaces()
	if err != nil {
		fmt.Println(err)
		return networkInfos
	}

	for _, iface := range interfaces {
		// 仅获取 e / w 开头的网卡
		logrus.Info("iface.Name: ", iface.Name)
		if !strings.HasPrefix(iface.Name, "e") && !strings.HasPrefix(iface.Name, "w") {
			continue
		}

		addrs, _ := iface.Addrs()
		connected := false
		ip := ""
		for _, addr := range addrs {
			if strings.Contains(addr.String(), ".") {
				connected = true
				ip = strings.Split(addr.String(), "/")[0]
				break
			}
		}

		var deviceType string
		if strings.HasPrefix(iface.Name, "w") {
			deviceType = "wifi"
		} else {
			deviceType = "ethernet"
		}

		networkInfos = append(networkInfos, NetworkInfo{
			IP:         ip,
			Device:     iface.Name,
			DeviceType: deviceType,
			Connected:  connected,
		})
	}

	return networkInfos
}

// 打开系统网络设置
func OpenNetworkSettings(c *gin.Context) {
	cmd := exec.Command("nm-connection-editor")
	if err := cmd.Start(); err != nil {
		resp.Error(c, "Failed to open network settings")
		return
	}
	resp.OK(c, nil)
}
