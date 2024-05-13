package hub

import (
	"strings"

	"github.com/google/uuid"
)

type ListenerCallback func(h *Hub, msg *Message)

type ListenerItem struct {
	ID       string
	Topic    string
	Callback ListenerCallback
}

var registeredListeners = make(map[string][]*ListenerItem)

func AddTopicListener(topic string, callback ListenerCallback) string {
	id := uuid.New().String()
	registeredListeners[topic] = append(registeredListeners[topic], &ListenerItem{
		ID:       id,
		Topic:    topic,
		Callback: callback,
	})

	return id
}

func RemoveTopicListener(topic string, id string) {
	listeners := registeredListeners[topic]
	for i, listener := range listeners {
		if listener.ID == id {
			registeredListeners[topic] = append(listeners[:i], listeners[i+1:]...)
			break
		}
	}
}

func AddListener(callback ListenerCallback) string {
	return AddTopicListener("#", callback)
}

func RemoveListener(id string) {
	RemoveTopicListener("#", id)
}

func handleBroadcastListener(h *Hub, msg *Message) {
	for _, listeners := range registeredListeners {
		for _, listener := range listeners {
			if matchesTopic(listener.Topic, msg.Topic) {
				listener.Callback(h, msg)
			}
		}
	}
}

func matchesTopic(listenerTopic, messageTopic string) bool {
	if listenerTopic == "#" {
		return true
	}
	listenerParts := strings.Split(listenerTopic, "::")
	messageParts := strings.Split(messageTopic, "::")

	for i, part := range listenerParts {
		if part == "#" {
			return true
		}
		if i >= len(messageParts) || part != messageParts[i] {
			return false
		}
	}
	return len(listenerParts) == len(messageParts)
}
