package hub

import "encoding/json"

type Message struct {
	Topic   string
	Payload map[string]interface{}
}

func (m *Message) ToJson() []byte {
	data, _ := json.Marshal(m)
	return data
}

func PraseMessageStr(data string) (*Message, error) {
	msg := &Message{}
	error := json.Unmarshal([]byte(data), msg)
	return msg, error
}

func PraseMessageByte(data []byte) (*Message, error) {
	msg := &Message{}
	error := json.Unmarshal(data, msg)
	return msg, error
}
