package ws

import (
	"encoding/json"
	"time"
)

type Message struct {
	Type string      `json:"type"`
	Data interface{} `json:"data,omitempty"`
	Ts   int64       `json:"ts"`
}

func NewMessage(msgType string, data interface{}) Message {
	return Message{Type: msgType, Data: data, Ts: time.Now().Unix()}
}

// Client -> Server
const (
	MsgJoin      = "join"
	MsgReady     = "ready"
	MsgRoll      = "roll"
	MsgBid       = "bid"
	MsgChallenge = "challenge"
	MsgEmoji     = "emoji"
)

// Server -> Client
const (
	MsgRoomState       = "room_state"
	MsgPlayerJoined    = "player_joined"
	MsgPlayerLeft      = "player_left"
	MsgGameStart       = "game_start"
	MsgRollResult      = "roll_result"
	MsgAllRolled       = "all_rolled"
	MsgBidMade         = "bid_made"
	MsgTurnChange      = "turn_change"
	MsgChallengeResult = "challenge_result"
	MsgPunishment      = "punishment"
	MsgRoundEnd        = "round_end"
	MsgPlayerEmoji     = "player_emoji"
	MsgError           = "error"
)

type Inbound struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

type ReadyData struct {
	Ready bool `json:"ready"`
}

type BidData struct {
	Count int    `json:"count"`
	Face  int    `json:"face"`
	Mode  string `json:"mode"`
}

type ChallengeData struct {
	TargetID string `json:"targetId"`
}

type EmojiData struct {
	Emoji string `json:"emoji"`
}
