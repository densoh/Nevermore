package objects

import (
	"testing"

	"github.com/ArcCS/Nevermore/permissions"
)

func TestMobAttackers(t *testing.T) {
	barb := &Character{Object: Object{Name: "Barb"}, Permission: permissions.Barbarian}
	healer := &Character{Object: Object{Name: "Healer"}, Permission: permissions.Cleric}
	m := &Mob{ThreatTable: map[string]int{}}

	if m.AttackedBy("Barb") {
		t.Error("fresh mob reports an attacker")
	}

	// Heal threat, rescues and the mob's own targeting add threat only.
	m.addThreat(10, healer)
	if m.AttackedBy("Healer") || m.ThreatTable["Healer"] != 10 {
		t.Errorf("support threat marked an attacker or lost threat: %v %d", m.AttackedBy("Healer"), m.ThreatTable["Healer"])
	}

	// Offensive threat marks the attacker, even at zero threat.
	m.AddThreatDamage(0, barb)
	if !m.AttackedBy("Barb") {
		t.Error("offensive threat did not mark the attacker")
	}

	// Marking directly works on a mob whose map was never made.
	fresh := &Mob{}
	fresh.MarkAttackedBy(barb)
	if !fresh.AttackedBy("Barb") {
		t.Error("MarkAttackedBy on a nil map did not record the attacker")
	}
}
