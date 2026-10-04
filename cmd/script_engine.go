package cmd

import (
	"strconv"
	"strings"

	"github.com/ArcCS/Nevermore/objects"
	menu "github.com/ArcCS/Nevermore/prompt"
	"github.com/ArcCS/Nevermore/utils"
)

// Scripts are the command strings builders attach to rooms, items and mobs
// with addCommand. A script is one or more $VERB steps separated by ";",
// for example:
//
//	$REQUIRE ITEM 4290 ; $WEAKEN 5120 812 10 8 ; $CONSUME ; $ECHOALL The chest shudders.
//
// Steps run in order through the normal dispatcher with scripting set. A
// step that refuses (see scriptEvent) stops the rest from running, so checks
// go before anything with side effects.
//
// A script is attached under a trigger. A plain word fires when a player
// types it. A word starting with "@" is an event the game fires itself:
//
//	@PUT    an item is put into this container. s.event.item is the item.
//	@DEATH  this mob was killed. The killer is the actor; see Mob.DeathCheck.

// scriptEvent is what an @EVENT script runs against. Verbs reach it through
// s.event, which is nil for scripts started by a typed trigger.
type scriptEvent struct {
	item  *objects.Item // the item the event is about, e.g. the one being PUT
	owner string        // name of what the script is attached to, for messages

	refused  bool // a step turned the event down; the triggering command stops
	consumed bool // $CONSUME destroyed item; the triggering command must not touch it
}

// refuse stops the running event script and tells the actor why.
func (ev *scriptEvent) refuse(s *state, msg string) {
	ev.refused = true
	s.msg.Actor.SendBad(msg)
}

// scriptRoomArg lists the verbs whose first argument is a room they act in.
// The engine takes those rooms' locks before running any step, because a
// handler that adds a lock makes sync rerun the whole command - and every
// step that already ran would run again.
var scriptRoomArg = map[string]bool{
	"$TELEPORTTO": true,
	"$WEAKEN":     true,
}

// scriptSteps splits a stored script into its steps.
func scriptSteps(script string) []string {
	var steps []string
	for _, step := range strings.Split(script, ";") {
		if step = strings.TrimSpace(step); step != "" {
			steps = append(steps, step)
		}
	}
	return steps
}

// validScript checks every step of script starts with a known $ verb and
// returns the first one that doesn't.
func validScript(script string) (string, bool) {
	steps := scriptSteps(script)
	if len(steps) == 0 {
		return script, false
	}
	for _, step := range steps {
		verb := strings.ToUpper(strings.Fields(step)[0])
		if _, ok := handlers[verb]; !ok || verb[0] != '$' {
			return verb, false
		}
	}
	return "", true
}

// lockScriptRooms adds the locks for every room the steps reach into and
// reports false if any were missing, in which case the caller must return
// straight away so sync can retry with them.
func (s *state) lockScriptRooms(steps []string) bool {
	ready := true
	for _, step := range steps {
		fields := strings.Fields(step)
		if len(fields) < 2 || !scriptRoomArg[strings.ToUpper(fields[0])] {
			continue
		}
		roomId, err := strconv.Atoi(fields[1])
		if err != nil {
			continue
		}
		if _, ok := objects.Rooms[roomId]; ok && !utils.IntIn(roomId, s.rLocks) {
			s.AddLocks(roomId)
			ready = false
		}
	}
	return ready
}

// runTrigger runs a script fired by a typed trigger, or by $RUN. extra is
// the player's own arguments, passed on to a single-step script the way it
// always was.
func (s *state) runTrigger(actorOnly bool, script string, extra ...string) {
	steps := scriptSteps(script)
	if len(steps) == 1 {
		steps[0] = strings.Join(append([]string{steps[0]}, extra...), " ")
	}
	if !s.lockScriptRooms(steps) {
		return
	}

	// A chain still needs somewhere to record a refusal, so a failed
	// $IFSTAGE stops the steps after it.
	ev := s.event
	if ev == nil {
		ev = &scriptEvent{}
		s.event = ev
		defer func() { s.event = nil }()
	}
	for _, step := range steps {
		s.script(true, !actorOnly, !actorOnly, step)
		if ev.refused {
			break
		}
	}
}

// runEvent runs the @event script in cmds, if there is one, against ev. It
// returns false when the script needs more room locks; the caller must then
// return without doing anything, and sync will run it again. Otherwise the
// caller checks ev.refused and ev.consumed.
func (s *state) runEvent(cmds map[string]menu.MenuItem, event string, ev *scriptEvent) bool {
	val, ok := cmds["@"+event]
	if !ok {
		return true
	}
	steps := scriptSteps(val.Command)
	if !s.lockScriptRooms(steps) {
		return false
	}

	prev := s.event
	s.event = ev
	defer func() { s.event = prev }()

	for _, step := range steps {
		s.scriptAll(step)
		if ev.refused {
			break
		}
	}
	return true
}
