package objects

import (
	"errors"
	"sort"
	"strconv"
	"strings"
	"time"
)

// QuestEventReport is what CheckQuestEvent found: Problems will make the
// event misbehave, Notes are worth knowing before it goes live.
type QuestEventReport struct {
	Problems []string
	Notes    []string
}

// stageScript is a script step naming an event: a $EVENTSTAGE that can
// reach a stage, or a $IFSTAGE that waits for one.
type stageScript struct {
	verb  string
	stage int
	where string
}

// CheckQuestEvent looks an event over against the world as it is now.
func CheckQuestEvent(name string) (QuestEventReport, error) {
	e, ok := QuestEventCopy(name)
	if !ok {
		return QuestEventReport{}, errors.New("there is no event by that name")
	}
	var rep QuestEventReport
	problem := func(msg string) { rep.Problems = append(rep.Problems, msg) }
	note := func(msg string) { rep.Notes = append(rep.Notes, msg) }
	now := time.Now()
	stamp := func(t time.Time) string { return t.Format("2006-01-02 15:04") }

	// When it runs.
	switch {
	case e.Expires.IsZero():
		problem("no expiration date is set")
	case !now.Before(e.Expires):
		problem("the expiration date " + stamp(e.Expires) + " has passed")
	}
	if !e.Starts.IsZero() {
		switch {
		case !e.Expires.IsZero() && !e.Starts.Before(e.Expires):
			problem("it is scheduled to start " + stamp(e.Starts) + ", after it expires")
		case !e.Active && !now.Before(e.Starts):
			note("its scheduled start " + stamp(e.Starts) + " has passed, so it will start within a minute")
		default:
			note("it starts automatically " + stamp(e.Starts))
		}
	} else if !e.Active {
		note("it has no scheduled start; a DM must start it")
	}
	if e.Message == "" {
		note("there is no login message")
	}

	// What it is made of.
	for _, id := range e.Rooms {
		if _, ok := Rooms[id]; !ok {
			problem("room " + strconv.Itoa(id) + " no longer exists")
		}
	}
	for _, id := range e.Mobs {
		if _, ok := Mobs[id]; !ok {
			problem("mob " + strconv.Itoa(id) + " no longer exists")
			continue
		}
		if spawnsIn := normalSpawnRooms(id); len(spawnsIn) > 0 {
			problem(mobLabel(id) + " also spawns normally (rooms " + idList(spawnsIn) +
				"); every copy in the world is removed when the event ends")
		}
	}
	for _, id := range e.Items {
		if _, ok := Items[id]; !ok {
			problem("item " + strconv.Itoa(id) + " no longer exists")
			continue
		}
		if placedIn := fixtureRooms(id); len(placedIn) == 0 {
			note(itemLabel(id) + " isn't placed as a permanent fixture anywhere; only fixtures are removed at the end")
		} else {
			note(itemLabel(id) + " is a fixture in rooms " + idList(placedIn) + " and is removed at the end")
		}
	}
	for _, other := range QuestEventNames() {
		o, _ := QuestEventCopy(other)
		if other == e.Name || !o.Active {
			continue
		}
		var shared []int
		for _, id := range e.Rooms {
			for _, oid := range o.Rooms {
				if id == oid {
					shared = append(shared, id)
				}
			}
		}
		if len(shared) > 0 {
			note("rooms " + idList(shared) + " are also in the running event " + other)
		}
	}

	// Stages and what reaches them.
	scripts := eventScripts(e.Name)
	reachable := map[int]bool{1: true}
	for _, n := range e.stageNumbers() {
		stage := e.Stages[n]
		var triggers []string
		if !stage.At.IsZero() {
			triggers = append(triggers, "at "+stamp(stage.At))
			if !e.Expires.IsZero() && !stage.At.Before(e.Expires) {
				problem("stage " + strconv.Itoa(n) + " is timed for " + stamp(stage.At) + ", after the event expires")
			}
		}
		for _, sc := range scripts {
			if sc.verb == "$EVENTSTAGE" && sc.stage >= n {
				triggers = append(triggers, sc.where)
			}
		}
		if n > 1 {
			if len(triggers) == 0 {
				note("only a DM can reach stage " + strconv.Itoa(n) + ": it has no time and no script advances to it")
			} else {
				reachable[n] = true
				note("stage " + strconv.Itoa(n) + " is reached " + strings.Join(triggers, ", or "))
			}
		}
		for _, id := range stage.Rooms {
			if _, ok := Rooms[id]; !ok {
				problem("stage " + strconv.Itoa(n) + " room " + strconv.Itoa(id) + " no longer exists")
			}
		}
		for _, gate := range stage.Exits {
			room, ok := Rooms[gate.Room]
			if !ok {
				problem("stage " + strconv.Itoa(n) + " seals an exit in room " + strconv.Itoa(gate.Room) + ", which no longer exists")
			} else if _, ok := room.Exits[gate.Exit]; !ok {
				problem("stage " + strconv.Itoa(n) + " seals exit " + gate.Exit + " in room " + strconv.Itoa(gate.Room) + ", which has no such exit")
			}
		}
	}
	for _, n := range e.stageNumbers() {
		for _, rd := range e.Stages[n].Redirects {
			label := "stage " + strconv.Itoa(n) + " redirects exit " + rd.Exit + " in room " + strconv.Itoa(rd.Room)
			if room, ok := Rooms[rd.Room]; !ok {
				problem(label + ", but that room no longer exists")
			} else if exit, ok := room.Exits[rd.Exit]; !ok {
				problem(label + ", which has no such exit")
			} else if _, ok := Rooms[rd.To]; !ok {
				problem(label + " to room " + strconv.Itoa(rd.To) + ", which no longer exists")
			} else {
				note(label + " from room " + strconv.Itoa(exit.ToId) + " to room " + strconv.Itoa(rd.To) + " until the event ends")
			}
		}
	}
	for _, sc := range scripts {
		if sc.verb == "$IFSTAGE" && sc.stage > 1 && !reachable[sc.stage] {
			note(sc.where + " waits for stage " + strconv.Itoa(sc.stage) + ", which only a DM can reach")
		}
	}

	// Spawn groups.
	groups := make([]string, 0, len(e.Spawns))
	for group := range e.Spawns {
		groups = append(groups, group)
	}
	sort.Strings(groups)
	for _, group := range groups {
		g := e.Spawns[group]
		label := "spawn group " + group
		switch {
		case len(g.Rooms) == 0:
			problem(label + " has no rooms")
		case len(g.Mobs) == 0:
			problem(label + " has no mobs")
		case g.Rate <= 0:
			problem(label + " has a 0% rate and will never spawn")
		}
		for id := range g.Mobs {
			if _, ok := Mobs[id]; !ok {
				problem(label + " mob " + strconv.Itoa(id) + " no longer exists")
			}
		}
		if g.Stage > 1 && !reachable[g.Stage] {
			note(label + " waits for stage " + strconv.Itoa(g.Stage) + ", which only a DM can reach")
		}
		if hasEncounters := encountersOff(g.Rooms); len(hasEncounters) > 0 {
			note(label + " rooms " + idList(hasEncounters) + " have encounters off; event spawns still roll there")
		}
	}
	return rep, nil
}

