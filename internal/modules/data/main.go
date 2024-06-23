package data

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"
	"ultraphx-core/internal/config"
	"ultraphx-core/internal/hub"
	"ultraphx-core/internal/models"
	"ultraphx-core/internal/router"
	"ultraphx-core/pkg/global"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func ConvertToTimeSeries(rawData global.SensorData, meta map[string]string) (string, error) {
	if len(rawData) == 0 {
		return "", fmt.Errorf("rawData cannot be empty")
	}

	var result strings.Builder

	// Convert meta map to a formatted string of labels for Prometheus
	var labels []string
	for key, value := range meta {
		labels = append(labels, fmt.Sprintf("%s=\"%s\"", key, value))
	}
	labelsString := "{" + strings.Join(labels, ",") + "}"

	// Format each sensor data point as a Prometheus time series
	for metric, value := range rawData {
		result.WriteString(fmt.Sprintf("%s%s %f\n", metric, labelsString, value))
	}

	return result.String(), nil
}

func writeToVM(data string) {
	vmURL := config.GetVmDBConfig().Url + "/api/v1/import/prometheus"
	client := http.Client{
		Timeout: 5 * time.Second,
	}
	logrus.Info("Writing to VM", data)
	req, err := http.NewRequest("POST", vmURL, bytes.NewBufferString(data))
	if err != nil {
		logrus.WithError(err).Error("Failed to create request")
	}

	req.Header.Set("Content-Type", "text/plain")
	resp, err := client.Do(req)
	if err != nil {
		logrus.WithError(err).Error("Failed to send request")
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNoContent {
		logrus.Info("Data written to VM")
	} else {
		logrus.WithField("status_code", resp.StatusCode).Error("Failed to write to VM")
	}
}

func handleDataListener(h *hub.Hub, msg *hub.Message) {
	logrus.Info("Data message received", msg)
	// handle data message
	payload := global.ParseSensorDataPayload(msg.Payload)

	client := models.Client{
		ID: payload.SenderID,
	}
	if err := client.Query().Find(&client).Error; err != nil {
		logrus.WithError(err).Error("Failed to find client")
		return
	}
	meta := map[string]string{
		"sensor_id": payload.SenderID,
		"name":      client.Name,
	}

	timeSeriesData, err := ConvertToTimeSeries(payload.Data, meta)
	if err != nil {
		logrus.WithError(err).Error("Failed to convert to time series")
		return
	}

	// send to vm
	writeToVM(timeSeriesData)
}

func Setup() {
	hub.AddTopicListener("data::#", handleDataListener)
	authRouter := router.GetAuthRouter()
	vmUrl, err := url.Parse(config.GetVmDBConfig().Url)
	if err != nil {
		logrus.WithError(err).Error("Failed to parse VM URL")
		return
	}
	proxy := httputil.NewSingleHostReverseProxy(vmUrl)
	// Proxy /vmdb/* to VMDB
	authRouter.Any("/vmdb/*path", func(c *gin.Context) {
		c.Request.URL.Path = c.Param("path")
		// 删除默认响应头
		// for k := range c.Writer.Header() {
		// 	c.Writer.Header().Del(k)
		// }
		c.Writer.Header().Del("Access-Control-Allow-Origin")
		proxy.ServeHTTP(c.Writer, c.Request)
	})

	logrus.Info("Data module ready")
}
