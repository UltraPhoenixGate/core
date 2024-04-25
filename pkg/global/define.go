package global

type SensorData struct {
	SensorID string                 `json:"sensor_id"`
	Data     map[string]interface{} `json:"data"`
}

type SensorEvent struct {
	SensorID  string                 `json:"sensor_id"`
	EventName string                 `json:"event_name"`
	Data      map[string]interface{} `json:"data"`
}
