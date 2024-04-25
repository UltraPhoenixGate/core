package hub

import "github.com/google/uuid"

type ListenerCallback func(msg *Message)

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

func handleBroadcast(msg *Message) {
	listeners := registeredListeners[msg.Topic]
	for _, listener := range listeners {
		listener.Callback(msg)
	}

	// broadcast to # listeners
	listeners = registeredListeners["#"]
	for _, listener := range listeners {
		listener.Callback(msg)
	}
}
