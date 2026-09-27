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
