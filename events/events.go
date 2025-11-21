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

// StreamMessage represents a message from a persistent stream.
type StreamMessage struct {
	Data      []byte // The event payload
	Sequence  uint64 // Monotonic sequence number
	Timestamp int64  // Unix timestamp (nanoseconds)
}

// StreamConsumer consumes messages from a persistent event stream.
// This is an abstraction over JetStream, Kafka, Pulsar, etc.
type StreamConsumer interface {
	// Fetch retrieves up to maxMessages from the stream.
	// Messages are returned in order from oldest to newest.
	// If limit is 0, fetches all available messages.
	Fetch(ctx context.Context, limit int) ([]StreamMessage, error)

	// SubscribeStream to new messages arriving on the stream (real-time).
	// This is different from Subscriber.Subscribe - it's for persistent streams only.
	SubscribeStream(ctx context.Context, handler HandlerFunc) error
}

// Stream provides both publishing and consuming with persistence.
type Stream interface {
	Publisher
	StreamConsumer
}
