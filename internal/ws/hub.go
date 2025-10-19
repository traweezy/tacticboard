package ws

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/atomic"
	"go.uber.org/zap"

	"github.com/traweezy/tacticboard/internal/config"
	"github.com/traweezy/tacticboard/internal/model"
	"github.com/traweezy/tacticboard/internal/observability"
	"github.com/traweezy/tacticboard/internal/store"
	"github.com/traweezy/tacticboard/internal/util"
)

var (
	writeWait      = 10 * time.Second
	pingInterval   = 20 * time.Second
	pongWait       = 60 * time.Second
	errClosedRoom  = errors.New("room closed")
	errSendTimeout = errors.New("send timeout")
)

// Hub orchestrates room fan-out and persistence.
type Hub struct {
	cfg    config.Config
	store  store.Store
	log    *zap.Logger
	tracer trace.Tracer

	metrics hubMetrics

	roomsMu sync.RWMutex
	rooms   map[string]*roomState
}

type hubMetrics struct {
	connections metric.Int64UpDownCounter
	operations  metric.Int64Counter
}

type roomState struct {
	id      string
	log     *zap.Logger
	clients map[*client]struct{}
	mu      sync.RWMutex
}

type client struct {
	hub    *Hub
	conn   *websocket.Conn
	roomID string
	role   util.CapabilityRole
	since  int64
	send   chan []byte
	log    *zap.Logger
	closed atomic.Bool
	stopCh chan struct{}
}

// NewHub constructs an observable websocket hub.
func NewHub(cfg config.Config, store store.Store, log *zap.Logger, telemetry *observability.Telemetry) *Hub {
	meter := telemetry.MeterProvider.Meter("github.com/traweezy/tacticboard/ws")

	connections, err := meter.Int64UpDownCounter(
		"ws.connections",
		metric.WithDescription("Active WebSocket connections"),
	)
	if err != nil {
		log.Warn("ws metrics: failed to create connection counter", zap.Error(err))
	}

	ops, err := meter.Int64Counter(
		"ws.operations",
		metric.WithDescription("Operations broadcasted to clients"),
	)
	if err != nil {
		log.Warn("ws metrics: failed to create operation counter", zap.Error(err))
	}

	return &Hub{
		cfg:    cfg,
		store:  store,
		log:    log.Named("ws_hub"),
		tracer: telemetry.TracerProvider.Tracer("github.com/traweezy/tacticboard/ws"),
		metrics: hubMetrics{
			connections: connections,
			operations:  ops,
		},
		rooms: make(map[string]*roomState),
	}
}

