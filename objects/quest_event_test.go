package objects

import (
	"strings"
	"testing"
	"time"

	"github.com/ArcCS/Nevermore/permissions"
	"github.com/ArcCS/Nevermore/prompt"
)

// withQuestEvents swaps in the given events and an in-memory store for the
// length of a test, bypassing the database.
func withQuestEvents(t *testing.T, events ...*QuestEvent) {
	restore := SetQuestEventStore(
		func(string, string) bool { return true },
		func(string) bool { return true },
	)
	questEventsMu.Lock()
	saved, savedRolls := questEvents, eventSpawnRolls
	questEvents, eventSpawnRolls = map[string]*QuestEvent{}, map[string]time.Time{}
	for _, e := range events {
		if e.Spawns == nil {
			e.Spawns = map[string]*SpawnGroup{}
		}
		if e.Stages == nil {
			e.Stages = map[int]*Stage{}
		}
		questEvents[e.Name] = e
	}
	reindexQuestEvents()
	questEventsMu.Unlock()
	t.Cleanup(func() {
		questEventsMu.Lock()
		questEvents, eventSpawnRolls = saved, savedRolls
		reindexQuestEvents()
		questEventsMu.Unlock()
		restore()
	})
}

func hour(n int) time.Time { return time.Now().Add(time.Duration(n) * time.Hour) }

var (
	player = &Character{Permission: permissions.Anyone | permissions.Player}
	staff  = &Character{Permission: permissions.Anyone | permissions.Dungeonmaster}
)

func TestQuestEventRoomFollowsExpiryAndStage(t *testing.T) {
	withQuestEvents(t,
		&QuestEvent{Name: "harvest", Active: true, Stage: 1, Expires: hour(1), Rooms: []int{10},
			Stages: map[int]*Stage{2: {Rooms: []int{12}}}},
		&QuestEvent{Name: "old", Active: true, Stage: 1, Expires: hour(-1), Rooms: []int{20}},
		&QuestEvent{Name: "later", Expires: hour(1), Rooms: []int{30}},
		&QuestEvent{Name: "peek", Active: true, Preview: true, Stage: 1, Expires: hour(1), Rooms: []int{40}},
	)
	for room, want := range map[int]bool{10: true, 12: false, 20: false, 30: false, 40: false} {
		if got := QuestEventRoom(room); got != want {
			t.Errorf("QuestEventRoom(%d) = %v, want %v", room, got, want)
		}
	}
	if !(&Room{RoomId: 10, Flags: map[string]bool{}}).InQuestMode() {
		t.Error("a running event's room is not in quest mode")
	}

	if _, err := AdvanceQuestEvent("harvest", 2, nil, true); err != nil {
		t.Fatal(err)
	}
	if !QuestEventRoom(12) {
		t.Error("stage 2 room didn't join quest mode on reaching stage 2")
	}
}

func TestQuestEventMessages(t *testing.T) {
	withQuestEvents(t,
		&QuestEvent{Name: "a", Active: true, Stage: 2, Expires: hour(1), Message: "The festival is on!",
			Stages: map[int]*Stage{2: {Message: "The gates are open!"}, 3: {Message: "not yet"}}},
		&QuestEvent{Name: "b", Active: true, Stage: 1, Expires: hour(-1), Message: "expired"},
		&QuestEvent{Name: "c", Expires: hour(1), Message: "not started"},
		&QuestEvent{Name: "d", Active: true, Preview: true, Stage: 1, Expires: hour(1), Message: "secret"},
	)
	got := QuestEventMessages(player)
	if len(got) != 1 || got[0] != "The gates are open!" {
		t.Errorf("player messages = %q, want the latest reached stage's", got)
	}
	got = QuestEventMessages(staff)
	if len(got) != 2 || got[1] != "[EVENT PREVIEW] secret" {
		t.Errorf("staff messages = %q, want the preview too", got)
	}
}

func TestQuestExitSealed(t *testing.T) {
	e := &QuestEvent{Name: "harvest", Expires: hour(1),
		Stages: map[int]*Stage{2: {Exits: []ExitGate{{Room: 10, Exit: "gate"}}}}}
	withQuestEvents(t, e)

	steps := []struct {
		label  string
		set    func()
		sealed bool
	}{
		{"before start", func() {}, true},
		{"in preview", func() { e.Active, e.Preview, e.Stage = true, true, 2 }, true},
		{"running, stage 1", func() { e.Preview, e.Stage = false, 1 }, true},
		{"running, stage 2", func() { e.Stage = 2 }, false},
		{"expired", func() { e.Expires = hour(-1) }, true},
	}
	for _, step := range steps {
		step.set()
		if got := QuestExitSealed(10, "GATE"); got != step.sealed {
			t.Errorf("%s: sealed = %v, want %v", step.label, got, step.sealed)
		}
	}
	if QuestExitSealed(10, "north") {
		t.Error("an ungated exit is sealed")
	}
}

