package ws

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/walker-games/chuiniu-server/internal/game"
	"github.com/walker-games/chuiniu-server/internal/service"
)

type Hub struct {
	rooms      map[string]map[string]*Client // roomID -> playerID -> Client
	mu         sync.RWMutex
	Register   chan *Client
	Unregister chan *Client
	Manager    *game.RoomManager
	LogService *service.GameLogService
}

func NewHub(manager *game.RoomManager, logService *service.GameLogService) *Hub {
	return &Hub{
		rooms:      make(map[string]map[string]*Client),
		Register:   make(chan *Client, 64),
		Unregister: make(chan *Client, 64),
		Manager:    manager,
		LogService: logService,
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.addClient(client)
		case client := <-h.Unregister:
			h.removeClient(client)
		}
	}
}

func (h *Hub) addClient(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.rooms[c.RoomID]; !ok {
		h.rooms[c.RoomID] = make(map[string]*Client)
	}
	h.rooms[c.RoomID][c.PlayerID] = c
}

func (h *Hub) removeClient(c *Client) {
	h.mu.Lock()
	clients, ok := h.rooms[c.RoomID]
	if !ok {
		h.mu.Unlock()
		return
	}
	if _, exists := clients[c.PlayerID]; exists {
		delete(clients, c.PlayerID)
		close(c.Send)
	}
	if len(clients) == 0 {
		delete(h.rooms, c.RoomID)
	}
	h.mu.Unlock()

	// Update game state
	room := h.Manager.GetRoom(c.RoomID)
	if room == nil {
		return
	}

	// During active game, just mark disconnected — don't remove player or room
	// They'll reconnect from the Game page in a moment
	if room.Status == game.StatusPlaying {
		p := room.GetPlayer(c.PlayerID)
		if p != nil {
			p.Connected = false
		}
		return
	}

	room.RemovePlayer(c.PlayerID)

	h.BroadcastToRoom(c.RoomID, NewMessage(MsgPlayerLeft, map[string]string{
		"player_id": c.PlayerID,
	}))

	// Transfer host if host left
	if c.PlayerID == room.Host && len(room.Players) > 0 {
		room.Host = room.Players[0].ID
	}

	// Always broadcast updated room state so clients refresh player list
	h.BroadcastRoomState(room)

	if room.IsEmpty() {
		h.Manager.RemoveRoom(c.RoomID)
	}
}

func (h *Hub) BroadcastToRoom(roomID string, msg Message) {
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("broadcast marshal error: %v", err)
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	clients, ok := h.rooms[roomID]
	if !ok {
		return
	}
	for _, c := range clients {
		select {
		case c.Send <- data:
		default:
			log.Printf("broadcast: buffer full for player %s", c.PlayerID)
		}
	}
}

func (h *Hub) SendToPlayer(roomID, playerID string, msg Message) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	clients, ok := h.rooms[roomID]
	if !ok {
		return
	}
	if c, exists := clients[playerID]; exists {
		c.SendMessage(msg)
	}
}

func (h *Hub) HandleMessage(c *Client, msg Inbound) {
	switch msg.Type {
	case MsgReady:
		h.handleReady(c, msg.Data)
	case MsgRoll:
		h.handleRoll(c)
	case MsgBid:
		h.handleBid(c, msg.Data)
	case MsgChallenge:
		h.handleChallenge(c, msg.Data)
	case MsgEmoji:
		h.handleEmoji(c, msg.Data)
	default:
		c.SendMessage(NewErrorMessage(game.ErrCodeUnknownMessageType, nil))
	}
}

func (h *Hub) handleReady(c *Client, data json.RawMessage) {
	var rd ReadyData
	if err := json.Unmarshal(data, &rd); err != nil {
		c.SendMessage(NewErrorMessage(game.ErrCodeInvalidRequest, nil))
		return
	}

	room := h.Manager.GetRoom(c.RoomID)
	if room == nil {
		return
	}

	room.SetReady(c.PlayerID, rd.Ready)

	// Broadcast updated room state to all players
	h.BroadcastRoomState(room)

	// Auto-start if all ready
	if room.AllReady() {
		firstPlayer := room.LastLoser
		if firstPlayer == "" {
			firstPlayer = room.Host
		}
		room.StartRound(firstPlayer)
		h.BroadcastToRoom(c.RoomID, NewMessage(MsgGameStart, map[string]interface{}{
			"round": room.RoundNum,
		}))
		h.BroadcastRoomState(room)
	}
}

