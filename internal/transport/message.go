package transport

import "encoding/json"

type Message struct {
	ID   string          `json:"id"`
	Data json.RawMessage `json:"data,omitempty"`
}

const PasswordResetTopic = "password-reset"
const UserCreatedTopic = "user-created"

type PasswordResetEvent struct {
	Recipient string `json:"recipient"`
	Channel   string `json:"channel,oneof=email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Token     string `json:"token"`
}

type UserCreatedEvent struct {
	Email     string `json:"email"`
	Token     string `json:"token"`
	ExpiresIn int    `json:"expires_in"`
}
