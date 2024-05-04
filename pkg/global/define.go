package global

type SensorData = map[string]float64

type SensorEventData struct {
	EventName string `json:"event_name"`
}

type SensorPayload struct {
	SensorID string            `json:"sensor_id"`
	Type     SensorPayloadType `json:"type"`
	Data     any               `json:"data"`
}

type SensorPayloadType string

const (
	SensorPayloadTypeData  SensorPayloadType = "data"
	SensorPayloadTypeEvent SensorPayloadType = "event"
)

type AlertPayload struct {
	SensorID string
	RuleName string
	Level    string
}
