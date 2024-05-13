package hub

import (
	"github.com/google/uuid"
)

type PermissionType int

const (
	PermissionTypeRead PermissionType = iota + 1
	PermissionTypeWrite
)

type Client struct {
	ID       string
	SendChan chan *Message
	Topics   map[string]bool
	Hub      *Hub
}

func NewClient(id string, hub *Hub) *Client {
	if id == "" {
		id = uuid.New().String()
	}
	return &Client{
		ID:       id,
		SendChan: make(chan *Message),
		Topics:   make(map[string]bool),
		Hub:      hub,
	}
}

// 向客户端发送消息
func (c *Client) Send(msg *Message) {
	c.SendChan <- msg
}

// 广播消息
func (c *Client) Broadcast(msg *Message) {
	c.Hub.Broadcast(msg)
}

func (c *Client) Subscribe(topic string) {
	c.Topics[topic] = true
}

func (c *Client) Unsubscribe(topic string) {
	delete(c.Topics, topic)
}
