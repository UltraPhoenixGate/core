package global

type SensorData = map[string]float64

type SensorPayload struct {
	SensorID string     `json:"sensor_id"`
	Data     SensorData `json:"data"`
}

type SensorEvent struct {
	SensorID  string                 `json:"sensor_id"`
	EventName string                 `json:"event_name"`
	Data      map[string]interface{} `json:"data"`
}