func (h *Hub) handleRoll(c *Client) {
	room := h.Manager.GetRoom(c.RoomID)
	if room == nil {
		return
	}

	if room.Round == nil || room.Round.Phase != game.PhaseRolling {
		c.SendMessage(NewErrorMessage(game.ErrCodeNotInRollingPhase, nil))
		return
	}

	dice := room.RollDice(c.PlayerID)
	if dice == nil {
		c.SendMessage(NewErrorMessage(game.ErrCodeRollFailed, nil))
		return
	}

	// Send dice result only to the roller
	h.SendToPlayer(c.RoomID, c.PlayerID, NewMessage(MsgRollResult, map[string]interface{}{
		"dice": dice,
	}))

	// Check if all rolled
	if room.AllRolled() {
		room.Round.Phase = game.PhaseBidding
		h.BroadcastToRoom(c.RoomID, NewMessage(MsgAllRolled, map[string]interface{}{
			"turn_player_id": room.CurrentTurnPlayerID(),
		}))
		h.BroadcastRoomState(room)
	}
}

func (h *Hub) handleBid(c *Client, data json.RawMessage) {
	var bd BidData
	if err := json.Unmarshal(data, &bd); err != nil {
		c.SendMessage(NewErrorMessage(game.ErrCodeInvalidRequest, nil))
		return
	}

	room := h.Manager.GetRoom(c.RoomID)
	if room == nil {
		return
	}

	if room.Round == nil || room.Round.Phase != game.PhaseBidding {
		c.SendMessage(NewErrorMessage(game.ErrCodeNotInBiddingPhase, nil))
		return
	}

	if room.CurrentTurnPlayerID() != c.PlayerID {
		c.SendMessage(NewErrorMessage(game.ErrCodeNotYourTurn, nil))
		return
	}

	newBid := &game.Bid{
		PlayerID: c.PlayerID,
		Count:    bd.Count,
		Face:     bd.Face,
		Mode:     bd.Mode,
	}

	if err := game.ValidateBid(room.Round.CurrentBid, newBid, len(room.Players), room.Settings.DicePerPlayer); err != nil {
		c.SendMessage(NewErrorFromGameErr(err))
		return
	}

	// Update round state
	room.Round.CurrentBid = newBid
	room.Round.BidHistory = append(room.Round.BidHistory, newBid)
	room.Round.TurnIndex = (room.Round.TurnIndex + 1) % len(room.Round.TurnOrder)

	h.BroadcastToRoom(c.RoomID, NewMessage(MsgBidMade, map[string]interface{}{
		"player_id": c.PlayerID,
		"count":     bd.Count,
		"face":      bd.Face,
		"mode":      bd.Mode,
	}))

	h.BroadcastToRoom(c.RoomID, NewMessage(MsgTurnChange, map[string]interface{}{
		"turn_player_id": room.CurrentTurnPlayerID(),
	}))

	h.BroadcastRoomState(room)
}

