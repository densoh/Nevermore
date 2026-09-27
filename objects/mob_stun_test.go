package objects

import (
	"testing"
	"time"
)

func TestMobStunned(t *testing.T) {
	m := &Mob{Flags: map[string]bool{}, MobTicker: time.NewTicker(time.Hour)}
	defer m.MobTicker.Stop()

	if m.Stunned() {
		t.Errorf("fresh mob should not be stunned")
	}
	m.Stun(5)
	if !m.Stunned() {
		t.Errorf("mob should be stunned right after Stun(5)")
	}
	// Simulate the stun expiring before the tick has cleared it.
	m.StunnedUntil = time.Now().Add(-time.Second)
	if m.Stunned() {
		t.Errorf("mob should not count as stunned once StunnedUntil has passed")
	}
	// The tick clears the flags entirely.
	m.Stun(5)
	m.IsStunned = false
	m.MobStunned = 0
	if m.Stunned() {
		t.Errorf("mob should not be stunned after the tick clears the flags")
	}
}

func TestMobStunnedNoStun(t *testing.T) {
	m := &Mob{Flags: map[string]bool{"no_stun": true}, MobTicker: time.NewTicker(time.Hour)}
	defer m.MobTicker.Stop()
	m.Stun(5)
	if m.Stunned() {
		t.Errorf("no_stun mob should never be stunned")
	}
}

func TestMobStunIsAFloorOnNextTick(t *testing.T) {
	m := &Mob{Flags: map[string]bool{}, MobTicker: time.NewTicker(time.Hour)}
	defer m.MobTicker.Stop()

	// Timer has 20s left: a 5s stun would bring the mob's action forward, so it no-ops.
	m.NextTick = time.Now().Add(20 * time.Second)
	m.Stun(5)
	if m.IsStunned {
		t.Errorf("stun shorter than the time left on the timer should be a no-op")
	}

	// Timer has 3s left: a 5s stun pushes the next tick out to now+5, not now+3+5.
	m.NextTick = time.Now().Add(3 * time.Second)
	before := time.Now()
	m.Stun(5)
	if !m.IsStunned {
		t.Fatalf("stun longer than the time left on the timer should land")
	}
	got := m.NextTick.Sub(before)
	if got < 4900*time.Millisecond || got > 5100*time.Millisecond {
		t.Errorf("next tick should be ~5s out, got %v", got)
	}

	// Re-stunning for the same length while stunned does not extend it.
	first := m.NextTick
	m.Stun(5)
	if m.NextTick.Sub(first) > 50*time.Millisecond {
		t.Errorf("repeating a stun should not push the mob out further, moved by %v", m.NextTick.Sub(first))
	}

	// A longer stun does take over.
	m.Stun(10)
	got = m.NextTick.Sub(before)
	if got < 9900*time.Millisecond || got > 10100*time.Millisecond {
		t.Errorf("longer stun should replace the shorter one, got %v", got)
	}
}
