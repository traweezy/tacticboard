package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/traweezy/tacticboard/internal/config"
	"github.com/traweezy/tacticboard/internal/model"
	"github.com/traweezy/tacticboard/internal/store"
	"github.com/traweezy/tacticboard/internal/util"
)

const (
	defaultShareTTL = 24 * time.Hour
	maxShareTTL     = 7 * 24 * time.Hour
)

type RoomHandler struct {
	cfg   config.Config
	store store.Store
	ids   *util.IDGenerator
	log   *zap.Logger
}

func NewRoomHandler(cfg config.Config, store store.Store, ids *util.IDGenerator, log *zap.Logger) *RoomHandler {
	return &RoomHandler{
		cfg:   cfg,
		store: store,
		ids:   ids,
		log:   log.Named("rooms_handler"),
	}
}

func (h *RoomHandler) CreateRoom(c *gin.Context) {
	ctx := c.Request.Context()
	now := time.Now().UTC()
	roomID := h.ids.New()
	initialState := json.RawMessage(`{"nodes":[],"layers":[],"meta":{}}`)

	room := model.Room{
		ID:         roomID,
		CreatedAt:  now,
		UpdatedAt:  now,
		CurrentSeq: 0,
		Snapshot: &model.Snapshot{
			RoomID:    roomID,
			Seq:       0,
			State:     initialState,
			CreatedAt: now,
		},
	}

	if _, err := h.store.CreateRoom(ctx, room); err != nil {
		h.log.Error("create room", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create room"})
		return
	}

	viewToken, viewExpiry, err := h.newCapability(roomID, util.RoleView, now, defaultShareTTL)
	if err != nil {
		h.log.Error("create view token", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create tokens"})
		return
	}

	editToken, editExpiry, err := h.newCapability(roomID, util.RoleEdit, now, defaultShareTTL)
	if err != nil {
		h.log.Error("create edit token", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create tokens"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":        roomID,
		"createdAt": now,
		"viewToken": viewToken,
		"editToken": editToken,
		"links": gin.H{
			"view": shareURL(roomID, viewToken),
			"edit": shareURL(roomID, editToken),
		},
		"expires": gin.H{
			"view": viewExpiry,
			"edit": editExpiry,
		},
	})
}

func (h *RoomHandler) GetRoom(c *gin.Context) {
	ctx := c.Request.Context()
	roomID := c.Param("id")

	room, err := h.store.GetRoom(ctx, roomID)
	if err != nil {
		if errors.Is(err, model.ErrRoomNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "room not found"})
			return
		}
		h.log.Error("get room", zap.String("room", roomID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load room"})
		return
	}

	resp := gin.H{
		"id":         room.ID,
		"createdAt":  room.CreatedAt,
		"updatedAt":  room.UpdatedAt,
		"currentSeq": room.CurrentSeq,
	}

	if room.Snapshot != nil {
		resp["snapshot"] = gin.H{
			"seq":   room.Snapshot.Seq,
			"state": room.Snapshot.State,
		}
	}

	c.JSON(http.StatusOK, resp)
}

type shareRequest struct {
	Role       util.CapabilityRole `json:"role" binding:"required"`
	TTLMinutes int                 `json:"ttlMinutes"`
}

func (h *RoomHandler) ShareRoom(c *gin.Context) {
	roomID := c.Param("id")
	ctx := c.Request.Context()

	if _, err := h.store.GetRoom(ctx, roomID); err != nil {
		if errors.Is(err, model.ErrRoomNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "room not found"})
			return
		}
		h.log.Error("lookup room before share", zap.String("room", roomID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load room"})
		return
	}

	var req shareRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	if req.Role != util.RoleView && req.Role != util.RoleEdit {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid role"})
		return
	}

	ttl := durationFromMinutes(req.TTLMinutes)
	now := time.Now().UTC()
	token, expiry, err := h.newCapability(roomID, req.Role, now, ttl)
	if err != nil {
		h.log.Error("generate share token", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":  token,
		"role":   req.Role,
		"expiry": expiry,
		"link":   shareURL(roomID, token),
	})
}

func (h *RoomHandler) newCapability(roomID string, role util.CapabilityRole, issuedAt time.Time, ttl time.Duration) (string, time.Time, error) {
	if ttl <= 0 {
		ttl = defaultShareTTL
	}
	if ttl > maxShareTTL {
		ttl = maxShareTTL
	}

	claims := util.CapabilityClaims{
		RoomID:    roomID,
		Role:      role,
		IssuedAt:  issuedAt,
		ExpiresAt: issuedAt.Add(ttl),
	}
	token, err := util.GenerateCapabilityToken([]byte(h.cfg.JWTSecret), claims)
	return token, claims.ExpiresAt, err
}

func shareURL(roomID, token string) string {
	return fmt.Sprintf("/room/%s?token=%s", url.PathEscape(roomID), url.QueryEscape(token))
}

func durationFromMinutes(minutes int) time.Duration {
	if minutes <= 0 {
		return defaultShareTTL
	}
	return time.Duration(minutes) * time.Minute
}