func TestAdvanceQuestEvent(t *testing.T) {
	e := &QuestEvent{Name: "harvest", Active: true, Stage: 1, Expires: hour(1)}
	peek := &QuestEvent{Name: "peek", Active: true, Preview: true, Stage: 1, Expires: hour(1)}
	withQuestEvents(t, e, peek)

	if changed, err := AdvanceQuestEvent("harvest", 3, player, false); err != nil || !changed || e.Stage != 3 {
		t.Errorf("script advance to 3: changed %v err %v stage %d", changed, err, e.Stage)
	}
	if changed, _ := AdvanceQuestEvent("harvest", 2, player, false); changed || e.Stage != 3 {
		t.Error("a script moved the event back a stage")
	}
	if _, err := AdvanceQuestEvent("harvest", 2, nil, true); err != nil || e.Stage != 2 {
		t.Error("a DM couldn't move the event back a stage")
	}
	if _, err := AdvanceQuestEvent("peek", 2, player, false); err == nil || peek.Stage != 1 {
		t.Error("a player's script advanced a preview")
	}
	if _, err := AdvanceQuestEvent("peek", 2, staff, false); err != nil || peek.Stage != 2 {
		t.Error("staff couldn't advance a preview by script")
	}
}

func TestCheckQuestEventsSchedules(t *testing.T) {
	scheduled := &QuestEvent{Name: "harvest", Starts: hour(-1), Expires: hour(2)}
	staged := &QuestEvent{Name: "siege", Active: true, Stage: 1, Expires: hour(2),
		Stages: map[int]*Stage{2: {At: hour(-1)}, 3: {At: hour(-1)}, 4: {At: hour(1)}}}
	future := &QuestEvent{Name: "later", Starts: hour(1), Expires: hour(2)}
	withQuestEvents(t, scheduled, staged, future)

	CheckQuestEvents()

	if !scheduled.Active || scheduled.Preview || scheduled.Stage != 1 || !scheduled.Starts.IsZero() {
		t.Errorf("scheduled event: active %v preview %v stage %d starts %v", scheduled.Active, scheduled.Preview, scheduled.Stage, scheduled.Starts)
	}
	if staged.Stage != 3 {
		t.Errorf("staged event at stage %d, want 3 (both past times reached, not the future one)", staged.Stage)
	}
	if future.Active {
		t.Error("an event started before its time")
	}
}

func TestStartQuestEventRefusals(t *testing.T) {
	past := &QuestEvent{Name: "past", Expires: hour(-1)}
	peek := &QuestEvent{Name: "peek", Active: true, Preview: true, Stage: 1, Expires: hour(1)}
	withQuestEvents(t, past, peek)
	if err := StartQuestEvent("past", false); err == nil {
		t.Error("started an event that has already expired")
	}
	if err := StartQuestEvent("peek", false); err == nil {
		t.Error("started an event for real while it was in preview")
	}
}

func TestEventSpawnRolls(t *testing.T) {
	const mobId = 999801
	Mobs[mobId] = &Mob{Object: Object{Name: "harvest sprite"}, MobId: mobId, Flags: map[string]bool{}}
	defer delete(Mobs, mobId)
	e := &QuestEvent{
		Name: "harvest", Active: true, Stage: 1, Expires: hour(1),
		Spawns: map[string]*SpawnGroup{
			"fields": {Rooms: []int{10}, Mobs: map[int]int{mobId: 1}, Rate: 100, Interval: 30},
			"keep":   {Rooms: []int{20}, Mobs: map[int]int{mobId: 1}, Rate: 100, Stage: 2},
		},
	}
	withQuestEvents(t, e)
	room := func(id int, chars ...*Character) *Room {
		return &Room{RoomId: id, Flags: map[string]bool{}, Chars: &CharInventory{Contents: chars}, Mobs: &MobInventory{}}
	}
	now := time.Now()

	in := room(10, player)
	if got := in.eventSpawnRolls(now); len(got) != 1 || got[0].mobId != mobId || got[0].event != "harvest" {
		t.Fatalf("first roll at 100%% = %+v, want one harvest sprite", got)
	}
	if got := in.eventSpawnRolls(now.Add(10 * time.Second)); len(got) != 0 {
		t.Error("rolled again before the interval was up")
	}
	if got := in.eventSpawnRolls(now.Add(31 * time.Second)); len(got) != 1 {
		t.Error("didn't roll again after the interval")
	}
	if got := room(11, player).eventSpawnRolls(now); len(got) != 0 {
		t.Error("spawned in a room outside the group")
	}
	if got := room(20, player).eventSpawnRolls(now); len(got) != 0 {
		t.Error("a stage 2 group spawned at stage 1")
	}
	e.Stage = 2
	if got := room(20, player).eventSpawnRolls(now); len(got) != 1 {
		t.Error("a stage 2 group didn't spawn at stage 2")
	}

	e.Preview = true
	if got := room(20, player).eventSpawnRolls(now.Add(time.Hour - time.Minute)); len(got) != 0 {
		t.Error("a preview spawned with only a player present")
	}
	if got := room(20, player, staff).eventSpawnRolls(now.Add(time.Hour - time.Minute)); len(got) != 1 {
		t.Error("a preview didn't spawn with staff present")
	}
}

