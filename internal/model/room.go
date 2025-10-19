package model

import (
	"encoding/json"
	"time"
)

// Room captures metadata about a collaborative room.
type Room struct {
	ID         string    `json:"id"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
	CurrentSeq int64     `json:"currentSeq"`
	Snapshot   *Snapshot `json:"snapshot,omitempty"`
}

// Snapshot represents the full state of a room at a particular sequence.
type Snapshot struct {
	RoomID    string          `json:"roomId"`
	Seq       int64           `json:"seq"`
	State     json.RawMessage `json:"state"`
	CreatedAt time.Time       `json:"createdAt"`
}

// Operation is a batch of ordered ops applied to a room.
type Operation struct {
	RoomID    string            `json:"roomId"`
	Seq       int64             `json:"seq"`
	Ops       []json.RawMessage `json:"ops"`
	CreatedAt time.Time         `json:"createdAt"`
}

// Clone returns a deep copy of the operation ensuring payload immutability.
func (o Operation) Clone() Operation {
	clone := Operation{
		RoomID:    o.RoomID,
		Seq:       o.Seq,
		CreatedAt: o.CreatedAt,
	}

	if len(o.Ops) > 0 {
		clone.Ops = make([]json.RawMessage, len(o.Ops))
		for i := range o.Ops {
			if o.Ops[i] != nil {
				buf := make([]byte, len(o.Ops[i]))
				copy(buf, o.Ops[i])
				clone.Ops[i] = json.RawMessage(buf)
			}
		}
	}

	return clone
}
