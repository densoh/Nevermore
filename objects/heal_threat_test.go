package objects

import (
	"testing"

	"github.com/ArcCS/Nevermore/permissions"
)

func TestSplitHealThreat(t *testing.T) {
	healer := &Character{Object: Object{Name: "Healer"}, Permission: permissions.Cleric}
	mob := func(target string) *Mob {
		return &Mob{CurrentTarget: target, ThreatTable: map[string]int{}}
	}
	a, b, c, other := mob("Tank"), mob("Tank"), mob("Tank"), mob("Rogue")

	// One attacker gets the full heal.
	splitHealThreat(healer, "Tank", 90, []*Mob{a, other})
	if a.ThreatTable["Healer"] != 90 || other.ThreatTable["Healer"] != 0 {
		t.Errorf("single attacker: got %d / bystander %d", a.ThreatTable["Healer"], other.ThreatTable["Healer"])
	}

	// Three attackers split it three ways.
	a.ThreatTable = map[string]int{}
	splitHealThreat(healer, "Tank", 90, []*Mob{a, b, c, other})
	for _, m := range []*Mob{a, b, c} {
		if m.ThreatTable["Healer"] != 30 {
			t.Errorf("three attackers: got %d, want 30", m.ThreatTable["Healer"])
		}
	}
	if other.ThreatTable["Healer"] != 0 {
		t.Error("mob attacking someone else drew heal threat")
	}

	// Nobody attacking the recipient: no threat anywhere.
	splitHealThreat(healer, "Nobody", 90, []*Mob{a, other})
	if _, in := other.ThreatTable["Healer"]; in {
		t.Error("heal on an unattacked target registered threat")
	}
}