func TestPickEventMobSkipsWrongTimeOfDay(t *testing.T) {
	const night, day = 999802, 999803
	Mobs[night] = &Mob{MobId: night, Flags: map[string]bool{"night_only": true}}
	Mobs[day] = &Mob{MobId: day, Flags: map[string]bool{}}
	defer delete(Mobs, night)
	defer delete(Mobs, day)
	saved := DayTime
	DayTime = true
	defer func() { DayTime = saved }()

	for i := 0; i < 50; i++ {
		if id, ok := pickEventMob(map[int]int{night: 100, day: 1}); !ok || id != day {
			t.Fatalf("picked %d in daytime, want only the day mob", id)
		}
	}
	if _, ok := pickEventMob(map[int]int{night: 5}); ok {
		t.Error("picked a night mob in daytime")
	}
}

// Ending sweeps event mobs and spawned copies everywhere, and permanent
// fixtures, but leaves loot and ordinary copies of shared mobs alone.
func TestSweepQuestEvent(t *testing.T) {
	const roomId, bossId, sharedId, chestId = -81, 999811, 999812, 999813
	boss := &Mob{MobId: bossId, Flags: map[string]bool{"permanent": true}}
	spawned := &Mob{MobId: sharedId, QuestEvent: "harvest", Flags: map[string]bool{}}
	ordinary := &Mob{MobId: sharedId, Flags: map[string]bool{}}
	chest := &Item{Object: Object{Name: "chest"}, ItemId: chestId, Flags: map[string]bool{"permanent": true}}
	loot := &Item{Object: Object{Name: "chest loot"}, ItemId: chestId, Flags: map[string]bool{}}
	r := &Room{
		RoomId: roomId, Flags: map[string]bool{},
		Chars: &CharInventory{},
		Mobs:  &MobInventory{ParentId: roomId, Contents: []*Mob{boss, spawned, ordinary}},
		Items: NewItemInventory(chest, loot),
	}
	Rooms[roomId] = r
	defer delete(Rooms, roomId)
	savedPending := RoomsPendingUpdate
	defer func() { RoomsPendingUpdate = savedPending }()

	sweepQuestEvent("harvest", []int{bossId}, []int{chestId})

	if len(r.Mobs.Contents) != 1 || r.Mobs.Contents[0] != ordinary {
		t.Errorf("mobs left = %d, want only the ordinary copy", len(r.Mobs.Contents))
	}
	if len(r.Items.Contents) != 1 || r.Items.Contents[0] != loot {
		t.Errorf("items left = %d, want only the loot", len(r.Items.Contents))
	}
	found := false
	for _, id := range RoomsPendingUpdate {
		found = found || id == roomId
	}
	if !found {
		t.Error("the cleaned room was not queued for saving")
	}
}

func TestCheckQuestEventReport(t *testing.T) {
	const roomId, mobId = -82, 999821
	Rooms[roomId] = &Room{RoomId: roomId, Flags: map[string]bool{}, EncounterTable: map[int]int{mobId: 10},
		Object: Object{Commands: map[string]prompt.MenuItem{}}}
	defer delete(Rooms, roomId)
	Mobs[mobId] = &Mob{Object: Object{Name: "imp"}, MobId: mobId, Flags: map[string]bool{}}
	defer delete(Mobs, mobId)
	Rooms[roomId].AddCommands("pull", "$IFSTAGE harvest 3 ; $EVENTSTAGE harvest 2")

	withQuestEvents(t, &QuestEvent{Name: "harvest", Expires: hour(-1), Mobs: []int{mobId},
		Stages: map[int]*Stage{2: {}, 3: {}},
		Spawns: map[string]*SpawnGroup{"empty": {Rooms: []int{roomId}, Rate: 10}}})

	rep, err := CheckQuestEvent("harvest")
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"has passed", "also spawns normally", "has no mobs"}
	for _, w := range want {
		if !containsAny(rep.Problems, w) {
			t.Errorf("problems %q missing %q", rep.Problems, w)
		}
	}
	for _, w := range []string{"stage 2 is reached PULL on room -82", "only a DM can reach stage 3", "waits for stage 3"} {
		if !containsAny(rep.Notes, w) {
			t.Errorf("notes %q missing %q", rep.Notes, w)
		}
	}
}

