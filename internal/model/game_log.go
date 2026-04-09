package model

import "time"

type GameLog struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	RoomID     string    `gorm:"size:20;index" json:"room_id"`
	RoundNum   int       `json:"round_num"`
	Players    JSON      `gorm:"type:json" json:"players"`
	AllDice    JSON      `gorm:"type:json" json:"all_dice"`
	Bids       JSON      `gorm:"type:json" json:"bids"`
	Challenger string    `gorm:"size:50" json:"challenger"`
	Target     string    `gorm:"size:50" json:"target"`
	Winner     string    `gorm:"size:50" json:"winner"`
	Loser      string    `gorm:"size:50" json:"loser"`
	Punishment string    `gorm:"size:100" json:"punishment"`
	CreatedAt  time.Time `json:"created_at"`
}
