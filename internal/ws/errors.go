package ws

import (
	"github.com/walker-games/chuiniu-server/internal/game"
)

// NewErrorMessage creates a WebSocket error message with structured code + params.
func NewErrorMessage(code string, params map[string]interface{}) Message {
	data := map[string]interface{}{"code": code}
	if params != nil {
		data["params"] = params
	}
	return NewMessage(MsgError, data)
}

// NewErrorFromGameErr maps a sentinel error to structured error message.
// Falls back to ErrCodeInvalidRequest if unmapped.
func NewErrorFromGameErr(err error) Message {
	code := game.MapErrorToCode(err)
	if code == "" {
		code = game.ErrCodeInvalidRequest
	}
	return NewErrorMessage(code, nil)
}