// HandleConnection performs the hello handshake and launches client loops.
func (h *Hub) HandleConnection(ctx context.Context, conn *websocket.Conn) {
	ctx, span := h.tracer.Start(ctx, "ws.HandleConnection", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	defer func() {
		if err := conn.Close(); err != nil {
			h.log.Warn("close websocket", zap.Error(err))
		}
	}()

	conn.SetReadLimit(h.cfg.WSReadLimit)
	conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(string) error {
		_ = conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	_, data, err := conn.ReadMessage()
	if err != nil {
		h.log.Debug("failed to read hello", zap.Error(err))
		return
	}

	envelope, err := DecodeClientMessage(data)
	if err != nil || envelope.Hello == nil {
		h.writeError(conn, ErrorInvalid, "expected hello message")
		return
	}

	if err := h.validateHello(envelope.Hello); err != nil {
		h.writeError(conn, ErrorInvalid, err.Error())
		return
	}

	claims, err := util.ParseCapabilityToken([]byte(h.cfg.JWTSecret), envelope.Hello.Token)
	if err != nil {
		h.writeError(conn, ErrorUnauthorized, "invalid capability token")
		return
	}

	if claims.RoomID != envelope.Hello.RoomID {
		h.writeError(conn, ErrorUnauthorized, "token does not match room")
		return
	}

	role := util.CapabilityRole(envelope.Hello.Role)
	if role != claims.Role {
		h.writeError(conn, ErrorUnauthorized, "capability role mismatch")
		return
	}

	room, err := h.store.GetRoom(ctx, envelope.Hello.RoomID)
	if err != nil {
		if errors.Is(err, model.ErrRoomNotFound) {
			h.writeError(conn, ErrorUnauthorized, "room not found")
			return
		}
		h.log.Error("load room", zap.Error(err))
		h.writeError(conn, ErrorServer, "failed to load room")
		return
	}

	if envelope.Hello.Since < 0 {
		envelope.Hello.Since = 0
	}

	if envelope.Hello.Since > room.CurrentSeq {
		h.writeError(conn, ErrorConflict, "since ahead of server")
		return
	}

	client := &client{
		hub:    h,
		conn:   conn,
		roomID: room.ID,
		role:   role,
		since:  envelope.Hello.Since,
		send:   make(chan []byte, 256),
		log:    h.log.With(zap.String("room", room.ID)),
		stopCh: make(chan struct{}),
	}

	state := h.getOrCreateRoom(room.ID)
	state.addClient(client)
	h.metrics.observeConnection(ctx, room.ID, +1)
	defer state.removeClient(client)
	defer h.metrics.observeConnection(ctx, room.ID, -1)

	if err := h.sendInitialState(ctx, client, room); err != nil {
		client.log.Warn("send initial state", zap.Error(err))
		return
	}

	go client.writeLoop()
	client.readLoop(ctx)
}

func (h *Hub) sendInitialState(ctx context.Context, c *client, room model.Room) error {
	ctx, span := h.tracer.Start(ctx, "ws.sendInitialState")
	defer span.End()
	span.SetAttributes(attribute.String("room.id", room.ID))

	if room.Snapshot != nil {
		if payload, err := EncodeSnapshot(room.ID, *room.Snapshot); err == nil {
			if err := c.queue(payload); err != nil {
				return err
			}
		} else {
			h.log.Error("encode snapshot", zap.Error(err))
		}
	}

	if room.CurrentSeq > c.since {
		ops, err := h.store.OperationsSince(ctx, room.ID, c.since, 0)
		if err != nil {
			return err
		}
		for _, op := range ops {
			payload, err := EncodeDelta(op)
			if err != nil {
				return err
			}
			if err := c.queue(payload); err != nil {
				return err
			}
		}
	}

	return nil
}

func (h *Hub) validateHello(msg *HelloMessage) error {
	if msg.RoomID == "" {
		return errors.New("roomId required")
	}
	if msg.Token == "" {
		return errors.New("token required")
	}
	if msg.Role != string(util.RoleEdit) && msg.Role != string(util.RoleView) {
		return errors.New("invalid capability role")
	}
	return nil
}

func (h *Hub) writeError(conn *websocket.Conn, code, msg string) {
	_ = conn.WriteMessage(websocket.TextMessage, EncodeError(code, msg))
	if err := conn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.ClosePolicyViolation, msg), time.Now().Add(writeWait)); err != nil {
		h.log.Debug("write error control", zap.Error(err))
	}
}

func (h *Hub) getOrCreateRoom(roomID string) *roomState {
	h.roomsMu.Lock()
	defer h.roomsMu.Unlock()

	state, ok := h.rooms[roomID]
	if !ok {
		state = &roomState{
			id:      roomID,
			log:     h.log.With(zap.String("room", roomID)),
			clients: make(map[*client]struct{}),
		}
		h.rooms[roomID] = state
	}
	return state
}

func (r *roomState) addClient(c *client) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.clients[c] = struct{}{}
	c.closed.Store(false)
	r.log.Info("client joined", zap.Int("total_clients", len(r.clients)))
}

func (r *roomState) removeClient(c *client) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.clients, c)
	if !c.closed.Swap(true) {
		close(c.stopCh)
		close(c.send)
	}
	r.log.Info("client left", zap.Int("total_clients", len(r.clients)))
}

func (r *roomState) broadcast(sender *client, payload []byte) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for client := range r.clients {
		if err := client.queue(payload); err != nil {
			client.log.Warn("drop message", zap.Error(err))
		}
	}
}

func (c *client) queue(payload []byte) error {
	if c.closed.Load() {
		return errClosedRoom
	}

	select {
	case c.send <- payload:
		return nil
	default:
		select {
		case <-c.send:
			c.log.Debug("backpressure: dropped oldest message")
		default:
		}

		select {
		case c.send <- payload:
			return nil
		default:
			return errSendTimeout
		}
	}
}

