package alert

import (
	"net/http"
	"time"
	"ultraphx-core/pkg/resp"
	"ultraphx-core/pkg/validator"
)

func GetAlertRules(w http.ResponseWriter, r *http.Request) {
	resp.OK(w, resp.H{
		"rules": rules,
	})
}

func AddAlertRule(w http.ResponseWriter, r *http.Request) {
	var rule AlertRule
	if err := validator.ShouldBind(r, &rule); err != nil {
		resp.Error(w, "Invalid request")
		return
	}

	if err := AddRule(&rule); err != nil {
		resp.Error(w, err.Error())
		return
	}

	resp.OK(w, resp.H{
		"rule": rule,
	})
}

func GetRule(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name == "" {
		resp.Error(w, "Invalid request")
		return
	}

	for _, rule := range rules {
		if rule.Name == name {
			resp.OK(w, resp.H{
				"rule": rule,
			})
			return
		}
	}

	resp.Error(w, "Rule not found")
}

func GetAlertRecords(w http.ResponseWriter, r *http.Request) {
	var req struct {
		StartAt  time.Time `json:"start_at"`
		EndAt    time.Time `json:"end_at"`
		ClientID string    `json:"client_id"`
	}

	if err := validator.ShouldBind(r, &req); err != nil {
		resp.Error(w, "Invalid request")
		return
	}

	var records []*AlertRecord
	query := (&AlertRecord{}).Query()
	if !req.StartAt.IsZero() {
		query = query.Where("created_at >= ?", req.StartAt)
	}
	if !req.EndAt.IsZero() {
		query = query.Where("created_at <= ?", req.EndAt)
	}
	if req.ClientID != "" {
		query = query.Where("client_id = ?", req.ClientID)
	}
	if err := query.Find(&records).Error; err != nil {
		resp.Error(w, "Failed to get records")
		return
	}
	resp.OK(w, resp.H{
		"records": records,
	})
}
