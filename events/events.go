package events

import "context"

// HandlerFunc processes an event message.
type HandlerFunc func(ctx context.Context, msg []byte) error

// Publisher publishes events to topics.
type Publisher interface {
	Publish(ctx context.Context, topic string, msg []byte) error
}

// Subscriber subscribes to topics and processes events.
type Subscriber interface {
	Subscribe(ctx context.Context, topic string, handler HandlerFunc) error
}
