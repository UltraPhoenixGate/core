package hub

import (
	"strings"

	"github.com/sirupsen/logrus"
)

var MAX_REGISTER_CHANNEL_SIZE = 100
var MAX_UNREGISTER_CHANNEL_SIZE = 100
var MAX_BROADCAST_CHANNEL_SIZE = 100

type Hub struct {
	clientMap  map[string]*Client
	register   chan *Client
	unregister chan *Client
	broadcast  chan *Message
}

func NewHub() *Hub {
	return &Hub{
		clientMap:  make(map[string]*Client),
		register:   make(chan *Client, MAX_REGISTER_CHANNEL_SIZE),
		unregister: make(chan *Client, MAX_UNREGISTER_CHANNEL_SIZE),
		broadcast:  make(chan *Message, MAX_BROADCAST_CHANNEL_SIZE),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clientMap[client.ID] = client
		case client := <-h.unregister:
			if _, ok := h.clientMap[client.ID]; ok {
				delete(h.clientMap, client.ID)
				close(client.SendChan)
			}
		case message := <-h.broadcast:
			logrus.Debug("broadcast message: ", message.ToJson())
			h.handleBroadcast(message)
		}
	}
}

func (h *Hub) handleBroadcast(message *Message) {
	for _, client := range h.clientMap {
		if isValidTopicMatch(message.Topic, client) {
			select {
			case client.SendChan <- message:
			default:
				h.closeClient(client)
			}
		}
	}
}

func isValidTopicMatch(topic string, client *Client) bool {
	if client.Topics["#"] {
		return true
	}

	if client.Topics[topic] {
		return true
	}

	// Check for wildcard subscriptions
	for subscribedTopic := range client.Topics {
		if strings.HasSuffix(subscribedTopic, "#") {
			baseTopic := strings.TrimSuffix(subscribedTopic, "#")
			if strings.HasPrefix(topic, baseTopic) {
				return true
			}
		}
	}

	return false
}

func (h *Hub) closeClient(client *Client) {
	close(client.SendChan)
	if !isClosed(client.SendChan) {
		close(client.SendChan)
	}
	delete(h.clientMap, client.ID)
}

func (h *Hub) Register(client *Client) {
	select {
	case h.register <- client:
	default:
		logrus.Printf("register channel is full, client %v discarded", client.ID)
	}
}

func (h *Hub) Unregister(client *Client) {
	select {
	case h.unregister <- client:
	default:
		logrus.Printf("unregister channel is full, client %v discarded", client.ID)
	}
}

func (h *Hub) Broadcast(message *Message) {
	go handleBroadcastListener(message) // not block
	select {
	case h.broadcast <- message:
	default:
		logrus.Printf("broadcast channel is full, message %v discarded", message)
	}
}

func isClosed(ch <-chan *Message) bool {
	select {
	case <-ch:
		return true
	default:
	}

	return false
}