// eventScripts finds every $EVENTSTAGE and $IFSTAGE step naming the event
// in the scripts on rooms, item templates and mob templates.
func eventScripts(name string) []stageScript {
	var found []stageScript
	scan := func(where string, cmds map[string]string) {
		for trigger, script := range cmds {
			for _, step := range strings.Split(script, ";") {
				f := strings.Fields(step)
				if len(f) < 3 || strings.ToLower(f[1]) != name {
					continue
				}
				verb := strings.ToUpper(f[0])
				if verb != "$EVENTSTAGE" && verb != "$IFSTAGE" {
					continue
				}
				if n, err := strconv.Atoi(f[2]); err == nil {
					found = append(found, stageScript{verb, n, trigger + " on " + where})
				}
			}
		}
	}
	commands := func(o Object) map[string]string {
		out := map[string]string{}
		for k, v := range o.Commands {
			out[k] = v.Command
		}
		return out
	}
	for id, r := range Rooms {
		scan("room "+strconv.Itoa(id), commands(r.Object))
	}
	for id, i := range Items {
		scan(itemLabel(id), commands(i.Object))
	}
	for id, m := range Mobs {
		scan(mobLabel(id), commands(m.Object))
	}
	sort.Slice(found, func(a, b int) bool { return found[a].where < found[b].where })
	return found
}

func normalSpawnRooms(mobId int) []int {
	var rooms []int
	for id, r := range Rooms {
		if _, ok := r.EncounterTable[mobId]; ok {
			rooms = append(rooms, id)
		}
	}
	return rooms
}

func fixtureRooms(itemId int) []int {
	var rooms []int
	for id, r := range Rooms {
		if r.Items == nil {
			continue
		}
		for _, i := range r.Items.Contents {
			if i.ItemId == itemId && i.Flags["permanent"] {
				rooms = append(rooms, id)
				break
			}
		}
	}
	return rooms
}

func encountersOff(rooms []int) []int {
	var off []int
	for _, id := range rooms {
		if r, ok := Rooms[id]; ok && !r.Flags["encounters_on"] {
			off = append(off, id)
		}
	}
	return off
}

func mobLabel(id int) string {
	if m, ok := Mobs[id]; ok {
		return m.Name + " (mob " + strconv.Itoa(id) + ")"
	}
	return "mob " + strconv.Itoa(id)
}

func itemLabel(id int) string {
	if i, ok := Items[id]; ok {
		return i.Name + " (item " + strconv.Itoa(id) + ")"
	}
	return "item " + strconv.Itoa(id)
}

// idList shows ids sorted, comma separated.
func idList(ids []int) string {
	sorted := append([]int(nil), ids...)
	sort.Ints(sorted)
	parts := make([]string, len(sorted))
	for i, id := range sorted {
		parts[i] = strconv.Itoa(id)
	}
	return strings.Join(parts, ", ")
}
