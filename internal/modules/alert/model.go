package alert

type AlertType string

const (
	AlertTypeWarning AlertType = "warning"
	AlertTypeError   AlertType = "error"
)

type AlertRule struct {
	Type        AlertRuleType
	Name        string
	Summary     string
	Description string
	Level       AlertType
	Conditions  []AlertRuleCondition
	Actions     []AlertAction
}

type AlertRuleType string

const (
	AlertRuleTypeRealtime AlertRuleType = "realtime"
	AlertRuleTypeStatic   AlertRuleType = "static"
)

type AlertRuleCondition struct {
	SensorID string
	Metric   string
	Type     AlertRuleConditionType
	Payload  any
}

type AlertRuleConditionType string

const (
	AlertRuleConditionTypeOperator AlertRuleConditionType = "operator"
	AlertRuleConditionTypeEvent    AlertRuleConditionType = "event"
)

type AlertRuleConditionPayloadOperator struct {
	Operator AlertRuleConditionOperator
	Value    float64
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
