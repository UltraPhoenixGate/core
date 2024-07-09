package sensor

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

type Metadata struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Version     int    `json:"version"`
}

type Entrypoint struct {
	Path        string `json:"path"`
	Description string `json:"description"`
	Method      string `json:"method"`
	Type        string `json:"type"`
}

type DeviceMetadata struct {
	Metadata    Metadata     `json:"metadata"`
	Entrypoints []Entrypoint `json:"entrypoints"`
}

type DeviceData struct {
	Temperature int `json:"temperature"`
	Humidity    int `json:"humidity"`
}

type Device struct {
	IP          string       `json:"ip"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Version     int          `json:"version"`
	Endpoints   []Entrypoint `json:"endpoints"`
}

var wg sync.WaitGroup

func ScanSensors() []Device {
	devices := make([]Device, 0)
	ips := getLocalIPs()

	for _, ip := range ips {
		wg.Add(1)
		go scanNetwork(ip, &devices)
	}

	wg.Wait()

	return devices
}

func getLocalIPs() []string {
	var ips []string
	ifaces, err := net.Interfaces()
	if err != nil {
		fmt.Println(err)
		return ips
	}

	for _, iface := range ifaces {
		// 跳过 br-、veth、docker
		if strings.HasPrefix(iface.Name, "br") || strings.HasPrefix(iface.Name, "veth") || strings.HasPrefix(iface.Name, "docker") {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			fmt.Println(err)
			continue
		}
		for _, addr := range addrs {
			if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
				if ipNet.IP.To4() != nil {
					ips = append(ips, ipNet.IP.String())
				}
			}
		}
	}
	return ips
}

func scanNetwork(ip string, devices *[]Device) {
	defer wg.Done()

	ipRange := getIPRange(ip)
	var innerWg sync.WaitGroup

	for _, targetIP := range ipRange {
		innerWg.Add(1)
		go func(targetIP string) {
			defer innerWg.Done()
			if isPortOpen(targetIP, 80) {
				device := scanDevice(targetIP)
				if device != nil {
					*devices = append(*devices, *device)
				}
			}
		}(targetIP)
	}

	innerWg.Wait()
}

func getIPRange(ip string) []string {
	var ips []string
	// 首先删除最后一位
	ip = strings.Join(strings.Split(ip, ".")[:3], ".") + "."
	for i := 1; i <= 254; i++ {
		ips = append(ips, fmt.Sprintf("%s.%d", ip[:len(ip)-1], i))
	}
	return ips
}

func isPortOpen(ip string, port int) bool {
	timeout := time.Second
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", ip, port), timeout)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

func scanDevice(ip string) *Device {
	url := fmt.Sprintf("http://%s/metadata", ip)
	resp, err := http.Get(url)
	if err == nil {
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err == nil {
			var metadata DeviceMetadata
			if json.Unmarshal(body, &metadata) == nil {
				logrus.Info("Found device: ", metadata.Metadata.Name)
				return &Device{
					IP:          ip,
					Name:        metadata.Metadata.Name,
					Description: metadata.Metadata.Description,
					Version:     metadata.Metadata.Version,
					Endpoints:   metadata.Entrypoints,
				}
			}
		}
	}

	url = fmt.Sprintf("http://%s/data/json", ip)
	resp, err = http.Get(url)
	if err == nil {
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err == nil {
			var deviceData DeviceData
			if json.Unmarshal(body, &deviceData) == nil {
				logrus.Info("Found device: ", ip)
				return &Device{
					IP:   ip,
					Name: ip,
					Endpoints: []Entrypoint{
						{
							Path:        "/data/json",
							Description: "获取数据",
							Method:      "GET",
							Type:        "json",
						},
					},
				}
			}
		}
	}

	return nil
}
