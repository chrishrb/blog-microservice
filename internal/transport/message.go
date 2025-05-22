package transport

import "encoding/json"

type Message struct {
	ID   string          `json:"id"`
	Type MessageType     `json:"type"`
	Data json.RawMessage `json:"data,omitempty"`
}

type MessageType string

var (
	MessageTypeUserRegistered MessageType = "UserRegistered"
)

const PasswordResetTopic = "password-reset"

type PasswordResetEvent struct {
	Email     string `json:"email"`
	Token     string `json:"token"`
	ExpiresIn int    `json:"expires_in"`
}
