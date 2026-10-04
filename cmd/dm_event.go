package cmd

import (
	"errors"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/ArcCS/Nevermore/objects"
	"github.com/ArcCS/Nevermore/permissions"
	"github.com/ArcCS/Nevermore/text"
)

func init() {
	addHandler(questEvent{},
		"Usage:  event list | show|check <name> | create|delete <name>\n"+
			"  event preview <name>      run it for staff only, to walk it before players can\n"+
			"  event start|stop <name>   stop also ends a preview, with the real cleanup\n"+
			"  event starts <name> <YYYY-MM-DD> [HH:MM] | none     start automatically\n"+
			"  event expires <name> <YYYY-MM-DD> [HH:MM]           date alone = end of that day\n"+
			"  event message <name> <text>               shown at every login while running\n"+
			"  event room|mob|item <name> add|remove <ids>\n"+
			"      rooms are in quest mode while running; on end, mobs are despawned everywhere\n"+
			"      and items are removed where they sit permanent on the ground (loot is kept)\n"+
			"  event spawn <name> <group> rooms add|remove <ids>\n"+
			"  event spawn <name> <group> mob <mob_id> <weight>      weight 0 removes it\n"+
			"  event spawn <name> <group> rate <percent> [seconds]   chance per roll, seconds between rolls (default 60)\n"+
			"  event spawn <name> <group> stage <n>                  only rolls from stage n\n"+
			"  event spawn <name> <group> delete\n"+
			"  Stages (events start at stage 1; a stage is reached by time, DM, or script, whichever is first):\n"+
			"  event advance <name> <n>                              move a running event to stage n\n"+
			"  event stage <name> <n> at <YYYY-MM-DD> [HH:MM] | at none\n"+
			"  event stage <name> <n> message <text>                 broadcast on reaching it; becomes the login message\n"+
			"  event stage <name> <n> room add|remove <ids>          rooms join quest mode from stage n\n"+
			"  event stage <name> <n> exit add|remove <room_id> <exit>   sealed to players until stage n\n"+
			"  event stage <name> <n> delete\n"+
			"  Scripts: $IFSTAGE <name> <n> stops a script below stage n; $EVENTSTAGE <name> <n> advances it.\n"+
			"  ids: 100 101 or 100-120 or 100,105-110",
		permissions.Dungeonmaster,
		"event", "questevent")
}

type questEvent cmd

func (questEvent) process(s *state) {
	if len(s.words) == 0 {
		s.msg.Actor.SendInfo("Event what? Try HELP EVENT.")
		return
	}
	if s.words[0] == "LIST" {
		eventList(s)
		return
	}
	if len(s.words) < 2 {
		s.msg.Actor.SendBad("Which event?")
		return
	}
	name := strings.ToLower(s.words[1])
	args, input := s.words[2:], s.input[2:]

	var err error
	switch s.words[0] {
	case "SHOW":
		eventShow(s, name)
		return
	case "CHECK":
		eventCheck(s, name)
		return
	case "CREATE":
		err = objects.CreateQuestEvent(name)
	case "DELETE":
		err = objects.DeleteQuestEvent(name)
	case "START":
		err = objects.StartQuestEvent(name, false)
	case "PREVIEW":
		err = objects.StartQuestEvent(name, true)
	case "STOP":
		// Ending sweeps every room, including the one this command holds.
		if e, ok := objects.QuestEventCopy(name); !ok || !e.Active {
			err = errors.New("that event is not running")
		} else {
			go objects.EndQuestEvent(name)
		}
	case "ADVANCE":
		var stage int
		if len(args) < 1 {
			err = errors.New("advance to which stage?")
		} else if stage, err = strconv.Atoi(args[0]); err != nil {
			err = errors.New("the stage must be a number")
		} else {
			_, err = objects.AdvanceQuestEvent(name, stage, s.actor, true)
		}
	case "MESSAGE":
		msg := strings.Join(input, " ")
		err = objects.EditQuestEvent(name, func(e *objects.QuestEvent) error {
			e.Message = msg
			return nil
		})
	case "STARTS":
		var starts time.Time
		if len(args) == 0 || args[0] != "NONE" {
			starts, err = parseEventTime(args, false)
		}
		if err == nil {
			err = objects.EditQuestEvent(name, func(e *objects.QuestEvent) error {
				e.Starts = starts
				return nil
			})
		}
	case "EXPIRES":
		var expires time.Time
		if expires, err = parseEventExpiry(args); err == nil {
			err = objects.EditQuestEvent(name, func(e *objects.QuestEvent) error {
				e.Expires = expires
				return nil
			})
		}
	case "ROOM", "MOB", "ITEM":
		err = eventMembers(name, s.words[0], args)
	case "SPAWN":
		err = eventSpawn(name, args)
	case "STAGE":
		err = eventStage(name, args, input)
	default:
		s.msg.Actor.SendBad("Unknown event option. Try HELP EVENT.")
		return
	}

	if err != nil {
		s.msg.Actor.SendBad(err.Error())
		return
	}
	s.msg.Actor.SendGood("Done.")
	s.ok = true
}

