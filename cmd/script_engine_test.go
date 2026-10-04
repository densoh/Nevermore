package cmd

import (
	"strings"
	"testing"
	"time"

	"github.com/ArcCS/Nevermore/objects"
	"github.com/ArcCS/Nevermore/permissions"
	menu "github.com/ArcCS/Nevermore/prompt"
)

func TestScriptSteps(t *testing.T) {
	got := scriptSteps(" $REQUIRE ITEM 1 ;$CONSUME;; $ECHO Hi there ")
	want := []string{"$REQUIRE ITEM 1", "$CONSUME", "$ECHO Hi there"}
	if strings.Join(got, "|") != strings.Join(want, "|") {
		t.Errorf("scriptSteps = %q, want %q", got, want)
	}
}

func TestValidScript(t *testing.T) {
	cases := []struct {
		script string
		bad    string
		ok     bool
	}{
		{"$ECHO hello", "", true},
		{"$require item 1 ; $consume", "", true},
		{"$ECHO hi ; LOOK", "LOOK", false},
		{"$ECHO hi ; $NOPE", "$NOPE", false},
		{" ; ", " ; ", false},
	}
	for _, c := range cases {
		bad, ok := validScript(c.script)
		if ok != c.ok || bad != c.bad {
			t.Errorf("validScript(%q) = %q, %v; want %q, %v", c.script, bad, ok, c.bad, c.ok)
		}
	}
}

// putState builds a player holding a gem, in a room with a chest whose @PUT
// script is script ("" for none), about to type "put gem chest".
func putState(t *testing.T, script string) (s *state, gem, chest *objects.Item) {
	const roomId = -71
	gem = &objects.Item{Object: objects.Object{Name: "black gem"}, ItemId: 4290, Flags: map[string]bool{}}
	chest = &objects.Item{
		Object:   objects.Object{Name: "hungry chest", Commands: map[string]menu.MenuItem{}},
		ItemType: 9, MaxUses: 5, Flags: map[string]bool{},
		Storage: objects.NewItemInventory(),
	}
	if script != "" {
		chest.AddCommands("@put", script)
	}
	actor := &objects.Character{
		Object:     objects.Object{Name: "Offerer"},
		Permission: permissions.Anyone | permissions.Player,
		ParentId:   roomId,
		Stam:       objects.Meter{Max: 10, Current: 10},
		Flags:      map[string]bool{},
		Inventory:  objects.NewItemInventory(gem),
		Equipment:  &objects.Equipment{},
	}
	room := &objects.Room{
		RoomId: roomId,
		Flags:  map[string]bool{},
		Chars:  &objects.CharInventory{Contents: []*objects.Character{actor}},
		Mobs:   &objects.MobInventory{},
		Items:  objects.NewItemInventory(chest),
	}
	objects.Rooms[roomId] = room
	t.Cleanup(func() { delete(objects.Rooms, roomId) })

	s = newState(actor, "put gem chest")
	s.msg.Allocate(s.rLocks)
	return s, gem, chest
}

func TestPutEventConsumes(t *testing.T) {
	s, gem, chest := putState(t, "$REQUIRE ITEM 4290 4291 ; $CONSUME ; $ECHO The chest Hums.")

	put{}.process(s)

	if len(chest.Storage.Contents) != 0 {
		t.Error("consumed gem ended up in the chest")
	}
	if s.actor.Inventory.Search("gem", 1) == gem {
		t.Error("consumed gem is still in the player's inventory")
	}
	if out := actorText(s); !strings.Contains(out, "The chest Hums.") {
		t.Errorf("script message missing or recased: %q", out)
	}
}

func TestPutEventRefusesWrongItem(t *testing.T) {
	s, gem, chest := putState(t, "$REQUIRE ITEM 1 ; $CONSUME ; $ECHO yum")

	put{}.process(s)

	if len(chest.Storage.Contents) != 0 {
		t.Error("refused gem ended up in the chest")
	}
	if s.actor.Inventory.Search("gem", 1) != gem {
		t.Error("refused gem left the player's inventory")
	}
	if out := actorText(s); strings.Contains(out, "yum") || !strings.Contains(out, "can't put") {
		t.Errorf("refusal should stop the script and say why: %q", out)
	}
}

func TestPutWithoutEventStillStores(t *testing.T) {
	s, gem, chest := putState(t, "")

	put{}.process(s)

	if len(chest.Storage.Contents) != 1 || chest.Storage.Contents[0] != gem {
		t.Error("a chest without a @PUT script no longer stores what is put in it")
	}
}

func TestPutEventPaysReward(t *testing.T) {
	s, _, _ := putState(t, "$REQUIRE ITEM 4290 ; $CONSUME ; $ECHO The priest thanks you for aiding his research. ; $GIVEGOLD 500")
	before := s.actor.Gold.Value

	put{}.process(s)

	if got := s.actor.Gold.Value - before; got != 500 {
		t.Errorf("reward paid %d gold, want 500", got)
	}
	out := actorText(s)
	if !strings.Contains(out, "The priest thanks you") || !strings.Contains(out, "500 gold") {
		t.Errorf("reward messages missing: %q", out)
	}
}

func TestRefusedPutPaysNothing(t *testing.T) {
	s, _, _ := putState(t, "$REQUIRE ITEM 1 ; $CONSUME ; $GIVEGOLD 500")
	before := s.actor.Gold.Value

	put{}.process(s)

	if s.actor.Gold.Value != before {
		t.Error("a refused offering still paid out")
	}
}

// startTestEvent creates a running event in memory only.
func startTestEvent(t *testing.T, name string) {
	restore := objects.SetQuestEventStore(
		func(string, string) bool { return true },
		func(string) bool { return true },
	)
	t.Cleanup(restore)
	if err := objects.CreateQuestEvent(name); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = objects.EditQuestEvent(name, func(e *objects.QuestEvent) error { e.Active = false; return nil })
		_ = objects.DeleteQuestEvent(name)
	})
	if err := objects.EditQuestEvent(name, func(e *objects.QuestEvent) error {
		e.Expires = time.Now().Add(time.Hour)
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if err := objects.StartQuestEvent(name, false); err != nil {
		t.Fatal(err)
	}
}

func TestIfStageGatesAndEventStageAdvances(t *testing.T) {
	startTestEvent(t, "offering")
	script := "$IFSTAGE offering 2 ; $REQUIRE ITEM 4290 ; $CONSUME ; $EVENTSTAGE offering 3"

	s, gem, _ := putState(t, script)
	put{}.process(s)
	if s.actor.Inventory.Search("gem", 1) != gem {
		t.Fatal("the chest took an offering before stage 2")
	}

	if _, err := objects.AdvanceQuestEvent("offering", 2, nil, true); err != nil {
		t.Fatal(err)
	}
	s, gem, _ = putState(t, script)
	put{}.process(s)
	if s.actor.Inventory.Search("gem", 1) == gem {
		t.Error("the chest refused an offering at stage 2")
	}
	if stage, live := objects.QuestEventStage("offering", s.actor); !live || stage != 3 {
		t.Errorf("after the offering the event is at stage %d (live %v), want 3", stage, live)
	}
}

func TestRunScriptStopsOnRefusal(t *testing.T) {
	s, _, _ := putState(t, "")
	s.input = strings.Fields("$IFSTAGE nosuchevent 1 ; $GIVEGOLD 500")
	before := s.actor.Gold.Value

	scriptRun{}.process(s)

	if s.actor.Gold.Value != before {
		t.Error("a $RUN chain kept going after $IFSTAGE refused")
	}
}
