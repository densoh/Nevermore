package cmd

import (
	"testing"

	"github.com/ArcCS/Nevermore/config"
)

// A no_instakill mob is never disintegrated: the bottom of the band, even
// the floor that lands when the turn itself would fail, becomes a turn.
func TestTurnOutcomeNoInstakill(t *testing.T) {
	const chance = 5
	band := config.TurnDisintegratePerMille(chance)
	if dis, turned := turnOutcome(1, chance, false); !dis || !turned {
		t.Errorf("roll 1: disintegrates %v turned %v, want both", dis, turned)
	}
	for _, roll := range []int{1, band} {
		if dis, turned := turnOutcome(roll, chance, true); dis || !turned {
			t.Errorf("no_instakill roll %d: disintegrates %v turned %v, want a plain turn", roll, dis, turned)
		}
	}
	if dis, turned := turnOutcome(chance*10+1, chance, true); dis || turned {
		t.Errorf("no_instakill roll past the band: disintegrates %v turned %v, want a failed turn", dis, turned)
	}
}
