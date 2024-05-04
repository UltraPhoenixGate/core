package alert

import (
	"ultraphx-core/internal/hub"
	"ultraphx-core/pkg/global"

	"github.com/sirupsen/logrus"
)

// only handle real-time alert rules
func handleAlertRT(h *hub.Hub, msg *hub.Message) {
	rules := GetRules()
	payload := msg.Payload.(global.SensorPayload)
	for _, rule := range rules {
		if rule.Type != AlertRuleTypeRealtime {
			continue
		}

		for _, condition := range rule.Conditions {
			if condition.SensorID != payload.SensorID {
				continue
			}

			if isMatched(&condition, &payload) {
				// TODO: save alert to database

				// Broadcast alert
				h.Broadcast(&hub.Message{
					Topic: "alert" + string(rule.Level),
					Payload: global.AlertPayload{
						SensorID: payload.SensorID,
						RuleName: rule.Name,
						Level:    string(rule.Level),
					},
				})
			}
		}
	}
}

func isMatched(condition *AlertRuleCondition, payload *global.SensorPayload) bool {
	if condition.Type == AlertRuleConditionTypeOperator {
		sensorData := payload.Data.(global.SensorData)
		// for operator type, check if the metric value satisfies the condition
		if _, ok := sensorData[condition.Metric]; !ok {
			return false
		}

		operator := condition.Payload.(AlertRuleConditionPayloadOperator)
		value := sensorData[condition.Metric]
		switch operator.Operator {
		case AlertRuleConditionOperatorEqual:
			return value == operator.Value
		case AlertRuleConditionOperatorNotEqual:
			return value != operator.Value
		case AlertRuleConditionOperatorGreaterThan:
			return value > operator.Value
		case AlertRuleConditionOperatorLessThan:
			return value < operator.Value

		default:
			return false
		}
	}

	if condition.Type == AlertRuleConditionTypeEvent {
		// for event type, check if the event name matches
		eventName := condition.Payload.(AlertRuleConditionPayloadEvent).EventName
		return payload.Data.(global.SensorEventData).EventName == eventName
	}
	return false
}

func Setup() {
	hub.AddTopicListener("data::#", handleAlertRT)
	logrus.Info("Alert module ready")
}
