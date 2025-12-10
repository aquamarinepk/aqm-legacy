package events

import (
	"context"
	"errors"
	"testing"
)

func TestHandlerFuncType(t *testing.T) {
	called := false
	var handler HandlerFunc = func(ctx context.Context, msg []byte) error {
		called = true
		return nil
	}

	err := handler(context.Background(), []byte("test"))
	if err != nil {
		t.Errorf("handler error: %v", err)
	}
	if !called {
		t.Error("handler should have been called")
	}
}

func TestHandlerFuncReturnsError(t *testing.T) {
	expectedErr := errors.New("handler error")
	var handler HandlerFunc = func(ctx context.Context, msg []byte) error {
		return expectedErr
	}

	err := handler(context.Background(), []byte("test"))
	if err != expectedErr {
		t.Errorf("error = %v, want %v", err, expectedErr)
	}
}

func TestStreamMessageStruct(t *testing.T) {
	msg := StreamMessage{
		Data:      []byte("test data"),
		Sequence:  42,
		Timestamp: 1234567890,
	}

	if string(msg.Data) != "test data" {
		t.Errorf("Data = %s, want test data", msg.Data)
	}
	if msg.Sequence != 42 {
		t.Errorf("Sequence = %d, want 42", msg.Sequence)
	}
	if msg.Timestamp != 1234567890 {
		t.Errorf("Timestamp = %d, want 1234567890", msg.Timestamp)
	}
}

func TestStreamMessageEmptyData(t *testing.T) {
	msg := StreamMessage{
		Data:      nil,
		Sequence:  0,
		Timestamp: 0,
	}

	if msg.Data != nil {
		t.Error("Data should be nil")
	}
}

// MockPublisher for testing Publisher interface
type mockPublisher struct {
	published []struct {
		topic string
		msg   []byte
	}
	err error
}

func (m *mockPublisher) Publish(ctx context.Context, topic string, msg []byte) error {
	if m.err != nil {
		return m.err
	}
	m.published = append(m.published, struct {
		topic string
		msg   []byte
	}{topic, msg})
	return nil
}

func TestPublisherInterface(t *testing.T) {
	var pub Publisher = &mockPublisher{}

	err := pub.Publish(context.Background(), "test-topic", []byte("test message"))
	if err != nil {
		t.Errorf("Publish error: %v", err)
	}
}

func TestPublisherInterfaceError(t *testing.T) {
	expectedErr := errors.New("publish failed")
	var pub Publisher = &mockPublisher{err: expectedErr}

	err := pub.Publish(context.Background(), "test-topic", []byte("test"))
	if err != expectedErr {
		t.Errorf("error = %v, want %v", err, expectedErr)
	}
}

// MockSubscriber for testing Subscriber interface
type mockSubscriber struct {
	subscribed []struct {
		topic   string
		handler HandlerFunc
	}
	err error
}

func (m *mockSubscriber) Subscribe(ctx context.Context, topic string, handler HandlerFunc) error {
	if m.err != nil {
		return m.err
	}
	m.subscribed = append(m.subscribed, struct {
		topic   string
		handler HandlerFunc
	}{topic, handler})
	return nil
}

func TestSubscriberInterface(t *testing.T) {
	var sub Subscriber = &mockSubscriber{}

	handler := func(ctx context.Context, msg []byte) error { return nil }
	err := sub.Subscribe(context.Background(), "test-topic", handler)
	if err != nil {
		t.Errorf("Subscribe error: %v", err)
	}
}

func TestSubscriberInterfaceError(t *testing.T) {
	expectedErr := errors.New("subscribe failed")
	var sub Subscriber = &mockSubscriber{err: expectedErr}

	err := sub.Subscribe(context.Background(), "test-topic", nil)
	if err != expectedErr {
		t.Errorf("error = %v, want %v", err, expectedErr)
	}
}

// MockStreamConsumer for testing StreamConsumer interface
type mockStreamConsumer struct {
	messages []StreamMessage
	err      error
}

func (m *mockStreamConsumer) Fetch(ctx context.Context, limit int) ([]StreamMessage, error) {
	if m.err != nil {
		return nil, m.err
	}
	if limit == 0 || limit > len(m.messages) {
		return m.messages, nil
	}
	return m.messages[:limit], nil
}

func (m *mockStreamConsumer) SubscribeStream(ctx context.Context, handler HandlerFunc) error {
	return m.err
}

func TestStreamConsumerFetch(t *testing.T) {
	messages := []StreamMessage{
		{Data: []byte("msg1"), Sequence: 1},
		{Data: []byte("msg2"), Sequence: 2},
		{Data: []byte("msg3"), Sequence: 3},
	}
	var consumer StreamConsumer = &mockStreamConsumer{messages: messages}

	result, err := consumer.Fetch(context.Background(), 0)
	if err != nil {
		t.Errorf("Fetch error: %v", err)
	}
	if len(result) != 3 {
		t.Errorf("len(result) = %d, want 3", len(result))
	}
}

func TestStreamConsumerFetchWithLimit(t *testing.T) {
	messages := []StreamMessage{
		{Data: []byte("msg1"), Sequence: 1},
		{Data: []byte("msg2"), Sequence: 2},
		{Data: []byte("msg3"), Sequence: 3},
	}
	var consumer StreamConsumer = &mockStreamConsumer{messages: messages}

	result, err := consumer.Fetch(context.Background(), 2)
	if err != nil {
		t.Errorf("Fetch error: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("len(result) = %d, want 2", len(result))
	}
}

func TestStreamConsumerFetchError(t *testing.T) {
	expectedErr := errors.New("fetch failed")
	var consumer StreamConsumer = &mockStreamConsumer{err: expectedErr}

	_, err := consumer.Fetch(context.Background(), 0)
	if err != expectedErr {
		t.Errorf("error = %v, want %v", err, expectedErr)
	}
}

func TestStreamConsumerSubscribeStream(t *testing.T) {
	var consumer StreamConsumer = &mockStreamConsumer{}

	handler := func(ctx context.Context, msg []byte) error { return nil }
	err := consumer.SubscribeStream(context.Background(), handler)
	if err != nil {
		t.Errorf("SubscribeStream error: %v", err)
	}
}

// MockStream for testing Stream interface
type mockStream struct {
	mockPublisher
	mockStreamConsumer
}

func TestStreamInterface(t *testing.T) {
	var stream Stream = &mockStream{}

	// Test Publisher
	err := stream.Publish(context.Background(), "topic", []byte("msg"))
	if err != nil {
		t.Errorf("Publish error: %v", err)
	}

	// Test StreamConsumer
	_, err = stream.Fetch(context.Background(), 0)
	if err != nil {
		t.Errorf("Fetch error: %v", err)
	}
}
