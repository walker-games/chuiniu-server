package game

import (
	"errors"
	"math/rand"
	"sync"
	"time"
)

var (
	ErrRoomFull         = errors.New("room is full")
	ErrGameInProgress   = errors.New("game already in progress")
	ErrNotYourTurn      = errors.New("not your turn")
	ErrInvalidBid       = errors.New("invalid bid")
	ErrNoBidToChallenge = errors.New("no bid to challenge")
)

type RoomStatus string

const (
	StatusWaiting  RoomStatus = "waiting"
	StatusPlaying  RoomStatus = "playing"
	StatusSettling RoomStatus = "settling"
)

type RoundPhase string

const (
	PhaseRolling     RoundPhase = "rolling"
	PhaseBidding     RoundPhase = "bidding"
	PhaseChallenging RoundPhase = "challenging"
	PhaseSettling    RoundPhase = "settling"
)

type Room struct {
	ID         string       `json:"id"`
	Code       string       `json:"code"`
	Host       string       `json:"host"`
	LastLoser  string       `json:"last_loser"`
	Players    []*Player    `json:"players"`
	Status     RoomStatus   `json:"status"`
	Round      *Round       `json:"round"`
	RoundNum   int          `json:"round_num"`
	Settings   RoomSettings `json:"settings"`
	CreatedAt  time.Time    `json:"created_at"`
	LastActive time.Time    `json:"-"`
	mu         sync.RWMutex
}

type Player struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Avatar    string `json:"avatar"`
	Dice      []int  `json:"dice,omitempty"`
	Ready     bool   `json:"ready"`
	Connected bool   `json:"connected"`
	Score     int    `json:"score"`
	Rolled    bool   `json:"rolled"`
}

type Round struct {
	Number     int        `json:"number"`
	CurrentBid *Bid       `json:"current_bid"`
	TurnIndex  int        `json:"turn_index"`
	TurnOrder  []string   `json:"turn_order"`
	Phase      RoundPhase `json:"phase"`
	BidHistory []*Bid     `json:"bid_history"`
}

type Bid struct {
	PlayerID string `json:"player_id"`
	Count    int    `json:"count"`
	Face     int    `json:"face"`
	Mode     string `json:"mode"` // "zhai" or "fei"
}

type RoomSettings struct {
	MaxPlayers    int          `json:"max_players"`
	DicePerPlayer int          `json:"dice_per_player"`
	Punishments   []Punishment `json:"punishments"`
}

type Punishment struct {
	Text   string `json:"text"`
	Level  int    `json:"level"`
	Weight int    `json:"weight"`
}

const inviteCodeChars = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"

func DefaultSettings() RoomSettings {
	return RoomSettings{
		MaxPlayers:    8,
		DicePerPlayer: 5,
		Punishments: []Punishment{
			{Text: "喝一杯", Level: 1, Weight: 40},
			{Text: "喝两杯", Level: 2, Weight: 20},
			{Text: "真心话", Level: 1, Weight: 20},
			{Text: "大冒险", Level: 2, Weight: 15},
			{Text: "连喝三杯", Level: 3, Weight: 5},
		},
	}
}

func generateRoomID() string {
	digits := make([]byte, 6)
	for i := range digits {
		digits[i] = '0' + byte(rand.Intn(10))
	}
	return string(digits)
}

func generateInviteCode() string {
	code := make([]byte, 8)
	for i := range code {
		code[i] = inviteCodeChars[rand.Intn(len(inviteCodeChars))]
	}
	return string(code)
}

func NewRoom(hostID, hostName, hostAvatar string) *Room {
	now := time.Now()
	return &Room{
		ID:   generateRoomID(),
		Code: generateInviteCode(),
		Host: hostID,
		Players: []*Player{
			{
				ID:        hostID,
				Name:      hostName,
				Avatar:    hostAvatar,
				Connected: true,
			},
		},
		Status:     StatusWaiting,
		Settings:   DefaultSettings(),
		CreatedAt:  now,
		LastActive: now,
	}
}

func (r *Room) AddPlayer(id, name, avatar string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if player already exists (rejoin)
	for _, p := range r.Players {
		if p.ID == id {
			p.Connected = true
			p.Name = name
			p.Avatar = avatar
			r.LastActive = time.Now()
			return nil
		}
	}

	if r.Status != StatusWaiting {
		return ErrGameInProgress
	}

	if len(r.Players) >= r.Settings.MaxPlayers {
		return ErrRoomFull
	}

	r.Players = append(r.Players, &Player{
		ID:        id,
		Name:      name,
		Avatar:    avatar,
		Connected: true,
	})
	r.LastActive = time.Now()
	return nil
}

func (r *Room) RemovePlayer(id string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.Status == StatusWaiting {
		for i, p := range r.Players {
			if p.ID == id {
				r.Players = append(r.Players[:i], r.Players[i+1:]...)
				break
			}
		}
	} else {
		for _, p := range r.Players {
			if p.ID == id {
				p.Connected = false
				break
			}
		}
	}
	r.LastActive = time.Now()
}

func (r *Room) SetReady(id string, ready bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, p := range r.Players {
		if p.ID == id {
			p.Ready = ready
			break
		}
	}
	r.LastActive = time.Now()
}

func (r *Room) AllReady() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if len(r.Players) < 2 {
		return false
	}
	for _, p := range r.Players {
		if !p.Ready {
			return false
		}
	}
	return true
}

func (r *Room) StartRound(firstPlayerID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.Status = StatusPlaying
	r.RoundNum++

	// Build turn order starting from firstPlayerID clockwise
	turnOrder := make([]string, 0, len(r.Players))
	startIdx := 0
	for i, p := range r.Players {
		if p.ID == firstPlayerID {
			startIdx = i
			break
		}
	}
	for i := 0; i < len(r.Players); i++ {
		idx := (startIdx + i) % len(r.Players)
		turnOrder = append(turnOrder, r.Players[idx].ID)
	}

	r.Round = &Round{
		Number:     r.RoundNum,
		TurnIndex:  0,
		TurnOrder:  turnOrder,
		Phase:      PhaseRolling,
		BidHistory: make([]*Bid, 0),
	}

	// Reset player state
	for _, p := range r.Players {
		p.Dice = nil
		p.Rolled = false
		p.Ready = false
	}

	r.LastActive = time.Now()
}

func (r *Room) RollDice(playerID string) []int {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, p := range r.Players {
		if p.ID == playerID {
			dice := make([]int, r.Settings.DicePerPlayer)
			for i := range dice {
				dice[i] = rand.Intn(6) + 1
			}
			p.Dice = dice
			p.Rolled = true
			r.LastActive = time.Now()
			return dice
		}
	}
	return nil
}

func (r *Room) AllRolled() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, p := range r.Players {
		if p.Connected && !p.Rolled {
			return false
		}
	}
	return true
}

func (r *Room) CurrentTurnPlayerID() string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.Round == nil || len(r.Round.TurnOrder) == 0 {
		return ""
	}
	return r.Round.TurnOrder[r.Round.TurnIndex]
}

func (r *Room) GetPlayer(id string) *Player {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, p := range r.Players {
		if p.ID == id {
			return p
		}
	}
	return nil
}

func (r *Room) ConnectedCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	count := 0
	for _, p := range r.Players {
		if p.Connected {
			count++
		}
	}
	return count
}

func (r *Room) IsEmpty() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.ConnectedCountLocked() == 0
}

// ConnectedCountLocked is the non-locking version for internal use.
func (r *Room) ConnectedCountLocked() int {
	count := 0
	for _, p := range r.Players {
		if p.Connected {
			count++
		}
	}
	return count
}
