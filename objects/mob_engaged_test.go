package objects

import (
	"testing"

	"github.com/ArcCS/Nevermore/config"
	"github.com/ArcCS/Nevermore/permissions"
)

func TestMobEngagedCount(t *testing.T) {
	const roomID = -9101
	lead := &Character{Object: Object{Name: "Lead"}, PartyFollowers: []string{"Pal", "Gone"}}
	pal := &Character{Object: Object{Name: "Pal"}}
	tank := &Character{Object: Object{Name: "Tank"}}
	idler := &Character{Object: Object{Name: "Idler"}}
	gm := &Character{Object: Object{Name: "Gm"}, Permission: permissions.Gamemaster}
	Rooms[roomID] = &Room{Chars: &CharInventory{ParentId: roomID, Contents: []*Character{lead, pal, tank, idler, gm}}}
	defer delete(Rooms, roomID)

	m := &Mob{ParentId: roomID, ThreatTable: map[string]int{}}
	if got := m.EngagedCount(idler); got != 1 {
		t.Errorf("lone character engaged = %d, want 1", got)
	}
	// Party members in the room count; "Gone" is not in the room.
	if got := m.EngagedCount(lead); got != 2 {
		t.Errorf("party engaged = %d, want 2", got)
	}
	// Threat-table names count once, absent ones and staff not at all.
	m.ThreatTable["Tank"] = 5
	m.ThreatTable["Pal"] = 5
	m.ThreatTable["Left"] = 5
	m.ThreatTable["Gm"] = 5
	if got := m.EngagedCount(lead); got != 3 {
		t.Errorf("party + threat engaged = %d, want 3", got)
	}
	if got := m.EngagedCount(idler); got != 3 {
		t.Errorf("bystander engaged = %d, want 3 (self + Tank + Pal on threat)", got)
	}
}

func TestRoomInQuestMode(t *testing.T) {
	var none *Room
	plain := &Room{Flags: map[string]bool{}}
	quest := &Room{Flags: map[string]bool{"quest_mode": true}}
	if none.InQuestMode() || plain.InQuestMode() || !quest.InQuestMode() {
		t.Fatal("room flag not honoured")
	}
	config.QuestMode = true
	defer func() { config.QuestMode = false }()
	if !none.InQuestMode() || !plain.InQuestMode() {
		t.Error("realm-wide quest mode not honoured")
	}
}
