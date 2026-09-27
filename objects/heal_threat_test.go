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

	// One attacker gets the capped half of the heal.
	splitHealThreat(healer, "Tank", 90, []*Mob{a, other}, nil)
	if a.ThreatTable["Healer"] != 45 || other.ThreatTable["Healer"] != 0 {
		t.Errorf("single attacker: got %d / bystander %d", a.ThreatTable["Healer"], other.ThreatTable["Healer"])
	}

	// Three attackers split it three ways.
	a.ThreatTable = map[string]int{}
	splitHealThreat(healer, "Tank", 90, []*Mob{a, b, c, other}, nil)
	for _, m := range []*Mob{a, b, c} {
		if m.ThreatTable["Healer"] != 30 {
			t.Errorf("three attackers: got %d, want 30", m.ThreatTable["Healer"])
		}
	}
	if other.ThreatTable["Healer"] != 0 {
		t.Error("mob attacking someone else drew heal threat")
	}

	// Nobody attacking the recipient: no threat anywhere.
	splitHealThreat(healer, "Nobody", 90, []*Mob{a, other}, nil)
	if _, in := other.ThreatTable["Healer"]; in {
		t.Error("heal on an unattacked target registered threat")
	}
}

func TestSplitHealThreatIncludesHealerVictim(t *testing.T) {
	healer := &Character{Object: Object{Name: "Healer"}, Permission: permissions.Cleric}
	mob := func(target string) *Mob {
		return &Mob{CurrentTarget: target, ThreatTable: map[string]int{}}
	}
	onTank, mine, other := mob("Tank"), mob("Rogue"), mob("Rogue")

	// The healer's own target shares the split even though it is not
	// attacking the recipient.
	splitHealThreat(healer, "Tank", 90, []*Mob{onTank, mine, other}, mine)
	if onTank.ThreatTable["Healer"] != 45 || mine.ThreatTable["Healer"] != 45 {
		t.Errorf("victim split: attacker %d, victim %d, want 45 each", onTank.ThreatTable["Healer"], mine.ThreatTable["Healer"])
	}
	if other.ThreatTable["Healer"] != 0 {
		t.Error("unrelated mob drew heal threat")
	}

	// A victim already attacking the recipient is counted once, and as the
	// only mob it takes the capped half.
	onTank.ThreatTable = map[string]int{}
	splitHealThreat(healer, "Tank", 90, []*Mob{onTank, other}, onTank)
	if onTank.ThreatTable["Healer"] != 45 {
		t.Errorf("victim attacking recipient: got %d, want 45", onTank.ThreatTable["Healer"])
	}

	// Self-heal while fighting: nothing targets the healer, but the victim
	// still takes the capped half.
	mine.ThreatTable = map[string]int{}
	splitHealThreat(healer, "Healer", 60, []*Mob{mine, other}, mine)
	if mine.ThreatTable["Healer"] != 30 {
		t.Errorf("self-heal: got %d, want 30", mine.ThreatTable["Healer"])
	}

	// A victim not in the room's mob list (stale pointer) is ignored.
	stale := mob("Tank")
	stale.ThreatTable = map[string]int{}
	splitHealThreat(healer, "Nobody", 60, []*Mob{other}, stale)
	if len(stale.ThreatTable) != 0 {
		t.Error("victim outside the room drew heal threat")
	}
}
