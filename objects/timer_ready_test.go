package objects

import (
	"testing"
	"time"
)

func timerChar() *Character {
	return &Character{
		Flags: map[string]bool{},
		Timers: map[string]time.Time{
			"global": time.Now().Add(10 * time.Second),
			"cast":   time.Now(),
			"use":    time.Now(),
		},
	}
}

func TestTimerReadyBlockedByGlobal(t *testing.T) {
	c := timerChar()
	if ready, _ := c.TimerReady("cast"); ready {
		t.Errorf("TimerReady(cast) should be blocked while the global timer is running")
	}
	if ready, _ := c.TimerReadyIgnoreGlobal("cast"); !ready {
		t.Errorf("TimerReadyIgnoreGlobal(cast) should ignore the running global timer")
	}
}

func TestTimerReadyIgnoreGlobalStillHonorsOwnTimer(t *testing.T) {
	c := timerChar()
	c.Timers["cast"] = time.Now().Add(5 * time.Second)
	ready, msg := c.TimerReadyIgnoreGlobal("cast")
	if ready {
		t.Errorf("TimerReadyIgnoreGlobal(cast) should still block on the cast timer")
	}
	if msg == "" {
		t.Errorf("expected a wait message when the cast timer is running")
	}
}

func TestTimerReadyIgnoreGlobalBlockedByStun(t *testing.T) {
	c := timerChar()
	c.Timers["stun"] = time.Now().Add(5 * time.Second)
	if ready, _ := c.TimerReadyIgnoreGlobal("cast"); ready {
		t.Errorf("TimerReadyIgnoreGlobal(cast) should be blocked while stunned")
	}
	c.Timers["stun"] = time.Now().Add(-1 * time.Second)
	if ready, _ := c.TimerReadyIgnoreGlobal("cast"); !ready {
		t.Errorf("TimerReadyIgnoreGlobal(cast) should be ready once the stun expires")
	}
}

func TestIsRestorativeSpell(t *testing.T) {
	for _, name := range []string{"vigor", "heal", "restore", "curepoison", "remove-disease", "remove-blindness"} {
		if !IsRestorativeSpell(name) {
			t.Errorf("%s should be restorative", name)
		}
	}
	for _, name := range []string{"hurt", "fireball", "stun", "teleport"} {
		if IsRestorativeSpell(name) {
			t.Errorf("%s should not be restorative", name)
		}
	}
}

// A stance toggle answers to its own cooldown only: global, combat and stun
// timers do not hold it back.
func TestStanceReadyIgnoresOtherTimers(t *testing.T) {
	c := timerChar()
	c.Timers["combat"] = time.Now().Add(10 * time.Second)
	c.Timers["stun"] = time.Now().Add(10 * time.Second)
	if ready, msg := c.StanceReady(); !ready {
		t.Errorf("StanceReady should ignore global, combat and stun timers, got %q", msg)
	}
}

// Toggling a stance starts the cooldown, and haste does not shorten it.
func TestSetStanceTimerBlocksUntilExpired(t *testing.T) {
	c := timerChar()
	c.Flags["haste"] = true
	c.SetStanceTimer()
	ready, msg := c.StanceReady()
	if ready {
		t.Fatal("StanceReady should block right after a toggle")
	}
	if msg == "" {
		t.Error("expected a wait message while the stance cooldown is running")
	}
	c.Timers["stance"] = time.Now().Add(-1 * time.Millisecond)
	if ready, _ := c.StanceReady(); !ready {
		t.Error("StanceReady should be ready once the cooldown expires")
	}
}
