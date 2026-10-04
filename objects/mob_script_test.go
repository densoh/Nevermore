package objects

import (
	"testing"
	"time"

	"github.com/ArcCS/Nevermore/permissions"
	"github.com/ArcCS/Nevermore/prompt"
	"github.com/jinzhu/copier"
)

// scriptFighter is a player who can be hit, stunned and moved by mob scripts.
func scriptFighter(name string) *Character {
	return &Character{
		Object:     Object{Name: name},
		Permission: permissions.Anyone | permissions.Player,
		Flags:      map[string]bool{},
		Modifiers:  map[string]int{},
		Timers:     map[string]time.Time{},
		Hooks:      map[string]map[string]*Hook{},
		Equipment:  &Equipment{},
		Stam:       Meter{Max: 100, Current: 100},
		Vit:        Meter{Max: 100, Current: 100},
	}
}

// scriptBoss puts a mob with the given @trigger scripts in a room with the
// given players, targeting the first of them.
func scriptBoss(t *testing.T, roomId int, scripts map[string]string, chars ...*Character) (*Mob, *Room) {
	m := &Mob{
		Object:      Object{Name: "boss", Commands: map[string]prompt.MenuItem{}},
		ParentId:    roomId,
		Flags:       map[string]bool{},
		Effects:     map[string]*Effect{},
		ThreatTable: map[string]int{},
		Stam:        Meter{Max: 1000, Current: 1000},
		NumDice:     1, SidesDice: 1, PlusDice: 99,
	}
	for trigger, script := range scripts {
		m.AddCommands(trigger, script)
	}
	r := &Room{
		RoomId: roomId, Flags: map[string]bool{},
		Chars: &CharInventory{ParentId: roomId, Contents: chars},
		Mobs:  &MobInventory{ParentId: roomId, Contents: []*Mob{m}},
		Items: NewItemInventory(),
	}
	for _, c := range chars {
		c.ParentId = roomId
	}
	if len(chars) > 0 {
		m.CurrentTarget = chars[0].Name
	}
	Rooms[roomId] = r
	t.Cleanup(func() { delete(Rooms, roomId) })
	return m, r
}

func TestValidMobScript(t *testing.T) {
	cases := []struct {
		script, bad string
		ok          bool
	}{
		{"$CHANCE 20 ; $COOLDOWN 30 ; $AOE 4d10+50 fire", "", true},
		{"$msay Fools! ; $enrage 30 50", "", true},
		{"$MSAY hi ; $GIVEGOLD 5", "$GIVEGOLD", false},
		{" ; ", " ; ", false},
	}
	for _, c := range cases {
		if bad, ok := ValidMobScript(c.script); ok != c.ok || bad != c.bad {
			t.Errorf("ValidMobScript(%q) = %q, %v; want %q, %v", c.script, bad, ok, c.bad, c.ok)
		}
	}
}

func TestRollAmount(t *testing.T) {
	if v, ok := rollAmount("120"); !ok || v != 120 {
		t.Errorf("flat amount = %d, %v", v, ok)
	}
	for i := 0; i < 50; i++ {
		if v, ok := rollAmount("4d10+50"); !ok || v < 54 || v > 90 {
			t.Fatalf("4d10+50 rolled %d, %v", v, ok)
		}
	}
	for _, bad := range []string{"", "d10", "4d", "fire", "4d0", "0d6"} {
		if _, ok := rollAmount(bad); ok {
			t.Errorf("rollAmount(%q) accepted", bad)
		}
	}
}

// Damage types go through the matching mitigation; staff are never hit.
func TestMobScriptDamageTargets(t *testing.T) {
	tank, mage := scriptFighter("Tank"), scriptFighter("Mage")
	dm := scriptFighter("Dm")
	dm.Permission = permissions.Anyone | permissions.Dungeonmaster
	m, _ := scriptBoss(t, -91, map[string]string{
		"@COMBAT": "$AOE 30 true",
		"@HP50":   "$DAMAGE target 20 true",
	}, tank, mage, dm)

	if !m.RunScript("COMBAT", nil) {
		t.Error("an $AOE didn't use the mob's turn")
	}
	if tank.Stam.Current != 70 || mage.Stam.Current != 70 {
		t.Errorf("true AoE left tank %d mage %d, want 70 each", tank.Stam.Current, mage.Stam.Current)
	}
	if dm.Stam.Current != 100 {
		t.Error("an AoE hit staff")
	}
	m.RunScript("HP50", nil)
	if tank.Stam.Current != 50 || mage.Stam.Current != 70 {
		t.Errorf("$DAMAGE target hit tank %d mage %d, want 50 and 70", tank.Stam.Current, mage.Stam.Current)
	}
}

