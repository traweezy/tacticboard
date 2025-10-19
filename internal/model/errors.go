package model

import "errors"

var (
	// ErrRoomNotFound indicates the requested room identifier does not exist.
	ErrRoomNotFound = errors.New("room not found")
	// ErrSequenceConflict signals a non-contiguous sequence number was provided.
	ErrSequenceConflict = errors.New("sequence conflict")
	// ErrSnapshotNotFound occurs when no snapshot is available for the room.
	ErrSnapshotNotFound = errors.New("snapshot not found")
)