func (h *Hub) handleChallenge(c *Client, data json.RawMessage) {
	room := h.Manager.GetRoom(c.RoomID)
	if room == nil {
		return
	}

	if room.Round == nil || room.Round.CurrentBid == nil {
		c.SendMessage(NewErrorMessage(game.ErrCodeNoBidToChallenge, nil))
		return
	}

	if room.CurrentTurnPlayerID() != c.PlayerID {
		c.SendMessage(NewErrorMessage(game.ErrCodeNotYourTurn, nil))
		return
	}

	// Collect all dice
	allDice := make(map[string][]int)
	for _, p := range room.Players {
		if p.Dice != nil {
			allDice[p.ID] = p.Dice
		}
	}

	bid := room.Round.CurrentBid
	log.Printf("[Challenge] bid: %d×%d mode=%s, allDice: %v", bid.Count, bid.Face, bid.Mode, allDice)
	winner, loser, actualCount := game.ResolveChallenge(c.PlayerID, bid, allDice)
	log.Printf("[Challenge] result: winner=%s loser=%s actual=%d", winner, loser, actualCount)

	// Pick punishment
	punishment := game.PickPunishment(room.Settings.Punishments)

	// Update scores
	winnerPlayer := room.GetPlayer(winner)
	loserPlayer := room.GetPlayer(loser)
	if winnerPlayer != nil {
		winnerPlayer.Score++
	}
	if loserPlayer != nil {
		loserPlayer.Score--
	}

	// Broadcast challenge result with all dice revealed
	h.BroadcastToRoom(c.RoomID, NewMessage(MsgChallengeResult, map[string]interface{}{
		"challenger":   c.PlayerID,
		"bid":          bid,
		"all_dice":     allDice,
		"actual_count": actualCount,
		"winner":       winner,
		"loser":        loser,
	}))

	h.BroadcastToRoom(c.RoomID, NewMessage(MsgPunishment, map[string]interface{}{
		"loser":           loser,
		"punishment_key":  punishment.Key,
		"punishment_text": punishment.Text,
		"level":           punishment.Level,
	}))

	h.BroadcastToRoom(c.RoomID, NewMessage(MsgRoundEnd, map[string]interface{}{
		"round":  room.RoundNum,
		"winner": winner,
		"loser":  loser,
	}))

	// Async log
	go h.LogService.SaveRound(room, allDice, c.PlayerID, winner, loser, punishment)

	// Record last loser for next round first-caller
	room.LastLoser = loser

	// Reset room to waiting
	room.Status = game.StatusWaiting
	room.Round = nil
	for _, p := range room.Players {
		p.Ready = false
		p.Dice = nil
		p.Rolled = false
	}

	h.BroadcastRoomState(room)
}

func (h *Hub) handleEmoji(c *Client, data json.RawMessage) {
	var ed EmojiData
	if err := json.Unmarshal(data, &ed); err != nil {
		c.SendMessage(NewErrorMessage(game.ErrCodeInvalidRequest, nil))
		return
	}

	h.BroadcastToRoom(c.RoomID, NewMessage(MsgPlayerEmoji, map[string]interface{}{
		"player_id": c.PlayerID,
		"emoji":     ed.Emoji,
	}))
}

func (h *Hub) BroadcastRoomState(room *game.Room) {
	h.mu.RLock()
	clients, ok := h.rooms[room.ID]
	h.mu.RUnlock()
	if !ok {
		return
	}

	for playerID, c := range clients {
		state := h.BuildRoomStateForPlayer(room, playerID)
		c.SendMessage(NewMessage(MsgRoomState, state))
	}
}

// BuildRoomStateForPlayer builds a sanitized room state for a specific viewer.
// Only the viewer's own dice are included; other players' dice are hidden.
func (h *Hub) BuildRoomStateForPlayer(room *game.Room, viewerID string) map[string]interface{} {
	players := make([]map[string]interface{}, 0, len(room.Players))
	for _, p := range room.Players {
		pData := map[string]interface{}{
			"id":        p.ID,
			"name":      p.Name,
			"avatar":    p.Avatar,
			"ready":     p.Ready,
			"connected": p.Connected,
			"score":     p.Score,
			"rolled":    p.Rolled,
		}
		if p.ID == viewerID {
			pData["dice"] = p.Dice
		}
		players = append(players, pData)
	}

	state := map[string]interface{}{
		"id":        room.ID,
		"code":      room.Code,
		"host":      room.Host,
		"status":    room.Status,
		"round_num": room.RoundNum,
		"players":   players,
		"settings":  room.Settings,
	}

	if room.Round != nil {
		state["round"] = map[string]interface{}{
			"number":      room.Round.Number,
			"phase":       room.Round.Phase,
			"current_bid": room.Round.CurrentBid,
			"turn_index":  room.Round.TurnIndex,
			"turn_order":  room.Round.TurnOrder,
			"bid_history": room.Round.BidHistory,
		}
	}

	return state
}
