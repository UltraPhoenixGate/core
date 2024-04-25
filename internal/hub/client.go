package hub

import "github.com/google/uuid"

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

func (c *Client) Send(msg *Message) {
	c.SendChan <- msg
}

func (c *Client) Subscribe(topic string) {
	c.Topics[topic] = true
}

func (c *Client) Unsubscribe(topic string) {
	delete(c.Topics, topic)
}
