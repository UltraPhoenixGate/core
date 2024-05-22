package alert

import (
	"time"
	"ultraphx-core/pkg/resp"

	"github.com/gin-gonic/gin"
)

func GetAlertRules(c *gin.Context) {
	resp.OK(c, resp.H{
		"rules": rules,
	})
}

func AddAlertRule(c *gin.Context) {
	var rule AlertRule
	if err := c.ShouldBindJSON(&rule); err != nil {
		resp.Error(c, "Invalid request")
		return
	}

	if err := AddRule(&rule); err != nil {
		resp.Error(c, err.Error())
		return
	}

	resp.OK(c, resp.H{
		"rule": rule,
	})
}

func GetAlertRule(c *gin.Context) {
	name := c.Query("name")
	if name == "" {
		resp.Error(c, "Invalid request")
		return
	}

	for _, rule := range rules {
		if rule.Name == name {
			resp.OK(c, resp.H{
				"rule": rule,
			})
			return
		}
	}

	resp.Error(c, "Rule not found")
}

func UpdateAlertRule(c *gin.Context) {
	var rule AlertRule
	if err := c.ShouldBindJSON(&rule); err != nil {
		resp.Error(c, "Invalid request")
		return
	}

	if err := UpdateRule(&rule); err != nil {
		resp.Error(c, err.Error())
		return
	}

	resp.OK(c, resp.H{
		"rule": rule,
	})
}

func DeleteAlertRule(c *gin.Context) {
	name := c.Query("name")
	if name == "" {
		resp.Error(c, "Invalid request")
		return
	}

	if err := DeleteRule(name); err != nil {
		resp.Error(c, err.Error())
		return
	}

	resp.OK(c, resp.H{})
}

func GetAlertRecords(c *gin.Context) {
	startAt := time.Time{}
	endAt := time.Time{}
	if startAtStr := c.Query("start_at"); startAtStr != "" {
		if err := startAt.UnmarshalText([]byte(startAtStr)); err != nil {
			resp.Error(c, "Invalid start_at")
			return
		}
	}
	if endAtStr := c.Query("end_at"); endAtStr != "" {
		if err := endAt.UnmarshalText([]byte(endAtStr)); err != nil {
			resp.Error(c, "Invalid end_at")
			return
		}
	}
	clientID := c.Query("client_id")

	var records []*AlertRecord
	query := (&AlertRecord{}).Query()
	if !startAt.IsZero() {
		query = query.Where("created_at >= ?", startAt)
	}
	if !endAt.IsZero() {
		query = query.Where("created_at <= ?", endAt)
	}
	if clientID != "" {
		query = query.Where("client_id = ?", clientID)
	}
	// Preload the client
	query = query.Preload("Client")
	query = query.Order("created_at DESC")
	if err := query.Find(&records).Error; err != nil {
		resp.Error(c, "Failed to get records")
		return
	}
	resp.OK(c, records)
}
