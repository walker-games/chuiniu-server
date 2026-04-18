package game

import "errors"

// Event keys broadcasted over WebSocket.
// Frontend maps these to locale messages.
const (
	EvtPlayerJoined    = "event.playerJoined"
	EvtPlayerLeft      = "event.playerLeft"
	EvtHostTransferred = "event.hostTransferred"
	EvtGameStarted     = "event.gameStarted"
	EvtRoundStarted    = "event.roundStarted"
	EvtPlayerRolled    = "event.playerRolled"
	EvtAllRolled       = "event.allRolled"

	EvtBid             = "game.bid"
	EvtChallenge       = "game.challenge"
	EvtChallengeResult = "game.challengeResult"
	EvtPunishment      = "game.punishment"
	EvtRoundEnd        = "game.roundEnd"
)

// Error codes returned via REST and WS.
const (
	ErrCodeRoomNotFound       = "error.roomNotFound"
	ErrCodeRoomFull           = "error.roomFull"
	ErrCodeGameInProgress     = "error.gameInProgress"
	ErrCodeNotYourTurn        = "error.notYourTurn"
	ErrCodeInvalidBid         = "error.invalidBid"
	ErrCodeNoBidToChallenge   = "error.noBidToChallenge"
	ErrCodeNotInRollingPhase  = "error.notInRollingPhase"
	ErrCodeNotInBiddingPhase  = "error.notInBiddingPhase"
	ErrCodeRollFailed         = "error.rollFailed"
	ErrCodeUnknownMessageType = "error.unknownMessageType"
	ErrCodeUnauthorized       = "error.unauthorized"
	ErrCodeInvalidRequest     = "error.invalidRequest"
)

// MapErrorToCode converts a sentinel game error to a frontend error code.
func MapErrorToCode(err error) string {
	switch {
	case err == nil:
		return ""
	case errors.Is(err, ErrRoomFull):
		return ErrCodeRoomFull
	case errors.Is(err, ErrGameInProgress):
		return ErrCodeGameInProgress
	case errors.Is(err, ErrNotYourTurn):
		return ErrCodeNotYourTurn
	case errors.Is(err, ErrInvalidBid):
		return ErrCodeInvalidBid
	case errors.Is(err, ErrNoBidToChallenge):
		return ErrCodeNoBidToChallenge
	default:
		return ""
	}
}
