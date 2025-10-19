package ws

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/traweezy/tacticboard/internal/model"
)

const (
	TypeHello    = "hello"
	TypeOp       = "op"
	TypePing     = "ping"
	TypeSnapshot = "snapshot"
	TypeDelta    = "delta"
	TypePong     = "pong"
	TypeError    = "error"
)

// Error codes that can be emitted to clients.
const (
	ErrorUnauthorized = "unauthorized"
	ErrorConflict     = "conflict"
	ErrorInvalid      = "invalid"
	ErrorServer       = "server_error"
)

// HelloMessage is the first message a client must send after connecting.
type HelloMessage struct {
	Type   string `json:"type"`
	RoomID string `json:"roomId"`
	Role   string `json:"cap"`
	Since  int64  `json:"since"`
	Token  string `json:"token"`
}

// OpMessage carries an ordered batch of operations.
type OpMessage struct {
	Type   string            `json:"type"`
	RoomID string            `json:"roomId"`
	Seq    int64             `json:"seq"`
	Ops    []json.RawMessage `json:"ops"`
}

// PingMessage keeps the connection alive.
type PingMessage struct {
	Type string `json:"type"`
	TS   int64  `json:"ts"`
}

// ClientEnvelope is the decoded websocket payload.
type ClientEnvelope struct {
	Hello *HelloMessage
	Op    *OpMessage
	Ping  *PingMessage
}

func DecodeClientMessage(data []byte) (ClientEnvelope, error) {
	var base struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(data, &base); err != nil {
		return ClientEnvelope{}, fmt.Errorf("decode message type: %w", err)
	}

	switch base.Type {
	case TypeHello:
		var msg HelloMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			return ClientEnvelope{}, fmt.Errorf("decode hello: %w", err)
		}
		return ClientEnvelope{Hello: &msg}, nil
	case TypeOp:
		var msg OpMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			return ClientEnvelope{}, fmt.Errorf("decode op: %w", err)
		}
		return ClientEnvelope{Op: &msg}, nil
	case TypePing:
		var msg PingMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			return ClientEnvelope{}, fmt.Errorf("decode ping: %w", err)
		}
		return ClientEnvelope{Ping: &msg}, nil
	default:
		return ClientEnvelope{}, errors.New("unsupported message type")
	}
}

// SnapshotPayload is emitted after a successful hello handshake.
type SnapshotPayload struct {
	Type   string          `json:"type"`
	RoomID string          `json:"roomId"`
	Seq    int64           `json:"seq"`
	State  json.RawMessage `json:"state"`
}

// DeltaPayload contains incremental updates applied to a room.
type DeltaPayload struct {
	Type string            `json:"type"`
	Room string            `json:"roomId"`
	From int64             `json:"from"`
	To   int64             `json:"to"`
	Ops  []json.RawMessage `json:"ops"`
}

// ErrorPayload transmits a problem to the client.
type ErrorPayload struct {
	Type  string `json:"type"`
	Code  string `json:"code"`
	Msg   string `json:"msg"`
	Trace string `json:"trace,omitempty"`
}

func EncodeSnapshot(roomID string, snapshot model.Snapshot) ([]byte, error) {
	payload := SnapshotPayload{
		Type:   TypeSnapshot,
		RoomID: roomID,
		Seq:    snapshot.Seq,
		State:  snapshot.State,
	}
	return json.Marshal(payload)
}

func EncodeDelta(op model.Operation) ([]byte, error) {
	payload := DeltaPayload{
		Type: TypeDelta,
		Room: op.RoomID,
		From: op.Seq - 1,
		To:   op.Seq,
		Ops:  op.Ops,
	}
	return json.Marshal(payload)
}

func EncodePong(ts int64) ([]byte, error) {
	if ts == 0 {
		ts = time.Now().UnixMilli()
	}
	return json.Marshal(PingMessage{
		Type: TypePong,
		TS:   ts,
	})
}

func EncodeError(code, msg string) []byte {
	payload, _ := json.Marshal(ErrorPayload{
		Type: TypeError,
		Code: code,
		Msg:  msg,
	})
	return payload
}
