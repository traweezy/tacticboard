package ws

import (
    "encoding/json"
    "testing"

    "github.com/stretchr/testify/require"
    "go.opentelemetry.io/otel/trace"
    "go.uber.org/zap"
)

func TestClientQueueDropsOldest(t *testing.T) {
    hub := &Hub{
        log:   zap.NewNop(),
        tracer: trace.NewNoopTracerProvider().Tracer("test"),
        metrics: hubMetrics{},
    }
    c := &client{
        hub:  hub,
        send: make(chan []byte, 2),
        log:  zap.NewNop(),
    }

	require.NoError(t, c.queue([]byte("a")))
	require.NoError(t, c.queue([]byte("b")))
	require.NoError(t, c.queue([]byte("c")))

	first := <-c.send
	second := <-c.send

	require.Equal(t, []byte("b"), first)
	require.Equal(t, []byte("c"), second)
}

func TestClientQueueClosed(t *testing.T) {
    c := &client{
        hub:  &Hub{tracer: trace.NewNoopTracerProvider().Tracer("test")},
        send: make(chan []byte, 1),
        log:  zap.NewNop(),
    }
	c.closed.Store(true)

	err := c.queue([]byte("payload"))
	require.ErrorIs(t, err, errClosedRoom)
}

func TestClientHandlePingEnqueuesPong(t *testing.T) {
    hub := &Hub{
        tracer: trace.NewNoopTracerProvider().Tracer("test"),
        metrics: hubMetrics{},
        log:    zap.NewNop(),
    }
    c := &client{
        hub:  hub,
        send: make(chan []byte, 1),
        log:  zap.NewNop(),
    }

	ping := &PingMessage{Type: TypePing, TS: 123}
	c.handlePing(ping)

	payload := <-c.send
	var pong PingMessage
	require.NoError(t, json.Unmarshal(payload, &pong))
	require.Equal(t, TypePong, pong.Type)
	require.EqualValues(t, 123, pong.TS)
}
