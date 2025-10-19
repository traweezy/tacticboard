package store

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/traweezy/tacticboard/internal/model"
)

func TestMemoryStore_CreateAndGetRoom(t *testing.T) {
	store := NewMemoryStore()
	room := model.Room{ID: "room-1"}

	created, err := store.CreateRoom(context.Background(), room)
	require.NoError(t, err)
	require.Equal(t, "room-1", created.ID)

	fetched, err := store.GetRoom(context.Background(), "room-1")
	require.NoError(t, err)
	require.Equal(t, created.ID, fetched.ID)
	require.Equal(t, created.CurrentSeq, fetched.CurrentSeq)
	require.NotZero(t, fetched.CreatedAt)
	require.NotZero(t, fetched.UpdatedAt)
}

func TestMemoryStore_AppendOperationSequence(t *testing.T) {
	store := NewMemoryStore()
	_, err := store.CreateRoom(context.Background(), model.Room{ID: "room-2"})
	require.NoError(t, err)

	op := model.Operation{
		RoomID: "room-2",
		Seq:    1,
		Ops:    []json.RawMessage{json.RawMessage(`{"k":"add"}`)},
	}

	_, err = store.AppendOperation(context.Background(), op)
	require.NoError(t, err)

	_, err = store.AppendOperation(context.Background(), model.Operation{
		RoomID: "room-2",
		Seq:    3,
		Ops:    []json.RawMessage{json.RawMessage(`{"k":"move"}`)},
	})
	require.ErrorIs(t, err, model.ErrSequenceConflict)
}

func TestMemoryStore_OperationsSince(t *testing.T) {
	store := NewMemoryStore()
	_, err := store.CreateRoom(context.Background(), model.Room{ID: "room-3"})
	require.NoError(t, err)

	for seq := int64(1); seq <= 5; seq++ {
		payload := json.RawMessage(fmt.Sprintf(`{"seq":%d}`, seq))
		_, err := store.AppendOperation(context.Background(), model.Operation{
			RoomID: "room-3",
			Seq:    seq,
			Ops:    []json.RawMessage{payload},
		})
		require.NoError(t, err)
	}

	ops, err := store.OperationsSince(context.Background(), "room-3", 2, 0)
	require.NoError(t, err)
	require.Len(t, ops, 3)
	require.EqualValues(t, 3, ops[0].Seq)

	opsLimited, err := store.OperationsSince(context.Background(), "room-3", 0, 2)
	require.NoError(t, err)
	require.Len(t, opsLimited, 2)
	require.EqualValues(t, 1, opsLimited[0].Seq)
}

func TestMemoryStore_SaveSnapshot(t *testing.T) {
	store := NewMemoryStore()
	_, err := store.CreateRoom(context.Background(), model.Room{ID: "room-4"})
	require.NoError(t, err)

	snapshot := model.Snapshot{
		RoomID:    "room-4",
		Seq:       10,
		State:     json.RawMessage(`{"nodes":[]}`),
		CreatedAt: time.Now().UTC(),
	}

	require.NoError(t, store.SaveSnapshot(context.Background(), snapshot))

	stored, err := store.LatestSnapshot(context.Background(), "room-4")
	require.NoError(t, err)
	require.Equal(t, snapshot.Seq, stored.Seq)

	room, err := store.GetRoom(context.Background(), "room-4")
	require.NoError(t, err)
	require.EqualValues(t, 10, room.CurrentSeq)
	require.NotNil(t, room.Snapshot)
}
