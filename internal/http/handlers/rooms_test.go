package handlers

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/traweezy/tacticboard/internal/config"
	"github.com/traweezy/tacticboard/internal/store"
	"github.com/traweezy/tacticboard/internal/util"
	"go.uber.org/zap"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type testDeps struct {
	handler *RoomHandler
	store   store.Store
}

func newTestDeps(t *testing.T) testDeps {
	t.Helper()
	cfg := config.Config{JWTSecret: strings.Repeat("s", 16)}
	ids, err := util.NewIDGenerator()
	require.NoError(t, err)
	st := store.NewMemoryStore()
	handler := NewRoomHandler(cfg, st, ids, zap.NewNop())
	return testDeps{handler: handler, store: st}
}

func TestRoomHandler_CreateRoom(t *testing.T) {
	gin.SetMode(gin.TestMode)
	deps := newTestDeps(t)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/rooms", nil)
	deps.handler.CreateRoom(c)
	require.Equal(t, http.StatusCreated, w.Code)
	var payload map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &payload))
	require.NotEmpty(t, payload["viewToken"])
	require.NotEmpty(t, payload["editToken"])
	require.NotEmpty(t, payload["id"])
}

func TestRoomHandler_ShareRoom_InvalidRole(t *testing.T) {
	gin.SetMode(gin.TestMode)
	deps := newTestDeps(t)
	// create baseline room via handler
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/rooms", nil)
	deps.handler.CreateRoom(c)
	require.Equal(t, http.StatusCreated, w.Code)
	var created map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &created))
	roomID := created["id"].(string)
	body := strings.NewReader(`{"role":"invalid"}`)
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodPost, "/api/rooms/"+roomID+"/share", body)
	req.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{gin.Param{Key: "id", Value: roomID}}
	c.Request = req
	deps.handler.ShareRoom(c)
	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRoomHandler_GetRoom_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	deps := newTestDeps(t)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{gin.Param{Key: "id", Value: "missing"}}
	c.Request = httptest.NewRequest(http.MethodGet, "/api/rooms/missing", nil)
	deps.handler.GetRoom(c)
	require.Equal(t, http.StatusNotFound, w.Code)
}