func containsAny(lines []string, sub string) bool {
	for _, l := range lines {
		if strings.Contains(l, sub) {
			return true
		}
	}
	return false
}

// An exit leads to its redirect only while the event is live at that stage,
// and back to its own destination the moment it isn't.
func TestQuestExitRedirect(t *testing.T) {
	for _, id := range []int{-110, -111, -112, -113} {
		Rooms[id] = &Room{RoomId: id, Flags: map[string]bool{}}
		defer delete(Rooms, id)
	}
	exit := &Exit{Object: Object{Name: "Old Gate"}, ParentId: -110, ToId: -111}
	e := &QuestEvent{Name: "harvest", Expires: hour(1), Stages: map[int]*Stage{
		1: {Redirects: []ExitRedirect{{Room: -110, Exit: "old gate", To: -112}}},
		3: {Redirects: []ExitRedirect{{Room: -110, Exit: "old gate", To: -113}}},
	}}
	withQuestEvents(t, e)

	steps := []struct {
		label         string
		set           func()
		player, staff int
	}{
		{"before start", func() {}, -111, -111},
		{"in preview", func() { e.Active, e.Preview, e.Stage = true, true, 1 }, -111, -112},
		{"running, stage 1", func() { e.Preview = false }, -112, -112},
		{"running, stage 2", func() { e.Stage = 2 }, -112, -112},
		{"running, stage 3 re-points it", func() { e.Stage = 3 }, -113, -113},
		{"destination deleted", func() { delete(Rooms, -113) }, -111, -111},
		{"ended", func() { e.Active, e.Stage = false, 0 }, -111, -111},
	}
	for _, step := range steps {
		step.set()
		if got := exit.ToFor(player); got != step.player {
			t.Errorf("%s: players go to %d, want %d", step.label, got, step.player)
		}
		if got := exit.ToFor(staff); got != step.staff {
			t.Errorf("%s: staff go to %d, want %d", step.label, got, step.staff)
		}
	}
	if exit.ToId != -111 {
		t.Error("a redirect changed the exit's own destination")
	}
}

// A script's timed redirect wins over an event's, and lapses on its own.
func TestRedirectExitFor(t *testing.T) {
	for _, id := range []int{-120, -121, -122, -123} {
		Rooms[id] = &Room{RoomId: id, Flags: map[string]bool{}, Exits: map[string]*Exit{}}
		defer delete(Rooms, id)
	}
	exit := &Exit{Object: Object{Name: "rope bridge"}, ParentId: -120, ToId: -121}
	Rooms[-120].Exits["rope bridge"] = exit
	withQuestEvents(t, &QuestEvent{Name: "harvest", Active: true, Stage: 1, Expires: hour(1), Stages: map[int]*Stage{
		1: {Redirects: []ExitRedirect{{Room: -120, Exit: "rope bridge", To: -122}}},
	}})
	defer RedirectExitFor(-120, "rope bridge", 0, 0)

	roomId, name, to, d, ok := ParseExitTo([]string{"-120", "ROPE", "BRIDGE", "-123", "60"})
	if !ok || roomId != -120 || name != "rope bridge" || to != -123 || d != time.Minute {
		t.Fatalf("ParseExitTo = %d %q %d %v %v", roomId, name, to, d, ok)
	}
	for _, bad := range [][]string{
		{"-120", "rope", "bridge", "-123"},           // no seconds
		{"-120", "trapdoor", "-123", "60"},           // no such exit
		{"-120", "rope", "bridge", "-999", "60"},     // no such destination
		{"-120", "rope", "bridge", "-123", "999999"}, // too long
	} {
		if _, _, _, _, ok := ParseExitTo(bad); ok {
			t.Errorf("ParseExitTo(%v) accepted", bad)
		}
	}

	RedirectExitFor(roomId, name, to, d)
	if got := exit.To(); got != -123 {
		t.Errorf("timed redirect: leads to %d, want -123", got)
	}
	RedirectExitFor(roomId, name, to, time.Nanosecond)
	time.Sleep(time.Millisecond)
	if got := exit.To(); got != -122 {
		t.Errorf("after the timed redirect lapsed: leads to %d, want the event's -122", got)
	}
	RedirectExitFor(roomId, name, to, time.Hour)
	RedirectExitFor(roomId, name, 0, 0)
	if got := exit.To(); got != -122 {
		t.Errorf("after clearing: leads to %d, want the event's -122", got)
	}
}
