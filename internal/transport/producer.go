package transport

import "context"

// Producer defines the contract for sending messages to other services.
type Producer interface {
	Produce(ctx context.Context, topic string, message *Message) error
}
