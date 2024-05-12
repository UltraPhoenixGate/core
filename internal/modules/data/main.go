package data

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"
	"time"
	"ultraphx-core/internal/config"
	"ultraphx-core/internal/hub"
	"ultraphx-core/internal/models"
	"ultraphx-core/pkg/global"

	"github.com/sirupsen/logrus"
)

func ConvertToTimeSeries(rawData global.SensorData, meta map[string]string) (string, error) {
	if len(rawData) == 0 {
		return "", fmt.Errorf("rawData cannot be empty")
	}

	timestamp := time.Now().Unix()

	var timeSeriesData []string
	for metricName, metricValue := range rawData {
		var labels []string
		for key, value := range meta {
			label := fmt.Sprintf("%s=\"%s\"", key, strings.ReplaceAll(value, "\"", "\\\""))
			labels = append(labels, label)
		}
		labelsStr := "{" + strings.Join(labels, ",") + "}"

		metricData := fmt.Sprintf("%s%s %f %d", metricName, labelsStr, metricValue, timestamp)
		timeSeriesData = append(timeSeriesData, metricData)
	}
	result := strings.Join(timeSeriesData, "\n")

	return result, nil
}

func writeToVM(data string) {
	vmURL := config.GetVmDBConfig().Url + "api/v1/write"
	client := http.Client{
		Timeout: 5 * time.Second,
	}
	req, err := http.NewRequest("POST", vmURL, bytes.NewBufferString(data))
	if err != nil {
		logrus.WithError(err).Error("Failed to create request")
	}

	req.Header.Set("Content-Type", "text/plain; version=0.0.4")
	resp, err := client.Do(req)
	if err != nil {
		logrus.WithError(err).Error("Failed to send request")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logrus.WithField("status_code", resp.StatusCode).Error("Failed to write to vm")
	}
}

func handleDataListener(h *hub.Hub, msg *hub.Message) {
	logrus.Info("Data message received", msg)
	// handle data message
	payload := msg.Payload.(global.SensorPayload)
	if payload.Type != global.SensorPayloadTypeData {
		return
	}

	client := models.Client{
		ID: payload.SensorID,
	}
	if err := client.Query().Find(&client).Error; err != nil {
		logrus.WithError(err).Error("Failed to find client")
		return
	}
	meta := map[string]string{
		"sensor_id": payload.SensorID,
		"name":      client.Name,
	}

	timeSeriesData, err := ConvertToTimeSeries(payload.Data.(global.SensorData), meta)
	if err != nil {
		logrus.WithError(err).Error("Failed to convert to time series")
		return
	}

	// send to vm
	writeToVM(timeSeriesData)
}

func Setup() {
	hub.AddTopicListener("data::#", handleDataListener)
	logrus.Info("Data module ready")
}
