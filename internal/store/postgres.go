package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/traweezy/tacticboard/internal/model"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
)

type postgresStore struct {
	db *gorm.DB
}

func newPostgresStore(dsn string) (Store, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return nil, err
	}

	return &postgresStore{db: db}, nil
}

func (s *postgresStore) CreateRoom(ctx context.Context, room model.Room) (model.Room, error) {
	now := time.Now().UTC()
	if room.ID == "" {
		return model.Room{}, errors.New("room id required")
	}
	if room.CreatedAt.IsZero() {
		room.CreatedAt = now
	}
	if room.UpdatedAt.IsZero() {
		room.UpdatedAt = room.CreatedAt
	}

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existing int64
		if err := tx.Model(&roomRow{}).Where("id = ?", room.ID).Count(&existing).Error; err != nil {
			return err
		}
		if existing > 0 {
			return errors.New("room already exists")
		}

		if err := tx.Create(&roomRow{
			ID:        room.ID,
			CreatedAt: room.CreatedAt,
		}).Error; err != nil {
			return err
		}

		if room.Snapshot != nil {
			snapshot := snapshotRow{
				RoomID:    room.ID,
				Seq:       room.Snapshot.Seq,
				Body:      cloneBytes(room.Snapshot.State),
				CreatedAt: room.Snapshot.CreatedAt,
			}
			if snapshot.CreatedAt.IsZero() {
				snapshot.CreatedAt = now
			}
			if err := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "room_id"}, {Name: "seq"}},
				DoNothing: true,
			}).Create(&snapshot).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return model.Room{}, err
	}

	return room, nil
}

func (s *postgresStore) GetRoom(ctx context.Context, roomID string) (model.Room, error) {
	var roomRec roomRow
	if err := s.db.WithContext(ctx).First(&roomRec, "id = ?", roomID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return model.Room{}, model.ErrRoomNotFound
		}
		return model.Room{}, err
	}

	var maxOp struct {
		Seq       sql.NullInt64
		UpdatedAt sql.NullTime
	}
	if err := s.db.WithContext(ctx).
		Model(&operationRow{}).
		Select("MAX(seq) AS seq, MAX(created_at) AS updated_at").
		Where("room_id = ?", roomID).
		Scan(&maxOp).Error; err != nil {
		return model.Room{}, err
	}

	var latestSnapshot snapshotRow
	snapshotErr := s.db.WithContext(ctx).
		Where("room_id = ?", roomID).
		Order("seq DESC").
		Limit(1).
		Take(&latestSnapshot).Error
	if snapshotErr != nil && !errors.Is(snapshotErr, gorm.ErrRecordNotFound) {
		return model.Room{}, snapshotErr
	}

	currentSeq := maxOp.Seq.Int64
	if latestSnapshot.Seq > currentSeq {
		currentSeq = latestSnapshot.Seq
	}

	updatedAt := roomRec.CreatedAt
	if maxOp.UpdatedAt.Valid && maxOp.UpdatedAt.Time.After(updatedAt) {
		updatedAt = maxOp.UpdatedAt.Time.UTC()
	}
	if !latestSnapshot.CreatedAt.IsZero() && latestSnapshot.CreatedAt.After(updatedAt) {
		updatedAt = latestSnapshot.CreatedAt
	}

	room := model.Room{
		ID:         roomRec.ID,
		CreatedAt:  roomRec.CreatedAt,
		UpdatedAt:  updatedAt,
		CurrentSeq: currentSeq,
	}

	if latestSnapshot.RoomID != "" {
		room.Snapshot = &model.Snapshot{
			RoomID:    latestSnapshot.RoomID,
			Seq:       latestSnapshot.Seq,
			State:     cloneBytes(latestSnapshot.Body),
			CreatedAt: latestSnapshot.CreatedAt,
		}
	}

	return room, nil
}

func (s *postgresStore) SaveSnapshot(ctx context.Context, snapshot model.Snapshot) error {
	if snapshot.RoomID == "" {
		return errors.New("room id required")
	}

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var exists int64
		if err := tx.Model(&roomRow{}).Where("id = ?", snapshot.RoomID).Count(&exists).Error; err != nil {
			return err
		}
		if exists == 0 {
			return model.ErrRoomNotFound
		}

		record := snapshotRow{
			RoomID:    snapshot.RoomID,
			Seq:       snapshot.Seq,
			Body:      cloneBytes(snapshot.State),
			CreatedAt: snapshot.CreatedAt,
		}
		if record.CreatedAt.IsZero() {
			record.CreatedAt = time.Now().UTC()
		}

		return tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "room_id"}, {Name: "seq"}},
			DoUpdates: clause.AssignmentColumns([]string{"body", "created_at"}),
		}).Create(&record).Error
	})
}