func (c *client) readLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-c.stopCh:
			return
		default:
		}

		if err := c.conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
			c.log.Warn("set read deadline", zap.Error(err))
			return
		}

		msgType, data, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.log.Warn("unexpected close", zap.Error(err))
			}
			return
		}

		if msgType != websocket.TextMessage {
			continue
		}

		envelope, err := DecodeClientMessage(data)
		if err != nil {
			c.log.Debug("decode client message", zap.Error(err))
			_ = c.queue(EncodeError(ErrorInvalid, "invalid payload"))
			continue
		}

		switch {
		case envelope.Op != nil:
			c.handleOp(ctx, envelope.Op)
		case envelope.Ping != nil:
			c.handlePing(envelope.Ping)
		default:
			c.log.Debug("unexpected message type")
		}
	}
}

func (c *client) writeLoop() {
	ticker := time.NewTicker(pingInterval)
	defer ticker.Stop()

	for {
		select {
		case payload, ok := <-c.send:
			if !ok {
				return
			}
			if err := c.writeMessage(payload); err != nil {
				c.log.Warn("write payload", zap.Error(err))
				return
			}
		case <-ticker.C:
			if err := c.conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(writeWait)); err != nil {
				c.log.Debug("ping control failed", zap.Error(err))
				return
			}
		case <-c.stopCh:
			return
		}
	}
}

func (c *client) writeMessage(payload []byte) error {
	if err := c.conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
		return err
	}
	return c.conn.WriteMessage(websocket.TextMessage, payload)
}

func (c *client) handlePing(msg *PingMessage) {
	response, err := EncodePong(msg.TS)
	if err != nil {
		c.log.Warn("encode pong", zap.Error(err))
		return
	}
	if err := c.queue(response); err != nil {
		c.log.Debug("queue pong", zap.Error(err))
	}
}

func (c *client) handleOp(ctx context.Context, msg *OpMessage) {
	ctx, span := c.hub.tracer.Start(ctx, "ws.handleOp")
	defer span.End()
	span.SetAttributes(
		attribute.String("room.id", c.roomID),
		attribute.Int64("op.seq", msg.Seq),
	)

	if c.role != util.RoleEdit {
		c.log.Warn("discard op from viewer")
		_ = c.queue(EncodeError(ErrorUnauthorized, "edit capability required"))
		return
	}

	if msg.RoomID != c.roomID {
		c.log.Warn("operation room mismatch", zap.String("roomId", msg.RoomID))
		_ = c.queue(EncodeError(ErrorInvalid, "room mismatch"))
		return
	}

	if len(msg.Ops) == 0 {
		return
	}

	op := model.Operation{
		RoomID: c.roomID,
		Seq:    msg.Seq,
		Ops:    msg.Ops,
	}

	op, err := c.hub.store.AppendOperation(ctx, op)
	if err != nil {
		if errors.Is(err, model.ErrSequenceConflict) {
			_ = c.queue(EncodeError(ErrorConflict, "sequence conflict"))
			return
		}
		c.log.Error("append operation", zap.Error(err))
		_ = c.queue(EncodeError(ErrorServer, "operation failed"))
		return
	}

	payload, err := EncodeDelta(op)
	if err != nil {
		c.log.Error("encode delta", zap.Error(err))
		return
	}

	c.hub.metrics.observeOperations(ctx, c.roomID, int64(len(op.Ops)))

	state := c.hub.getOrCreateRoom(c.roomID)
	state.broadcast(c, payload)
}

func (m hubMetrics) observeConnection(ctx context.Context, roomID string, delta int64) {
	if m.connections == nil {
		return
	}
	m.connections.Add(ctx, delta, metric.WithAttributes(attribute.String("room.id", roomID)))
}

func (m hubMetrics) observeOperations(ctx context.Context, roomID string, count int64) {
	if m.operations == nil {
		return
	}
	m.operations.Add(ctx, count, metric.WithAttributes(attribute.String("room.id", roomID)))
}
