package game

import (
	"errors"
	"fmt"
	"testing"
)

func TestMapErrorToCode(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{"nil returns empty", nil, ""},
		{"ErrRoomFull direct", ErrRoomFull, ErrCodeRoomFull},
		{"ErrRoomFull wrapped", fmt.Errorf("wrap: %w", ErrRoomFull), ErrCodeRoomFull},
		{"ErrGameInProgress", ErrGameInProgress, ErrCodeGameInProgress},
		{"ErrNotYourTurn", ErrNotYourTurn, ErrCodeNotYourTurn},
		{"ErrInvalidBid wrapped", fmt.Errorf("%w: face out of range", ErrInvalidBid), ErrCodeInvalidBid},
		{"ErrNoBidToChallenge", ErrNoBidToChallenge, ErrCodeNoBidToChallenge},
		{"unmapped error returns empty", errors.New("random"), ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MapErrorToCode(tt.err)
			if got != tt.want {
				t.Errorf("MapErrorToCode() = %q, want %q", got, tt.want)
			}
		})
	}
}
