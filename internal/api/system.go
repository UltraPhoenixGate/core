package api

import (
	"fmt"
	"ultraphx-core/pkg/resp"

	"github.com/gin-gonic/gin"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/load"
	"github.com/shirou/gopsutil/v4/mem"
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
