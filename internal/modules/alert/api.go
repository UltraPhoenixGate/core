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
	startAt := time.Time{}
	endAt := time.Time{}
	if startAtStr := r.URL.Query().Get("start_at"); startAtStr != "" {
		if err := startAt.UnmarshalText([]byte(startAtStr)); err != nil {
			resp.Error(w, "Invalid start_at")
			return
		}
	}
	if endAtStr := r.URL.Query().Get("end_at"); endAtStr != "" {
		if err := endAt.UnmarshalText([]byte(endAtStr)); err != nil {
			resp.Error(w, "Invalid end_at")
			return
		}
	}
	clientID := r.URL.Query().Get("client_id")

	var records []*AlertRecord
	query := (&AlertRecord{}).Query()
	if !startAt.IsZero() {
		query = query.Where("created_at >= ?", startAt)
	}
	if !startAt.IsZero() {
		query = query.Where("created_at <= ?", endAt)
	}
	if clientID != "" {
		query = query.Where("client_id = ?", clientID)
	}
	if err := query.Find(&records).Error; err != nil {
		resp.Error(w, "Failed to get records")
		return
	}
	resp.OK(w, resp.H{
		"records": records,
	})
}
