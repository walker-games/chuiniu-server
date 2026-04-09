package handler

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/walker-games/chuiniu-server/internal/game"
	"github.com/walker-games/chuiniu-server/internal/middleware"
	"github.com/walker-games/chuiniu-server/internal/ws"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for TG Mini App
	},
}

type WSHandler struct {
	hub     *ws.Hub
	manager *game.RoomManager
}

func NewWSHandler(hub *ws.Hub, manager *game.RoomManager) *WSHandler {
	return &WSHandler{hub: hub, manager: manager}
}

func (h *WSHandler) Handle(c *gin.Context) {
	roomID := c.Param("roomId")
	userID := middleware.GetUserID(c)
	username := middleware.GetUsername(c)

	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not identified"})
		return
	}

	room := h.manager.GetRoom(roomID)
	if room == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "room not found"})
		return
	}

	// Ensure player is in the room (add if rejoining)
	if err := room.AddPlayer(userID, username, ""); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("ws upgrade error: %v", err)
		return
	}

	client := &Client{
		Hub:      h.hub,
		Conn:     conn,
		Send:     make(chan []byte, 256),
		RoomID:   roomID,
		PlayerID: userID,
	}

	h.hub.Register <- client

	go client.WritePump()
	go client.ReadPump()

	// Send room state directly to the new client (don't rely on broadcast — Register is async)
	state := h.hub.BuildRoomStateForPlayer(room, userID)
	client.SendMessage(ws.NewMessage(ws.MsgRoomState, state))

	// Broadcast to existing players after a short delay so Register completes
	go func() {
		time.Sleep(100 * time.Millisecond)
		h.hub.BroadcastRoomState(room)
	}()
}

// Client wraps ws.Client for the handler package.
type Client = ws.Client