func eventList(s *state) {
	names := objects.QuestEventNames()
	if len(names) == 0 {
		s.msg.Actor.SendInfo("There are no events.")
		return
	}
	for _, name := range names {
		e, _ := objects.QuestEventCopy(name)
		s.msg.Actor.SendInfo(name + " - " + eventStatus(e))
	}
}

const eventTimeFormat = "2006-01-02 15:04"

func eventStatus(e objects.QuestEvent) string {
	expires := "no expiry set"
	if !e.Expires.IsZero() {
		expires = "expires " + e.Expires.Format(eventTimeFormat)
	}
	switch {
	case e.Running(time.Now()) && e.Preview:
		return "in staff preview at stage " + strconv.Itoa(e.Stage) + ", " + expires
	case e.Running(time.Now()):
		return "running at stage " + strconv.Itoa(e.Stage) + ", " + expires
	case e.Active:
		return "expired, ending shortly"
	case !e.Starts.IsZero():
		return "starts " + e.Starts.Format(eventTimeFormat) + ", " + expires
	default:
		return "not running, " + expires
	}
}

func eventShow(s *state, name string) {
	e, ok := objects.QuestEventCopy(name)
	if !ok {
		s.msg.Actor.SendBad("There is no event by that name.")
		return
	}
	s.msg.Actor.SendInfo(e.Name + ": " + eventStatus(e))
	s.msg.Actor.SendInfo("Message: " + e.Message)
	s.msg.Actor.SendInfo("Rooms: " + compactIds(e.Rooms))
	s.msg.Actor.SendInfo("Mobs: " + namedIds("MOB", e.Mobs))
	s.msg.Actor.SendInfo("Items: " + namedIds("ITEM", e.Items))
	groups := make([]string, 0, len(e.Spawns))
	for group := range e.Spawns {
		groups = append(groups, group)
	}
	sort.Strings(groups)
	for _, group := range groups {
		g := e.Spawns[group]
		interval := g.Interval
		if interval <= 0 {
			interval = objects.DefaultSpawnInterval
		}
		var mobs []string
		for _, id := range sortedKeys(g.Mobs) {
			mobs = append(mobs, namedIds("MOB", []int{id})+" x"+strconv.Itoa(g.Mobs[id]))
		}
		from := ""
		if g.Stage > 1 {
			from = " from stage " + strconv.Itoa(g.Stage)
		}
		s.msg.Actor.SendInfo("Spawn group " + group + ": " + strconv.Itoa(g.Rate) + "% every " +
			strconv.Itoa(interval) + "s" + from + " in rooms " + compactIds(g.Rooms) + "; mobs " + strings.Join(mobs, ", "))
	}
	numbers := make([]int, 0, len(e.Stages))
	for n := range e.Stages {
		numbers = append(numbers, n)
	}
	sort.Ints(numbers)
	for _, n := range numbers {
		stage := e.Stages[n]
		line := "Stage " + strconv.Itoa(n) + ":"
		if !stage.At.IsZero() {
			line += " at " + stage.At.Format(eventTimeFormat) + ";"
		}
		if len(stage.Rooms) > 0 {
			line += " rooms " + compactIds(stage.Rooms) + ";"
		}
		for _, gate := range stage.Exits {
			line += " unseals " + gate.Exit + " in room " + strconv.Itoa(gate.Room) + ";"
		}
		s.msg.Actor.SendInfo(line)
		if stage.Message != "" {
			s.msg.Actor.SendInfo("  Message: " + stage.Message)
		}
	}
}

