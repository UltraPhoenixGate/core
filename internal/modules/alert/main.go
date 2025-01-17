package alert

import (
	"ultraphx-core/internal/hub"
	"ultraphx-core/internal/models"
	"ultraphx-core/internal/router"
	"ultraphx-core/pkg/global"

	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
)

// only handle real-time alert rules
func handleAlertRT(h *hub.Hub, msg *hub.Message) {
	rules := GetRules()
	for _, rule := range rules {
		if rule.Type != AlertRuleTypeRealtime {
			continue
		}
		senderID := msg.Payload["senderID"].(string)

		for _, condition := range rule.Conditions {
			if condition.SensorID != senderID {
				continue
			}

			if isMatched(&condition, msg.Payload) {
				alert := AlertRecord{
					ClientID: senderID,
					RuleName: rule.Name,
					Summary:  rule.Summary,
					Level:    rule.Level,
				}
				if err := alert.Query().Create(&alert).Error; err != nil {
					logrus.WithError(err).Error("Failed to save alert record")
				}

				// Broadcast alert
				h.Broadcast(&hub.Message{
					Topic: "alert" + string(rule.Level),
					Payload: global.ToMap(global.AlertPayload{
						ClientID: senderID,
						RuleName: rule.Name,
						Level:    string(rule.Level),
					}),
				})

				go processAlertActions(rule, msg.Payload)
			}
		}
	}
}

func processAlertActions(rule *AlertRule, payload map[string]interface{}) {
	senderID := payload["senderID"].(string)
	for _, action := range rule.Actions {
		switch action.Type {
		case AlertActionTypeEmail:
			email := AlertActionPayloadEmail{}
			mapstructure.Decode(action.Payload, &email)
			// send email
			logrus.Infof("Send email to %s: %s", email.To, "Alert"+senderID)
		case AlertActionTypeSMS:
			sms := AlertActionPayloadSMS{}
			mapstructure.Decode(action.Payload, &sms)
			// send sms
			logrus.Infof("Send sms to %s: %s", sms.To, "Alert"+senderID)
		case AlertActionTypeWebhook:
			webhook := AlertActionPayloadWebhook{}
			mapstructure.Decode(action.Payload, &webhook)
			// send webhook
			logrus.Infof("Send webhook to %s: %s", webhook.URL, "Alert"+senderID)
		}
	}
}

func isMatched(condition *AlertRuleCondition, payload map[string]interface{}) bool {
	if condition.Type == AlertRuleConditionTypeOperator {
		sensorData := global.ParseSensorDataPayload(payload).Data
		// for operator type, check if the metric value satisfies the condition
		if _, ok := sensorData[condition.Metric]; !ok {
			return false
		}
		operator := AlertRuleConditionPayloadOperator{}
		mapstructure.Decode(condition.Payload, &operator)
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
		eventType := AlertRuleConditionPayloadEvent{}
		mapstructure.Decode(condition.Payload, &eventType)
		return global.ParseSensorEventPayload(payload).EventName == eventType.EventName
	}
	return false
}

func Setup() {
	hub.AddTopicListener("data::#", handleAlertRT)

	// migrate
	models.AutoMigrate(&AlertRecord{})

	authRouter := router.GetAuthRouter()
	authRouter.GET("/alert/rules", GetAlertRules)
	authRouter.POST("/alert/rule", AddAlertRule)
	authRouter.GET("/alert/rule", GetAlertRule)
	authRouter.PUT("/alert/rule", UpdateAlertRule)
	authRouter.DELETE("/alert/rule", DeleteAlertRule)

	authRouter.GET("/alert/records", GetAlertRecords)
	logrus.Info("Alert module ready")
}
