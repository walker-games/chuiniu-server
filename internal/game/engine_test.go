package game

import (
	"testing"
)

func TestValidateBid_FirstBidAlwaysValid(t *testing.T) {
	err := ValidateBid(nil, &Bid{Count: 3, Face: 5, Mode: "fei"}, 3, 5)
	if err != nil {
		t.Errorf("first bid should be valid, got: %v", err)
	}
}

func TestValidateBid_HigherCountSameFace(t *testing.T) {
	prev := &Bid{Count: 3, Face: 5, Mode: "fei"}
	next := &Bid{Count: 4, Face: 5, Mode: "fei"}
	if err := ValidateBid(prev, next, 3, 5); err != nil {
		t.Errorf("higher count same face should be valid, got: %v", err)
	}
}

func TestValidateBid_SameCountHigherFace(t *testing.T) {
	prev := &Bid{Count: 3, Face: 3, Mode: "fei"}
	next := &Bid{Count: 3, Face: 5, Mode: "fei"}
	if err := ValidateBid(prev, next, 3, 5); err != nil {
		t.Errorf("same count higher face should be valid, got: %v", err)
	}
}

func TestValidateBid_HigherCountLowerFace(t *testing.T) {
	prev := &Bid{Count: 3, Face: 5, Mode: "fei"}
	next := &Bid{Count: 4, Face: 2, Mode: "fei"}
	if err := ValidateBid(prev, next, 3, 5); err != nil {
		t.Errorf("higher count lower face should be valid, got: %v", err)
	}
}

func TestValidateBid_SameCountLowerFace_Invalid(t *testing.T) {
	prev := &Bid{Count: 3, Face: 5, Mode: "fei"}
	next := &Bid{Count: 3, Face: 3, Mode: "fei"}
	if err := ValidateBid(prev, next, 3, 5); err == nil {
		t.Error("same count lower face should be invalid")
	}
}

func TestValidateBid_SameCountSameFace_Invalid(t *testing.T) {
	prev := &Bid{Count: 3, Face: 5, Mode: "fei"}
	next := &Bid{Count: 3, Face: 5, Mode: "fei"}
	if err := ValidateBid(prev, next, 3, 5); err == nil {
		t.Error("same count same face should be invalid")
	}
}

func TestValidateBid_FeiToZhai_HalvesCount(t *testing.T) {
	// fei count=5 → zhai min = ceil(5/2) = 3
	prev := &Bid{Count: 5, Face: 3, Mode: "fei"}
	next := &Bid{Count: 3, Face: 3, Mode: "zhai"}
	if err := ValidateBid(prev, next, 3, 5); err != nil {
		t.Errorf("fei→zhai halves count should be valid, got: %v", err)
	}
}

func TestValidateBid_FeiToZhai_BelowHalf_Invalid(t *testing.T) {
	// fei count=5 → zhai min = ceil(5/2) = 3
	prev := &Bid{Count: 5, Face: 3, Mode: "fei"}
	next := &Bid{Count: 2, Face: 5, Mode: "zhai"}
	if err := ValidateBid(prev, next, 3, 5); err == nil {
		t.Error("fei→zhai below half should be invalid")
	}
}

func TestValidateBid_ZhaiToFei_DoublePlusOne(t *testing.T) {
	// zhai count=3 → fei min = 3*2+1 = 7
	prev := &Bid{Count: 3, Face: 4, Mode: "zhai"}
	next := &Bid{Count: 7, Face: 4, Mode: "fei"}
	if err := ValidateBid(prev, next, 3, 5); err != nil {
		t.Errorf("zhai→fei double+1 should be valid, got: %v", err)
	}
}

func TestValidateBid_ZhaiToFei_BelowDoublePlusOne_Invalid(t *testing.T) {
	// zhai count=3 → fei min = 3*2+1 = 7
	prev := &Bid{Count: 3, Face: 4, Mode: "zhai"}
	next := &Bid{Count: 6, Face: 5, Mode: "fei"}
	if err := ValidateBid(prev, next, 3, 5); err == nil {
		t.Error("zhai→fei below double+1 should be invalid")
	}
}

func TestValidateBid_FaceMustBe1To6(t *testing.T) {
	if err := ValidateBid(nil, &Bid{Count: 3, Face: 0, Mode: "fei"}, 3, 5); err == nil {
		t.Error("face=0 should be invalid")
	}
	if err := ValidateBid(nil, &Bid{Count: 3, Face: 7, Mode: "fei"}, 3, 5); err == nil {
		t.Error("face=7 should be invalid")
	}
}

