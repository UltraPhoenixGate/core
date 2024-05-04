package alert

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

const baseConfigPath = "config/alert/rules"

var rules = []AlertRule{}

func init() {
	RefreshRules()
}

func GetRules() []AlertRule {
	return rules
}

func RefreshRules() {

	if _, err := os.Stat(baseConfigPath); os.IsNotExist(err) {
		if err := os.MkdirAll(baseConfigPath, 0755); err != nil {
			logrus.WithError(err).Error("Failed to create alert rules directory")
		}
	}
	ruleFiles, err := os.ReadDir(baseConfigPath)
	if err != nil {
		logrus.WithError(err).Error("Failed to read alert rules directory")
	}

	rules = []AlertRule{}
	for _, ruleFile := range ruleFiles {
		if ruleFile.IsDir() {
			continue
		}
		rule, err := LoadRule(ruleFile.Name())
		if err != nil {
			logrus.WithError(err).Error("Failed to load alert rule")
			continue
		}
		rules = append(rules, rule)
	}
}

func LoadRule(name string) (AlertRule, error) {
	filePath := filepath.Join(baseConfigPath, name)
	ruleFile, err := os.Open(filePath)
	if err != nil {
		return AlertRule{}, err
	}
	defer ruleFile.Close()

	rule := AlertRule{}
	if err := json.NewDecoder(ruleFile).Decode(&rule); err != nil {
		return AlertRule{}, err
	}

	return rule, nil
}

func AddRule(rule AlertRule) error {
	// if rule already exists, return error
	for _, r := range rules {
		if r.Name == rule.Name {
			return fmt.Errorf("rule %s already exists", rule.Name)
		}
	}

	filePath := filepath.Join(baseConfigPath, fmt.Sprintf("%s.json", rule.Name))
	ruleFile, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer ruleFile.Close()

	if err := json.NewEncoder(ruleFile).Encode(&rule); err != nil {
		return err
	}

	rules = append(rules, rule)
	return nil
}
