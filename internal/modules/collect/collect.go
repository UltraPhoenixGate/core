package collect

import (
	"encoding/json"
	"io"
	"net/http"
	"sync"
	"time"
	"ultraphx-core/internal/hub"
	"ultraphx-core/internal/models"

	"github.com/sirupsen/logrus"
)

type failureTracker struct {
	mu          sync.Mutex
	failures    map[string]int
	suspendTime map[string]time.Time
}

func newFailureTracker() *failureTracker {
	return &failureTracker{
		failures:    make(map[string]int),
		suspendTime: make(map[string]time.Time),
	}
}

func (ft *failureTracker) recordFailure(endpoint string) {
	ft.mu.Lock()
	defer ft.mu.Unlock()
	ft.failures[endpoint]++
	if ft.failures[endpoint] >= 3 {
		ft.suspendTime[endpoint] = time.Now().Add(2 * time.Minute)
	}
}

func (ft *failureTracker) canRequest(endpoint string) bool {
	ft.mu.Lock()
	defer ft.mu.Unlock()
	if suspendUntil, ok := ft.suspendTime[endpoint]; ok {
		if time.Now().Before(suspendUntil) {
			return false
		}
		delete(ft.failures, endpoint)
		delete(ft.suspendTime, endpoint)
	}
	return true
}

var tracker = newFailureTracker()

func runTick(h *hub.Hub) {
	allClients := make([]models.Client, 0)
	(&models.Client{}).Query().Where("status = ?", models.ClientStatusActive).Where("type = ?", models.ClientTypeSensorActive).Preload("Collection").Find(&allClients)
	for _, client := range allClients {
		go runCollect(&client, h)
	}
}

func Setup(h *hub.Hub) {
	ticker := time.NewTicker(1 * time.Second)
	go func() {
		for range ticker.C {
			runTick(h)
		}
	}()
}

func runCollect(client *models.Client, h *hub.Hub) {
	if client.Collection == nil {
		return
	}
	collection := client.Collection

	if !tracker.canRequest(collection.CollectionEndpoint) {
		logrus.Infof("Collection endpoint %s is suspended due to multiple failures", collection.CollectionEndpoint)
		return
	}

	if (collection.LastCollectionTime.Add(time.Duration(collection.CollectionPeriod) * time.Second)).After(time.Now()) {
		return
	}

	data, err := pullData(collection)
	if err != nil {
		tracker.recordFailure(collection.CollectionEndpoint)
		logrus.Errorf("Failed to pull data from collection %s: %s", collection.CollectionEndpoint, err)
		return
	}

	pushData(client, data, h)
}

type PullDataResult struct {
	Data   map[string]float64
	Labels map[string]string
}

func pullData(collection *models.CollectionInfo) (result PullDataResult, err error) {
	httpClient := &http.Client{
		Timeout: 5 * time.Second,
	}
	req, err := http.NewRequest("GET", collection.CollectionEndpoint, nil)
	if err != nil {
		return PullDataResult{}, err
	}
	req.Header.Set("Authorization", collection.AuthToken)
	resp, err := httpClient.Do(req)
	if err != nil {
		return PullDataResult{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return PullDataResult{}, err
	}

	var data PullDataResult
	bodyData, _ := io.ReadAll(resp.Body)
	if collection.DataType == models.CollectionDataTypeJSON {
		if err := json.Unmarshal(bodyData, &data); err != nil {
			return PullDataResult{}, err
		}
	}
	return data, nil
}

func pushData(client *models.Client, data PullDataResult, h *hub.Hub) {
	h.Broadcast(&hub.Message{
		Topic: "data",
		Payload: map[string]interface{}{
			"senderID": client.ID,
			"data":     data.Data,
		},
	})
}
