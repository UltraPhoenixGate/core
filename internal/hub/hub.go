package hub

import "github.com/sirupsen/logrus"

type Hub struct {
	clientMap  map[string]*Client
	register   chan *Client
	unregister chan *Client
	broadcast  chan *Message
}

func NewHub() *Hub {
	return &Hub{
		clientMap:  make(map[string]*Client),
		register:   make(chan *Client, 100),  // 设置缓冲区大小为100
		unregister: make(chan *Client, 100),  // 设置缓冲区大小为100
		broadcast:  make(chan *Message, 100), // 设置缓冲区大小为100
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
				if client.Topics[message.Type] {
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