func (s *postgresStore) LatestSnapshot(ctx context.Context, roomID string) (model.Snapshot, error) {
	var record snapshotRow
	if err := s.db.WithContext(ctx).
		Where("room_id = ?", roomID).
		Order("seq DESC").
		Limit(1).
		Take(&record).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return model.Snapshot{}, model.ErrSnapshotNotFound
		}
		return model.Snapshot{}, err
	}

	return model.Snapshot{
		RoomID:    record.RoomID,
		Seq:       record.Seq,
		State:     cloneBytes(record.Body),
		CreatedAt: record.CreatedAt,
	}, nil
}

func (s *postgresStore) AppendOperation(ctx context.Context, op model.Operation) (model.Operation, error) {
	if op.RoomID == "" {
		return model.Operation{}, errors.New("room id required")
	}

	var persisted model.Operation
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var lastSeq sql.NullInt64
		if err := tx.Model(&operationRow{}).
			Select("MAX(seq)").
			Where("room_id = ?", op.RoomID).
			Scan(&lastSeq).Error; err != nil {
			return err
		}

		expected := lastSeq.Int64 + 1
		if op.Seq != expected {
			return model.ErrSequenceConflict
		}

		body, err := json.Marshal(op.Ops)
		if err != nil {
			return err
		}

		record := operationRow{
			RoomID:    op.RoomID,
			Seq:       op.Seq,
			Body:      body,
			CreatedAt: time.Now().UTC(),
		}

		if err := tx.Create(&record).Error; err != nil {
			return err
		}

		persisted = model.Operation{
			RoomID:    record.RoomID,
			Seq:       record.Seq,
			CreatedAt: record.CreatedAt,
		}

		if len(op.Ops) > 0 {
			persisted.Ops = make([]json.RawMessage, len(op.Ops))
			for i, raw := range op.Ops {
				persisted.Ops[i] = cloneBytes(raw)
			}
		}

		return nil
	})

	if err != nil {
		return model.Operation{}, err
	}

	return persisted, nil
}

func (s *postgresStore) OperationsSince(ctx context.Context, roomID string, sinceSeq int64, limit int) ([]model.Operation, error) {
	var records []operationRow
	query := s.db.WithContext(ctx).
		Where("room_id = ? AND seq > ?", roomID, sinceSeq).
		Order("seq ASC")
	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Find(&records).Error; err != nil {
		return nil, err
	}

	ops := make([]model.Operation, 0, len(records))
	for _, rec := range records {
		var payload []json.RawMessage
		if len(rec.Body) > 0 {
			if err := json.Unmarshal(rec.Body, &payload); err != nil {
				return nil, err
			}
		}

		ops = append(ops, model.Operation{
			RoomID:    rec.RoomID,
			Seq:       rec.Seq,
			Ops:       payload,
			CreatedAt: rec.CreatedAt,
		})
	}

	return ops, nil
}

type roomRow struct {
	ID        string    `gorm:"column:id;primaryKey"`
	CreatedAt time.Time `gorm:"column:created_at"`
}

func (roomRow) TableName() string { return "rooms" }

type snapshotRow struct {
	RoomID    string    `gorm:"column:room_id;primaryKey"`
	Seq       int64     `gorm:"column:seq;primaryKey"`
	Body      []byte    `gorm:"column:body"`
	CreatedAt time.Time `gorm:"column:created_at"`
}

func (snapshotRow) TableName() string { return "snapshots" }

type operationRow struct {
	RoomID    string    `gorm:"column:room_id;primaryKey"`
	Seq       int64     `gorm:"column:seq;primaryKey"`
	Body      []byte    `gorm:"column:body"`
	CreatedAt time.Time `gorm:"column:created_at"`
}

func (operationRow) TableName() string { return "ops" }

func cloneBytes(src []byte) []byte {
	if src == nil {
		return nil
	}
	dst := make([]byte, len(src))
	copy(dst, src)
	return dst
}
