package cmd

import (
	"testing"
	"time"

	"github.com/ArcCS/Nevermore/config"
	"github.com/ArcCS/Nevermore/message"
	"github.com/ArcCS/Nevermore/objects"
)

// meditateState builds a tier 10 monk with half-empty pools whose last
// meditate was the given number of seconds ago (negative: never).
func meditateState(sinceLast int) *state {
	s := attackerState(config.MONK, nil)
	s.actor.Tier = 10
	s.actor.FlagProviders = map[string][]string{}
	s.actor.Effects = map[string]*objects.Effect{}
	s.actor.Timers = map[string]time.Time{}
	s.actor.Vit = objects.Meter{Max: 200, Current: 100}
	s.actor.Stam = objects.Meter{Max: 200, Current: 100}
	s.actor.Mana = objects.Meter{Max: 200, Current: 100}
	if sinceLast >= 0 {
		s.actor.Timers["combat_meditate"] = time.Now().Add(time.Duration(config.MeditateChargeTime-sinceLast) * time.Second)
	}
	s.msg.Actor = &message.Buffer{}
	return s
}

// trance is the seconds left on the monk's trance, zero without one.
func trance(s *state) float64 {
	if e, ok := s.actor.Effects["meditate"]; ok {
		return e.TimeRemaining()
	}
	return 0
}

// A full charge restores the whole percentage and holds the full trance.
func TestMeditateFullCharge(t *testing.T) {
	s := meditateState(-1)
	meditate{}.process(s)

	pct := config.MeditateRestorePercent(10, 0)
	if want := 100 + 200*pct/100; s.actor.Stam.Current != want {
		t.Errorf("stamina after a full meditate = %d, want %d", s.actor.Stam.Current, want)
	}
	if got, want := trance(s), float64(config.MeditateDuration(10)); got < want-2 || got > want {
		t.Errorf("full trance = %.0fs, want %.0fs", got, want)
	}
}

// Half the recharge restores half as much and holds a trance half as long.
func TestMeditateHalfCharge(t *testing.T) {
	s := meditateState(config.MeditateChargeTime / 2)
	meditate{}.process(s)

	pct := float64(config.MeditateRestorePercent(10, 0)) * 0.5
	if want := 100 + int(200*pct/100); s.actor.Stam.Current != want {
		t.Errorf("stamina after a half meditate = %d, want %d", s.actor.Stam.Current, want)
	}
	if got, want := trance(s), float64(config.MeditateDuration(10))/2; got < want-2 || got > want+1 {
		t.Errorf("half trance = %.0fs, want about %.0fs", got, want)
	}
}

// Under the minimum charge nothing happens.
func TestMeditateBelowMinimum(t *testing.T) {
	s := meditateState(config.MeditateMinCharge - 5)
	meditate{}.process(s)

	if s.actor.Stam.Current != 100 || s.actor.CheckFlag("meditate") {
		t.Error("meditate went off below the minimum charge")
	}
}

// No recasting while the trance holds, however much time has passed.
func TestMeditateNotDuringTrance(t *testing.T) {
	s := meditateState(-1)
	s.actor.Flags["meditate"] = true
	meditate{}.process(s)

	if s.actor.Stam.Current != 100 {
		t.Error("meditate went off during a trance")
	}
}

// The purge happens at any charge.
func TestMeditateMinimumChargeStillPurges(t *testing.T) {
	s := meditateState(config.MeditateMinCharge)
	s.actor.ApplyEffect("poison", "60", 0, 0, func(int) {}, func() {})
	meditate{}.process(s)

	if s.actor.HasEffect("poison") {
		t.Error("a minimum-charge meditate left the poison")
	}
}
