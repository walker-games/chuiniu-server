package service

import (
	"encoding/json"
	"log"

	"github.com/walker-games/chuiniu-server/internal/game"
	"github.com/walker-games/chuiniu-server/internal/model"
	"gorm.io/gorm"
)

type GameLogService struct {
	db *gorm.DB
}

func NewGameLogService(db *gorm.DB) *GameLogService {
	return &GameLogService{db: db}
}

func (s *GameLogService) SaveRound(room *game.Room, allDice map[string][]int, challengerID, winner, loser string, punishment game.Punishment) {
	// Marshal players
	playerInfo := make([]map[string]interface{}, 0, len(room.Players))
	for _, p := range room.Players {
		playerInfo = append(playerInfo, map[string]interface{}{
			"id":   p.ID,
			"name": p.Name,
		})
	}
	playersJSON, err := json.Marshal(playerInfo)
	if err != nil {
		log.Printf("gamelog: marshal players error: %v", err)
		return
	}

	// Marshal all dice
	diceJSON, err := json.Marshal(allDice)
	if err != nil {
		log.Printf("gamelog: marshal dice error: %v", err)
		return
	}

	// Marshal bids
	var bidsJSON []byte
	if room.Round != nil {
		bidsJSON, err = json.Marshal(room.Round.BidHistory)
		if err != nil {
			log.Printf("gamelog: marshal bids error: %v", err)
			return
		}
	} else {
		bidsJSON = []byte("[]")
	}

	// Determine target (the player who made the last bid)
	target := ""
	if room.Round != nil && room.Round.CurrentBid != nil {
		target = room.Round.CurrentBid.PlayerID
	}

	gameLog := model.GameLog{
		RoomID:     room.ID,
		RoundNum:   room.RoundNum,
		Players:    model.JSON(playersJSON),
		AllDice:    model.JSON(diceJSON),
		Bids:       model.JSON(bidsJSON),
		Challenger: challengerID,
		Target:     target,
		Winner:     winner,
		Loser:      loser,
		Punishment: punishment.Text,
	}

	if err := s.db.Create(&gameLog).Error; err != nil {
		log.Printf("gamelog: save error: %v", err)
	}
}
