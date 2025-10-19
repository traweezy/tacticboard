package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"

	"github.com/traweezy/tacticboard/internal/config"
	"github.com/traweezy/tacticboard/internal/ws"
)

// WSHandler upgrades HTTP connections into websocket sessions handled by the hub.
type WSHandler struct {
	hub *ws.Hub
	log *zap.Logger
	upg websocket.Upgrader
}

func NewWSHandler(cfg config.Config, hub *ws.Hub, log *zap.Logger) *WSHandler {
	upgrader := websocket.Upgrader{
		ReadBufferSize:    1024,
		WriteBufferSize:   cfg.WSWriteBuffer,
		EnableCompression: true,
		CheckOrigin: func(r *http.Request) bool {
			// Rely on capability tokens for authorization. Additional origin checks can be configured via reverse proxy.
			return true
		},
	}

	return &WSHandler{
		hub: hub,
		log: log.Named("ws_handler"),
		upg: upgrader,
	}
}

func (h *WSHandler) Serve(c *gin.Context) {
	conn, err := h.upg.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		h.log.Warn("upgrade failed", zap.Error(err))
		return
	}

	h.hub.HandleConnection(c.Request.Context(), conn)
}
