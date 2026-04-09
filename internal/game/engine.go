package game

import (
	"fmt"
	"math"
)

// ValidateBid validates a new bid against the previous bid according to Hong Kong Liar's Dice rules.
func ValidateBid(prev, next *Bid, playerCount, dicePerPlayer int) error {
	totalDice := playerCount * dicePerPlayer

	// Basic validation
	if next.Face < 1 || next.Face > 6 {
		return fmt.Errorf("%w: face must be 1-6", ErrInvalidBid)
	}
	if next.Count < 1 {
		return fmt.Errorf("%w: count must be positive", ErrInvalidBid)
	}
	if next.Count > totalDice {
		return fmt.Errorf("%w: count cannot exceed total dice (%d)", ErrInvalidBid, totalDice)
	}
	if next.Mode != "zhai" && next.Mode != "fei" {
		return fmt.Errorf("%w: mode must be zhai or fei", ErrInvalidBid)
	}

	// First bid is always valid
	if prev == nil {
		return nil
	}

	// Same mode: must escalate
	if prev.Mode == next.Mode {
		if next.Count > prev.Count {
			return nil // higher count, any face
		}
		if next.Count == prev.Count && next.Face > prev.Face {
			return nil // same count, higher face
		}
		return fmt.Errorf("%w: must escalate bid", ErrInvalidBid)
	}

	// fei → zhai: count can halve (ceil)
	if prev.Mode == "fei" && next.Mode == "zhai" {
		minCount := int(math.Ceil(float64(prev.Count) / 2.0))
		if next.Count > minCount {
			return nil
		}
		if next.Count == minCount {
			if next.Face >= prev.Face {
				return nil
			}
			return fmt.Errorf("%w: at half count, face must be >= previous face", ErrInvalidBid)
		}
		return fmt.Errorf("%w: zhai count too low (min %d)", ErrInvalidBid, minCount)
	}

	// zhai → fei: count must be >= prev.Count*2+1
	if prev.Mode == "zhai" && next.Mode == "fei" {
		minCount := prev.Count*2 + 1
		if next.Count > minCount {
			return nil
		}
		if next.Count == minCount {
			if next.Face >= prev.Face {
				return nil
			}
			return fmt.Errorf("%w: at minimum count, face must be >= previous face", ErrInvalidBid)
		}
		return fmt.Errorf("%w: fei count too low (min %d)", ErrInvalidBid, minCount)
	}

	return nil
}

// CountDice counts how many dice match the given face across all players.
// In fei mode, 1s are wild (count as any face) unless face=1.
// In zhai mode, only exact matches count.
func CountDice(allDice map[string][]int, face int, mode string) int {
	count := 0
	for _, dice := range allDice {
		for _, d := range dice {
			if d == face {
				count++
			} else if mode == "fei" && face != 1 && d == 1 {
				count++
			}
		}
	}
	return count
}

// ResolveChallenge determines the winner and loser of a challenge.
// Returns (winner, loser, actualCount).
func ResolveChallenge(challengerID string, bid *Bid, allDice map[string][]int) (winner, loser string, actualCount int) {
	actualCount = CountDice(allDice, bid.Face, bid.Mode)
	if actualCount >= bid.Count {
		// Bidder wins, challenger loses
		return bid.PlayerID, challengerID, actualCount
	}
	// Challenger wins, bidder loses
	return challengerID, bid.PlayerID, actualCount
}
