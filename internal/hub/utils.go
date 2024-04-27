package hub

import "strings"

func GetTopicPermission(topic string) string {
	if topic == "" {
		return ""
	}
	if strings.Contains(topic, "::") {
		topics := strings.Split(topic, "::")
		return topics[0]
	}
	return topic
}