func TestMobScriptControlSteps(t *testing.T) {
	tank := scriptFighter("Tank")
	m, _ := scriptBoss(t, -92, map[string]string{
		"@COMBAT": "$COOLDOWN 60 ; $DAMAGE target 10 true",
		"@IDLE":   "$CHANCE 0 ; $DAMAGE target 10 true",
		"@AGGRO":  "$ONCE ; $DAMAGE target 5 true",
		"@SPAWN":  "$MSAY You dare? ; $FLAG no_stun on",
	}, tank)

	if !m.RunScript("COMBAT", nil) || tank.Stam.Current != 90 {
		t.Fatalf("first run didn't act: stamina %d", tank.Stam.Current)
	}
	if m.RunScript("COMBAT", nil) || tank.Stam.Current != 90 {
		t.Error("the script ran again inside its cooldown")
	}
	if m.RunScript("IDLE", nil) || tank.Stam.Current != 90 {
		t.Error("a 0% chance passed")
	}
	m.RunScript("AGGRO", nil)
	m.RunScript("AGGRO", nil)
	if tank.Stam.Current != 85 {
		t.Errorf("$ONCE script ran more or less than once: stamina %d", tank.Stam.Current)
	}
	if m.RunScript("SPAWN", nil) {
		t.Error("talking and flags used the mob's turn")
	}
	if !m.Flags["no_stun"] {
		t.Error("$FLAG didn't set the flag")
	}
}

// Each health phase fires once, in order, and a script that acts takes the
// tick; @AGGRO fires when a target is first picked up.
func TestRunTickScripts(t *testing.T) {
	tank := scriptFighter("Tank")
	m, _ := scriptBoss(t, -93, map[string]string{
		"@HP75":   "$DAMAGE target 1 true",
		"@HP25":   "$DAMAGE target 2 true",
		"@AGGRO":  "$DAMAGE target 4 true",
		"@COMBAT": "$DAMAGE target 8 true",
	}, tank)
	lost := func() int { return 100 - tank.Stam.Current }

	if !m.runTickScripts() || lost() != 4 {
		t.Fatalf("first tick with a target: lost %d, want the aggro script's 4", lost())
	}
	if m.runTickScripts(); lost() != 12 {
		t.Errorf("second tick: lost %d, want aggro once then combat (12)", lost())
	}
	m.Stam.Current = 200 // 20%: both 75 and 25 are due, and 50 has no script
	if m.runTickScripts(); lost() != 15 {
		t.Errorf("phase tick: lost %d, want both phases and no combat script (15)", lost())
	}
	if m.runTickScripts(); lost() != 23 {
		t.Errorf("tick after phases: lost %d, want only the combat script (23)", lost())
	}
}

func TestMobScriptStunHealEnrage(t *testing.T) {
	tank := scriptFighter("Tank")
	m, _ := scriptBoss(t, -94, map[string]string{
		"@COMBAT": "$STUN target 10",
		"@HP50":   "$HEAL 25",
		"@HP25":   "$ENRAGE 30 50",
	}, tank)

	m.RunScript("COMBAT", nil)
	if ready, _ := tank.TimerReady("stun"); ready {
		t.Error("$STUN didn't stun the target")
	}

	if m.RunScript("HP50", nil) {
		t.Error("a heal at full health used the turn")
	}
	m.Stam.Current = 400
	if !m.RunScript("HP50", nil) || m.Stam.Current != 650 {
		t.Errorf("$HEAL 25 from 400/1000 gave %d, want 650", m.Stam.Current)
	}

	if base := m.InflictDamage(); base != 100 {
		t.Fatalf("base damage = %d, want 100", base)
	}
	m.RunScript("HP25", nil)
	if got := m.InflictDamage(); got != 150 {
		t.Errorf("enraged damage = %d, want 150", got)
	}
	m.RunScript("HP25", nil)
	if got := m.InflictDamage(); got != 150 {
		t.Errorf("a second enrage stacked: damage %d", got)
	}
	m.expireEffect("enrage", m.Effects["enrage"])
	if got := m.InflictDamage(); got != 100 {
		t.Errorf("damage after the rage ended = %d, want 100", got)
	}
}

