package objects

import (
	"testing"
	"time"

	"github.com/ArcCS/Nevermore/config"
)

func TestInCombatWindow(t *testing.T) {
	c := &Character{}
	if c.InCombat() {
		t.Error("fresh character should not be in combat")
	}
	c.MarkCombat()
	if !c.InCombat() {
		t.Error("should be in combat right after MarkCombat")
	}
	c.LastCombat = time.Now().Add(-time.Duration(config.CombatRegenWindowSeconds+1) * time.Second)
	if c.InCombat() {
		t.Error("should leave combat once the window has passed")
	}
}

func TestAddThreatDamageMarksAttackerInCombat(t *testing.T) {
	attacker := &Character{Object: Object{Name: "att"}}
	m := &Mob{ThreatTable: map[string]int{}}
	m.AddThreatDamage(10, attacker)
	if !attacker.InCombat() {
		t.Error("landing a hit should put the attacker in combat")
	}
}
