package alert

import (
	"ultraphx-core/internal/models"

	"gorm.io/gorm"
)

// alert
type AlertRecord struct {
	models.Model
	ClientID string    `json:"clientID"`
	RuleName string    `json:"ruleName"`
	Level    AlertType `json:"level"`
}

func (a *AlertRecord) Query() *gorm.DB {
	return models.DB.Model(a)
}

type AlertType string

const (
	AlertTypeWarning AlertType = "warning"
	AlertTypeError   AlertType = "error"
)

type AlertRule struct {
	Type        AlertRuleType        `json:"type" validate:"required"`
	Name        string               `json:"name" validate:"required"`
	Summary     string               `json:"summary"`
	Description string               `json:"description"`
	Level       AlertType            `json:"level" validate:"required"`
	Conditions  []AlertRuleCondition `json:"conditions" validate:"required"`
	Actions     []AlertAction        `json:"actions"`
}

type AlertRuleType string

const (
	AlertRuleTypeRealtime AlertRuleType = "realtime"
	AlertRuleTypeStatic   AlertRuleType = "static"
)

type AlertRuleCondition struct {
	SensorID string                 `json:"sensorId" validate:"required"`
	Metric   string                 `json:"metric" validate:"required"`
	Type     AlertRuleConditionType `json:"type" validate:"required"`
	Payload  any                    `json:"payload" validate:"required"`
}

type AlertRuleConditionType string

const (
	AlertRuleConditionTypeOperator AlertRuleConditionType = "operator"
	AlertRuleConditionTypeEvent    AlertRuleConditionType = "event"
)

type AlertRuleConditionPayloadOperator struct {
	Operator AlertRuleConditionOperator `json:"operator" validate:"required"`
	Value    float64                    `json:"value" validate:"required"`
}

type AlertRuleConditionOperator string

const (
	AlertRuleConditionOperatorEqual       AlertRuleConditionOperator = "eq"
	AlertRuleConditionOperatorNotEqual    AlertRuleConditionOperator = "ne"
	AlertRuleConditionOperatorGreaterThan AlertRuleConditionOperator = "gt"
	AlertRuleConditionOperatorLessThan    AlertRuleConditionOperator = "lt"
)

type AlertRuleConditionPayloadEvent struct {
	EventName string
}

type AlertAction struct {
	Type    AlertActionType
	Payload any
}

type AlertActionType string

const (
	AlertActionTypeEmail   AlertActionType = "email"
	AlertActionTypeSMS     AlertActionType = "sms"
	AlertActionTypeWebhook AlertActionType = "webhook"
)