// eventCheck is the dry run on paper: what would go wrong, what to know, and
// the login message as players would see it.
func eventCheck(s *state, name string) {
	rep, err := objects.CheckQuestEvent(name)
	if err != nil {
		s.msg.Actor.SendBad(err.Error())
		return
	}
	e, _ := objects.QuestEventCopy(name)
	s.msg.Actor.SendInfo(e.Name + ": " + eventStatus(e))
	if len(rep.Problems) == 0 {
		s.msg.Actor.SendGood("No problems found.")
	}
	for _, p := range rep.Problems {
		s.msg.Actor.SendBad("PROBLEM: " + p)
	}
	for _, n := range rep.Notes {
		s.msg.Actor.SendInfo("Note: " + n)
	}
	if e.Message != "" {
		s.msg.Actor.SendInfo("Players will see at login:")
		s.msg.Actor.Send(text.Yellow + e.Message + text.Reset)
	}
}

// parseEventExpiry reads an expiry; a date alone means the end of that day.
func parseEventExpiry(args []string) (time.Time, error) {
	return parseEventTime(args, true)
}

// parseEventTime reads a date and optional time in the server's zone. A
// date alone means the start of that day, or its very end with endOfDay.
func parseEventTime(args []string, endOfDay bool) (time.Time, error) {
	if len(args) == 0 {
		return time.Time{}, errors.New("when? YYYY-MM-DD [HH:MM]")
	}
	if len(args) >= 2 {
		t, err := time.ParseInLocation(eventTimeFormat, args[0]+" "+args[1], time.Local)
		if err != nil {
			return time.Time{}, errors.New("couldn't read that, use YYYY-MM-DD HH:MM")
		}
		return t, nil
	}
	t, err := time.ParseInLocation("2006-01-02", args[0], time.Local)
	if err != nil {
		return time.Time{}, errors.New("couldn't read that, use YYYY-MM-DD")
	}
	if endOfDay {
		t = t.Add(24*time.Hour - time.Second)
	}
	return t, nil
}

// eventStage handles "event stage <name> <n> ...".
func eventStage(name string, args, input []string) error {
	if len(args) < 2 {
		return errors.New("which stage, and what to change?")
	}
	n, err := strconv.Atoi(args[0])
	if err != nil || n < 1 {
		return errors.New("stages are numbered from 1")
	}
	op, rest := args[1], args[2:]

	edit := func(change func(stage *objects.Stage) error) error {
		return objects.EditQuestEvent(name, func(e *objects.QuestEvent) error {
			stage, ok := e.Stages[n]
			if !ok {
				stage = &objects.Stage{}
			}
			if err := change(stage); err != nil {
				return err
			}
			e.Stages[n] = stage
			return nil
		})
	}

	switch op {
	case "DELETE":
		return objects.EditQuestEvent(name, func(e *objects.QuestEvent) error {
			if _, ok := e.Stages[n]; !ok {
				return errors.New("that stage has nothing set")
			}
			delete(e.Stages, n)
			return nil
		})
	case "AT":
		if n == 1 {
			return errors.New("stage 1 is the start; use event starts")
		}
		var at time.Time
		if len(rest) == 0 || rest[0] != "NONE" {
			if at, err = parseEventTime(rest, false); err != nil {
				return err
			}
		}
		return edit(func(stage *objects.Stage) error {
			stage.At = at
			return nil
		})
	case "MESSAGE":
		if n == 1 {
			return errors.New("stage 1's message is the event message; use event message")
		}
		msg := strings.Join(input[2:], " ")
		return edit(func(stage *objects.Stage) error {
			stage.Message = msg
			return nil
		})
	case "ROOM":
		if len(rest) < 2 || (rest[0] != "ADD" && rest[0] != "REMOVE") {
			return errors.New("add or remove which rooms?")
		}
		ids, err := parseIds(rest[1:], "ROOM")
		if err != nil {
			return err
		}
		return edit(func(stage *objects.Stage) error {
			stage.Rooms = editIds(stage.Rooms, ids, rest[0] == "ADD")
			return nil
		})
	case "EXIT":
		if len(rest) < 3 || (rest[0] != "ADD" && rest[0] != "REMOVE") {
			return errors.New("add or remove which room_id and exit?")
		}
		roomId, err := strconv.Atoi(rest[1])
		if err != nil {
			return errors.New("the room must be a number")
		}
		exit := strings.ToLower(strings.Join(rest[2:], " "))
		if room, ok := objects.Rooms[roomId]; !ok {
			return errors.New("room " + rest[1] + " doesn't exist")
		} else if _, ok := room.Exits[exit]; !ok && rest[0] == "ADD" {
			return errors.New("room " + rest[1] + " has no exit " + exit)
		}
		gate := objects.ExitGate{Room: roomId, Exit: exit}
		return edit(func(stage *objects.Stage) error {
			kept := stage.Exits[:0:0]
			for _, g := range stage.Exits {
				if g != gate {
					kept = append(kept, g)
				}
			}
			if rest[0] == "ADD" {
				kept = append(kept, gate)
			}
			stage.Exits = kept
			return nil
		})
	}
	return errors.New("stages take at, message, room, exit or delete")
}

