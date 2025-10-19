package store

import (
	"context"
	"errors"
	"sort"
	"sync"
	"time"

	"github.com/traweezy/tacticboard/internal/model"
)

// memoryStore implements Store using process memory. It is safe for concurrent use.
type memoryStore struct {
	mu    sync.RWMutex
	rooms map[string]*roomRecord
}

type roomRecord struct {
	room     model.Room
	snapshot *model.Snapshot
	ops      []model.Operation
}

// NewMemoryStore constructs the default in-memory store.
func NewMemoryStore() Store {
	return &memoryStore{
		rooms: make(map[string]*roomRecord),
	}
}

func (m *memoryStore) CreateRoom(_ context.Context, room model.Room) (model.Room, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.rooms[room.ID]; exists {
		return model.Room{}, errors.New("room already exists")
	}

	now := time.Now().UTC()
	if room.CreatedAt.IsZero() {
		room.CreatedAt = now
	}
	if room.UpdatedAt.IsZero() {
		room.UpdatedAt = room.CreatedAt
	}

	record := &roomRecord{
		room: room,
	}
	if room.Snapshot != nil {
		record.snapshot = cloneSnapshot(room.Snapshot)
	}

	m.rooms[room.ID] = record
	return copyRoom(record), nil
}

func (m *memoryStore) GetRoom(_ context.Context, roomID string) (model.Room, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	record, ok := m.rooms[roomID]
	if !ok {
		return model.Room{}, model.ErrRoomNotFound
	}

	return copyRoom(record), nil
}

func (m *memoryStore) SaveSnapshot(_ context.Context, snapshot model.Snapshot) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	record, ok := m.rooms[snapshot.RoomID]
	if !ok {
		return model.ErrRoomNotFound
	}

	clone := cloneSnapshot(&snapshot)
	record.snapshot = clone
	record.room.Snapshot = clone
	record.room.CurrentSeq = snapshot.Seq
	record.room.UpdatedAt = snapshot.CreatedAt
	return nil
}

func (m *memoryStore) LatestSnapshot(_ context.Context, roomID string) (model.Snapshot, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	record, ok := m.rooms[roomID]
	if !ok {
		return model.Snapshot{}, model.ErrRoomNotFound
	}

	if record.snapshot == nil {
		return model.Snapshot{}, model.ErrSnapshotNotFound
	}

	return *cloneSnapshot(record.snapshot), nil
}

func (m *memoryStore) AppendOperation(_ context.Context, op model.Operation) (model.Operation, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	record, ok := m.rooms[op.RoomID]
	if !ok {
		return model.Operation{}, model.ErrRoomNotFound
	}

	expectedSeq := record.room.CurrentSeq + 1
	if op.Seq != expectedSeq {
		return model.Operation{}, model.ErrSequenceConflict
	}

	op.CreatedAt = time.Now().UTC()
	record.ops = append(record.ops, op.Clone())
	record.room.CurrentSeq = op.Seq
	record.room.UpdatedAt = op.CreatedAt

	return op.Clone(), nil
}

func (m *memoryStore) OperationsSince(_ context.Context, roomID string, sinceSeq int64, limit int) ([]model.Operation, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	record, ok := m.rooms[roomID]
	if !ok {
		return nil, model.ErrRoomNotFound
	}

	ops := make([]model.Operation, 0)
	idx := sort.Search(len(record.ops), func(i int) bool {
		return record.ops[i].Seq > sinceSeq
	})

	for i := idx; i < len(record.ops); i++ {
		if limit > 0 && len(ops) >= limit {
			break
		}
		ops = append(ops, record.ops[i].Clone())
	}

	return ops, nil
}

func cloneSnapshot(src *model.Snapshot) *model.Snapshot {
	if src == nil {
		return nil
	}

	dst := *src
	if src.State != nil {
		buf := make([]byte, len(src.State))
		copy(buf, src.State)
		dst.State = buf
	}

	return &dst
}

func copyRoom(record *roomRecord) model.Room {
	room := record.room
	room.Snapshot = cloneSnapshot(record.snapshot)
	return room
}
