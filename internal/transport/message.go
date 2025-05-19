package transport

import "encoding/json"

const (
	MessageTypeUserRegistered = "UserRegistered"
)

type Message struct {
	ID   string          `json:"id"`
	Type string          `json:"type"`
	Data json.RawMessage `json:"data,omitempty"`
}
