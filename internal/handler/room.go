package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/walker-games/chuiniu-server/internal/game"
	"github.com/walker-games/chuiniu-server/internal/middleware"
)

type RoomHandler struct {
	manager *game.RoomManager
}

func NewRoomHandler(manager *game.RoomManager) *RoomHandler {
	return &RoomHandler{manager: manager}
}

func (h *RoomHandler) Create(c *gin.Context) {
	userID := middleware.GetUserID(c)
	username := middleware.GetUsername(c)
	if userID == "" {
		RespondError(c, http.StatusUnauthorized, game.ErrCodeUnauthorized, nil)
		return
	}

	// Accept name from request body or query
	var req struct {
		Name string `json:"name"`
	}
	c.ShouldBindJSON(&req)
	if req.Name != "" {
		username = req.Name
	}

	avatar := ""
	room := h.manager.CreateRoom(userID, username, avatar)

	c.JSON(http.StatusOK, gin.H{
		"room_id": room.ID,
		"code":    room.Code,
	})
}

func (h *RoomHandler) Get(c *gin.Context) {
	roomID := c.Param("id")
	room := h.manager.GetRoom(roomID)
	if room == nil {
		RespondError(c, http.StatusNotFound, game.ErrCodeRoomNotFound, nil)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"room_id":      room.ID,
		"code":         room.Code,
		"host":         room.Host,
		"status":       room.Status,
		"player_count": len(room.Players),
		"max_players":  room.Settings.MaxPlayers,
	})
}

type JoinRequest struct {
	Code string `json:"code" binding:"required"`
	Name string `json:"name"`
}

func (h *RoomHandler) Join(c *gin.Context) {
	var req JoinRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, http.StatusBadRequest, game.ErrCodeInvalidRequest, nil)
		return
	}

	userID := middleware.GetUserID(c)
	username := middleware.GetUsername(c)
	if req.Name != "" {
		username = req.Name
	}
	if userID == "" {
		RespondError(c, http.StatusUnauthorized, game.ErrCodeUnauthorized, nil)
		return
	}

	// Try invite code first, then room ID
	room := h.manager.GetRoomByCode(req.Code)
	if room == nil {
		room = h.manager.GetRoom(req.Code)
	}
	if room == nil {
		RespondError(c, http.StatusNotFound, game.ErrCodeRoomNotFound, nil)
		return
	}

	avatar := ""
	if err := room.AddPlayer(userID, username, avatar); err != nil {
		RespondGameError(c, http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"room_id": room.ID,
		"code":    room.Code,
	})
}
