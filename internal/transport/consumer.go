package transport

import "context"

type MessageHandler interface {
	// Handle a Message produced by the broker.
	Handle(ctx context.Context, message *Message)
}

type MessageHandlerFunc func(ctx context.Context, message *Message)

func (h MessageHandlerFunc) Handle(ctx context.Context, message *Message) {
	h(ctx, message)
}

type Consumer interface {
	// Connect establishes a connection to the broker and subscribes to receive messages
	//
	// Returns either a Connection on success or an error.
	Consume(ctx context.Context, topic string, handler MessageHandler) (Connection, error)
}

type Connection interface {
	// Disconnect drops the previously established connection to the broker.
	Disconnect(ctx context.Context) error
}