func TestValidateBid_CountMustBePositive(t *testing.T) {
	if err := ValidateBid(nil, &Bid{Count: 0, Face: 3, Mode: "fei"}, 3, 5); err == nil {
		t.Error("count=0 should be invalid")
	}
}

func TestValidateBid_CountCannotExceedTotalDice(t *testing.T) {
	// 3 players * 5 dice = 15 total
	if err := ValidateBid(nil, &Bid{Count: 16, Face: 3, Mode: "fei"}, 3, 5); err == nil {
		t.Error("count=16 exceeding total 15 should be invalid")
	}
}

// Test data for CountDice: p1:{1,3,5,5,2}, p2:{5,1,6,3,4}, p3:{2,2,5,4,1}
// 1s=3, 5s=4, 3s=2; fei 5=4+3=7, fei 3=2+3=5
func testAllDice() map[string][]int {
	return map[string][]int{
		"p1": {1, 3, 5, 5, 2},
		"p2": {5, 1, 6, 3, 4},
		"p3": {2, 2, 5, 4, 1},
	}
}

func TestCountDice_FeiMode_Fives(t *testing.T) {
	// fei mode 5: count 5s + 1s = 4 fives + 3 ones = 7
	allDice := testAllDice()
	count := CountDice(allDice, 5, "fei")
	if count != 7 {
		t.Errorf("fei mode 5s+1s expected 7, got %d", count)
	}
}

func TestCountDice_ZhaiMode_Fives(t *testing.T) {
	// zhai mode 5: only exact 5s = 4
	allDice := testAllDice()
	count := CountDice(allDice, 5, "zhai")
	if count != 4 {
		t.Errorf("zhai mode 5s expected 4, got %d", count)
	}
}

func TestCountDice_FeiMode_Threes(t *testing.T) {
	// fei mode 3: count 3s + 1s = 2 threes + 3 ones = 5
	allDice := testAllDice()
	count := CountDice(allDice, 3, "fei")
	if count != 5 {
		t.Errorf("fei mode 3s+1s expected 5, got %d", count)
	}
}

func TestCountDice_ZhaiMode_Threes(t *testing.T) {
	// zhai mode 3: only exact 3s = 2
	allDice := testAllDice()
	count := CountDice(allDice, 3, "zhai")
	if count != 2 {
		t.Errorf("zhai mode 3s expected 2, got %d", count)
	}
}

func TestCountDice_FeiMode_Ones(t *testing.T) {
	// fei mode 1: when face=1, no wild, just count 1s = 3
	allDice := testAllDice()
	count := CountDice(allDice, 1, "fei")
	if count != 3 {
		t.Errorf("fei mode 1s expected 3, got %d", count)
	}
}

func TestCountDice_ZhaiMode_Ones(t *testing.T) {
	// zhai mode 1: only exact 1s = 3
	allDice := testAllDice()
	count := CountDice(allDice, 1, "zhai")
	if count != 3 {
		t.Errorf("zhai mode 1s expected 3, got %d", count)
	}
}

func TestResolveChallenge_BidderWins(t *testing.T) {
	// Bid 4个5 fei, p1:{1,3,5,5,2} p2:{5,1,6,3,1} → actual=6 (3 fives+3 ones from 2 players... wait)
	// Using only p1 and p2: p1={1,3,5,5,2} p2={5,1,6,3,1}
	// fei 5: 5s={p1:2, p2:1}=3, 1s={p1:1, p2:2}=3 → total=6
	allDice := map[string][]int{
		"p1": {1, 3, 5, 5, 2},
		"p2": {5, 1, 6, 3, 1},
	}
	bid := &Bid{PlayerID: "p1", Count: 4, Face: 5, Mode: "fei"}
	winner, loser, actual := ResolveChallenge("p2", bid, allDice)
	if actual != 6 {
		t.Errorf("expected actual=6, got %d", actual)
	}
	if winner != "p1" || loser != "p2" {
		t.Errorf("bidder should win: winner=%s loser=%s", winner, loser)
	}
}

func TestResolveChallenge_ChallengerWins(t *testing.T) {
	// Bid 7个5 fei, p1:{1,3,5,5,2} p2:{5,1,6,3,1} → actual=6 < 7
	allDice := map[string][]int{
		"p1": {1, 3, 5, 5, 2},
		"p2": {5, 1, 6, 3, 1},
	}
	bid := &Bid{PlayerID: "p1", Count: 7, Face: 5, Mode: "fei"}
	winner, loser, actual := ResolveChallenge("p2", bid, allDice)
	if actual != 6 {
		t.Errorf("expected actual=6, got %d", actual)
	}
	if winner != "p2" || loser != "p1" {
		t.Errorf("challenger should win: winner=%s loser=%s", winner, loser)
	}
}
