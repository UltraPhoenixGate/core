package hub

import "github.com/sirupsen/logrus"

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
			for _, client := range h.clientMap {
				// # 通配符表示订阅所有主题
				if client.Topics[message.Topic] || client.Topics["#"] {
					select {
					case client.SendChan <- message:
					default:
						close(client.SendChan) // 关闭通道前先检查是否已关闭，避免重复关闭
						if !isClosed(client.SendChan) {
							close(client.SendChan)
						}
						delete(h.clientMap, client.ID)
					}
				}
			}
		}
	}
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
	handleBroadcast(message)
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
