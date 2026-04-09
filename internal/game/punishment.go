package game

import "math/rand"

// PickPunishment selects a random punishment from the pool using weighted random selection.
// If the pool is empty, returns a default "喝一杯" punishment.
func PickPunishment(pool []Punishment) Punishment {
	if len(pool) == 0 {
		return Punishment{Text: "喝一杯", Level: 1, Weight: 100}
	}

	totalWeight := 0
	for _, p := range pool {
		totalWeight += p.Weight
	}

	r := rand.Intn(totalWeight)
	cumulative := 0
	for _, p := range pool {
		cumulative += p.Weight
		if r < cumulative {
			return p
		}
	}

	// Fallback (should not reach here)
	return pool[len(pool)-1]
}