// eventMembers handles "event room|mob|item <name> add|remove <ids>".
func eventMembers(name, kind string, args []string) error {
	if len(args) < 2 || (args[0] != "ADD" && args[0] != "REMOVE") {
		return errors.New("add or remove which ids?")
	}
	ids, err := parseIds(args[1:], kind)
	if err != nil {
		return err
	}
	return objects.EditQuestEvent(name, func(e *objects.QuestEvent) error {
		list := map[string]*[]int{"ROOM": &e.Rooms, "MOB": &e.Mobs, "ITEM": &e.Items}[kind]
		*list = editIds(*list, ids, args[0] == "ADD")
		return nil
	})
}

// eventSpawn handles "event spawn <name> <group> ...".
func eventSpawn(name string, args []string) error {
	if len(args) < 2 {
		return errors.New("which spawn group, and what to change?")
	}
	group := strings.ToLower(args[0])
	op, rest := args[1], args[2:]

	switch op {
	case "DELETE":
		return objects.EditQuestEvent(name, func(e *objects.QuestEvent) error {
			if _, ok := e.Spawns[group]; !ok {
				return errors.New("there is no spawn group by that name")
			}
			delete(e.Spawns, group)
			return nil
		})
	case "ROOMS":
		if len(rest) < 2 || (rest[0] != "ADD" && rest[0] != "REMOVE") {
			return errors.New("add or remove which rooms?")
		}
		ids, err := parseIds(rest[1:], "ROOM")
		if err != nil {
			return err
		}
		return editSpawnGroup(name, group, func(g *objects.SpawnGroup) {
			g.Rooms = editIds(g.Rooms, ids, rest[0] == "ADD")
		})
	case "MOB":
		if len(rest) < 2 {
			return errors.New("which mob id and weight?")
		}
		ids, err := parseIds(rest[:1], "MOB")
		if err != nil {
			return err
		}
		weight, err := strconv.Atoi(rest[1])
		if err != nil || weight < 0 {
			return errors.New("the weight must be 0 or more")
		}
		return editSpawnGroup(name, group, func(g *objects.SpawnGroup) {
			if weight == 0 {
				delete(g.Mobs, ids[0])
			} else {
				g.Mobs[ids[0]] = weight
			}
		})
	case "STAGE":
		if len(rest) < 1 {
			return errors.New("from which stage?")
		}
		stage, err := strconv.Atoi(rest[0])
		if err != nil || stage < 1 {
			return errors.New("stages are numbered from 1")
		}
		return editSpawnGroup(name, group, func(g *objects.SpawnGroup) {
			g.Stage = stage
		})
	case "RATE":
		if len(rest) < 1 {
			return errors.New("what percent chance per roll?")
		}
		rate, err := strconv.Atoi(rest[0])
		if err != nil || rate < 0 || rate > 100 {
			return errors.New("the rate must be 0 to 100")
		}
		interval := 0
		if len(rest) >= 2 {
			if interval, err = strconv.Atoi(rest[1]); err != nil || interval < 10 {
				return errors.New("the interval must be 10 seconds or more")
			}
		}
		return editSpawnGroup(name, group, func(g *objects.SpawnGroup) {
			g.Rate = rate
			if interval > 0 {
				g.Interval = interval
			}
		})
	}
	return errors.New("spawn groups take rooms, mob, rate, stage or delete")
}

