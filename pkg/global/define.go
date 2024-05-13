package global

import "github.com/mitchellh/mapstructure"

type SensorData = map[string]float64

type SensorEventPayload struct {
	SenderID  string `json:"senderID" mapstructure:"senderID"`
	EventName string `json:"eventName" mapstructure:"eventName"`
}

func ParseSensorDataPayload(payload map[string]interface{}) *SensorDataPayload {
	payloadObj := &SensorDataPayload{}
	mapstructure.Decode(payload, payloadObj)
	return payloadObj
}

type SensorDataPayload struct {
	SenderID string     `json:"senderID" mapstructure:"senderID"`
	Data     SensorData `json:"data" mapstructure:"data"`
}

func ParseSensorEventPayload(payload map[string]interface{}) *SensorEventPayload {
	payloadObj := &SensorEventPayload{}
	mapstructure.Decode(payload, payloadObj)
	return payloadObj
}

type AlertPayload struct {
	ClientID string `json:"clientID" mapstructure:"clientID"`
	RuleName string `json:"ruleName" mapstructure:"ruleName"`
	Level    string `json:"level" mapstructure:"level"`
}

func ParseAlertPayload(payload map[string]interface{}) *AlertPayload {
	payloadObj := &AlertPayload{}
	mapstructure.Decode(payload, payloadObj)
	return payloadObj
}

func ToMap(data interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	mapstructure.Decode(data, &result)
	return result
}