func TestMobScriptHitTrigger(t *testing.T) {
	tank, rogue := scriptFighter("Tank"), scriptFighter("Rogue")
	m, _ := scriptBoss(t, -95, map[string]string{
		"@HIT": "$DAMAGE attacker 7 true ; $MTELEPORT attacker -96",
	}, tank, rogue)
	Rooms[-96] = &Room{RoomId: -96, Flags: map[string]bool{}, Chars: &CharInventory{ParentId: -96}, Mobs: &MobInventory{}}
	defer delete(Rooms, -96)

	m.AddThreatDamage(50, rogue)

	if rogue.Stam.Current != 93 || tank.Stam.Current != 100 {
		t.Errorf("thorns hit rogue %d tank %d, want only the attacker", rogue.Stam.Current, tank.Stam.Current)
	}
	if rogue.ParentId != -95 {
		t.Error("a @HIT script teleported the attacker out from under their own command")
	}
	m.AddThreatDamage(0, rogue)
	if rogue.Stam.Current != 93 {
		t.Error("@HIT fired on a hit that did no damage")
	}
}

func TestMobScriptTeleport(t *testing.T) {
	tank := scriptFighter("Tank")
	m, r := scriptBoss(t, -97, map[string]string{"@COMBAT": "$MTELEPORT target -98"}, tank)
	dest := &Room{RoomId: -98, Flags: map[string]bool{}, Chars: &CharInventory{ParentId: -98, Contents: []*Character{scriptFighter("Other")}}, Mobs: &MobInventory{}}
	Rooms[-98] = dest
	defer delete(Rooms, -98)

	dest.Lock()
	if m.RunScript("COMBAT", nil) || tank.ParentId != -97 {
		t.Error("teleported into a room that was locked by someone else")
	}
	dest.Unlock()

	if !m.RunScript("COMBAT", nil) {
		t.Fatal("the teleport didn't happen")
	}
	if tank.ParentId != -98 || len(r.Chars.Contents) != 0 || len(dest.Chars.Contents) != 2 {
		t.Errorf("after teleport: parent %d, %d left behind, %d at destination", tank.ParentId, len(r.Chars.Contents), len(dest.Chars.Contents))
	}
}

// A spawned copy carries its template's scripts but none of another copy's
// cooldowns or fired phases.
func TestMobCopyGetsScriptsNotState(t *testing.T) {
	tank := scriptFighter("Tank")
	template, _ := scriptBoss(t, -99, map[string]string{"@COMBAT": "$ONCE ; $DAMAGE target 10 true"}, tank)
	template.RunScript("COMBAT", nil)
	if tank.Stam.Current != 90 {
		t.Fatalf("setup: stamina %d", tank.Stam.Current)
	}

	spawned := Mob{}
	if err := copier.CopyWithOption(&spawned, template, copier.Option{DeepCopy: true}); err != nil {
		t.Fatal(err)
	}
	if spawned.scripts != nil {
		t.Fatal("the copy shares or inherited the original's script state")
	}
	if !spawned.RunScript("COMBAT", nil) || tank.Stam.Current != 80 {
		t.Errorf("the copy's $ONCE script didn't run fresh: stamina %d", tank.Stam.Current)
	}
}

// A script can't start another of the mob's scripts from inside itself.
func TestMobScriptNotReentrant(t *testing.T) {
	tank := scriptFighter("Tank")
	m, _ := scriptBoss(t, -100, map[string]string{"@HIT": "$DAMAGE attacker 5 true"}, tank)
	m.scriptState().running = true
	m.AddThreatDamage(10, tank)
	if tank.Stam.Current != 100 {
		t.Error("@HIT ran while another of the mob's scripts was running")
	}
}

// A permanent mob past its wander count still gets its @IDLE script.
func TestPermanentMobIdleScriptStillTicks(t *testing.T) {
	m, r := scriptBoss(t, -101, map[string]string{"@IDLE": "$FLAG woke on"})
	r.Chars.Contents = []*Character{}
	m.Flags["permanent"] = true
	m.NumWander, m.TicksAlive = 5, 50

	m.Tick()

	if !m.Flags["woke"] {
		t.Error("@IDLE didn't run for a long-standing permanent mob")
	}
}

// The wander count doesn't apply to permanent mobs: a hostile one that has
// stood idle far longer still picks up a player and gets on with its turn.
func TestPermanentHostileMobAggrosPastWanderCount(t *testing.T) {
	tank := scriptFighter("Tank")
	m, _ := scriptBoss(t, -102, map[string]string{"@AGGRO": "$DAMAGE target 5 true"}, tank)
	m.CurrentTarget = ""
	m.Flags["permanent"], m.Flags["hostile"] = true, true
	m.NumWander, m.TicksAlive = 5, 500

	m.Tick()

	if m.CurrentTarget != "Tank" {
		t.Fatalf("target = %q, want the player in the room", m.CurrentTarget)
	}
	if tank.Stam.Current != 95 {
		t.Errorf("@AGGRO didn't fire on the pickup: stamina %d", tank.Stam.Current)
	}
}
