package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/walker-games/chuiniu-server/internal/config"
	"github.com/walker-games/chuiniu-server/internal/game"
	"github.com/walker-games/chuiniu-server/internal/middleware"
	"github.com/walker-games/chuiniu-server/internal/service"
	"github.com/walker-games/chuiniu-server/internal/ws"
)

func SetupRouter(cfg *config.Config, manager *game.RoomManager, hub *ws.Hub, logService *service.GameLogService) *gin.Engine {
	r := gin.Default()

	// CORS
	r.Use(middleware.CORS())

	// Health
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":     "ok",
			"service":    cfg.App.Name,
			"room_count": manager.RoomCount(),
		})
	})

	// Auth middleware
	var authMiddleware gin.HandlerFunc
	if cfg.IAM.JWKSURL != "" {
		authMiddleware = middleware.NewIAMAuth(cfg.IAM.JWKSURL).Middleware()
	} else {
		authMiddleware = middleware.NewLocalAuth("dev-secret").Middleware()
	}

	// REST API
	api := r.Group("/api/v1", authMiddleware)
	{
		roomHandler := NewRoomHandler(manager)
		api.POST("/rooms", roomHandler.Create)
		api.GET("/rooms/:id", roomHandler.Get)
		api.POST("/rooms/join", roomHandler.Join)
	}

	// WebSocket
	wsHandler := NewWSHandler(hub, manager)
	r.GET("/ws/:roomId", authMiddleware, wsHandler.Handle)

	return r
}
