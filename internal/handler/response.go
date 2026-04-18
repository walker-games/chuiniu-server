package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/walker-games/chuiniu-server/internal/game"
)

// RespondError sends a structured error response with i18n code.
func RespondError(c *gin.Context, status int, code string, params map[string]interface{}) {
	body := gin.H{"code": code}
	if params != nil {
		body["params"] = params
	}
	c.JSON(status, body)
}

// RespondGameError maps a game sentinel error to an HTTP response.
func RespondGameError(c *gin.Context, status int, err error) {
	code := game.MapErrorToCode(err)
	if code == "" {
		code = game.ErrCodeInvalidRequest
	}
	RespondError(c, status, code, nil)
}
