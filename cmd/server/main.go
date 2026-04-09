package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/walker-games/chuiniu-server/internal/config"
	"github.com/walker-games/chuiniu-server/internal/game"
	"github.com/walker-games/chuiniu-server/internal/handler"
	"github.com/walker-games/chuiniu-server/internal/model"
	"github.com/walker-games/chuiniu-server/internal/service"
	"github.com/walker-games/chuiniu-server/internal/ws"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	configPath := flag.String("config", "config/config.yaml", "config file path")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	// Database
	dsn := cfg.Database.DSN()
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("connect database: %v", err)
	}
	if err := db.AutoMigrate(&model.GameLog{}); err != nil {
		log.Fatalf("auto migrate: %v", err)
	}
	log.Println("database connected and migrated")

	// Room Manager
	idleTimeout, err := time.ParseDuration(cfg.Room.IdleTimeout)
	if err != nil {
		idleTimeout = 30 * time.Minute
		log.Printf("invalid idle_timeout, using default: %v", idleTimeout)
	}
	manager := game.NewRoomManager(idleTimeout)

	// GameLog Service
	logService := service.NewGameLogService(db)

	// WebSocket Hub
	hub := ws.NewHub(manager, logService)
	go hub.Run()

	// Router
	r := handler.SetupRouter(cfg, manager, hub, logService)

	addr := fmt.Sprintf(":%d", cfg.App.Port)
	log.Printf("starting %s on %s", cfg.App.Name, addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