// editSpawnGroup applies edit to a spawn group, creating it if need be.
func editSpawnGroup(name, group string, edit func(g *objects.SpawnGroup)) error {
	return objects.EditQuestEvent(name, func(e *objects.QuestEvent) error {
		g, ok := e.Spawns[group]
		if !ok {
			g = &objects.SpawnGroup{Mobs: map[int]int{}}
			e.Spawns[group] = g
		}
		if g.Mobs == nil {
			g.Mobs = map[int]int{}
		}
		edit(g)
		return nil
	})
}

// maxEventIds bounds one command's id list, so a typo'd range can't add
// half the world.
const maxEventIds = 1000

// parseIds reads ids like "100 101", "100-120" or "100,105-110" and checks
// each one exists as a room, mob or item.
func parseIds(args []string, kind string) ([]int, error) {
	var ids []int
	for _, part := range strings.Split(strings.Join(args, ","), ",") {
		if part = strings.TrimSpace(part); part == "" {
			continue
		}
		lo, hi := part, part
		if dash := strings.Index(part, "-"); dash > 0 {
			lo, hi = part[:dash], part[dash+1:]
		}
		from, err1 := strconv.Atoi(lo)
		to, err2 := strconv.Atoi(hi)
		if err1 != nil || err2 != nil || to < from {
			return nil, errors.New("couldn't read the id " + part)
		}
		if len(ids)+to-from+1 > maxEventIds {
			return nil, errors.New("that is more than " + strconv.Itoa(maxEventIds) + " ids at once")
		}
		for id := from; id <= to; id++ {
			if !idExists(kind, id) {
				// A range may cross gaps; a single id must be real.
				if from == to {
					return nil, errors.New(strings.ToLower(kind) + " " + strconv.Itoa(id) + " doesn't exist")
				}
				continue
			}
			ids = append(ids, id)
		}
	}
	if len(ids) == 0 {
		return nil, errors.New("no existing " + strings.ToLower(kind) + " ids given")
	}
	return ids, nil
}

func idExists(kind string, id int) bool {
	switch kind {
	case "ROOM":
		_, ok := objects.Rooms[id]
		return ok
	case "MOB":
		_, ok := objects.Mobs[id]
		return ok
	case "ITEM":
		_, ok := objects.Items[id]
		return ok
	}
	return false
}

// editIds adds or removes ids from list, keeping it sorted and unique.
func editIds(list, ids []int, add bool) []int {
	set := map[int]bool{}
	for _, id := range list {
		set[id] = true
	}
	for _, id := range ids {
		if add {
			set[id] = true
		} else {
			delete(set, id)
		}
	}
	out := make([]int, 0, len(set))
	for id := range set {
		out = append(out, id)
	}
	sort.Ints(out)
	return out
}

// compactIds shows a sorted id list with runs collapsed: 1-3, 7.
func compactIds(ids []int) string {
	if len(ids) == 0 {
		return "none"
	}
	sorted := append([]int(nil), ids...)
	sort.Ints(sorted)
	var parts []string
	for i := 0; i < len(sorted); {
		j := i
		for j+1 < len(sorted) && sorted[j+1] == sorted[j]+1 {
			j++
		}
		if i == j {
			parts = append(parts, strconv.Itoa(sorted[i]))
		} else {
			parts = append(parts, strconv.Itoa(sorted[i])+"-"+strconv.Itoa(sorted[j]))
		}
		i = j + 1
	}
	return strings.Join(parts, ", ")
}

// namedIds shows mob or item ids with their names; one deleted since it
// was added shows as missing.
func namedIds(kind string, ids []int) string {
	if len(ids) == 0 {
		return "none"
	}
	var parts []string
	for _, id := range ids {
		label := "(missing)"
		if kind == "MOB" && idExists(kind, id) {
			label = objects.Mobs[id].Name
		} else if kind == "ITEM" && idExists(kind, id) {
			label = objects.Items[id].Name
		}
		parts = append(parts, label+" ("+strconv.Itoa(id)+")")
	}
	return strings.Join(parts, ", ")
}

func sortedKeys(m map[int]int) []int {
	keys := make([]int, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	return keys
}
