package cmd

import (
	"log"
	"strconv"

	"github.com/ArcCS/Nevermore/objects"
	"github.com/ArcCS/Nevermore/permissions"
)

// Syntax: $IFSTAGE event stage
//
// Stops the script unless the event is running for the actor and has
// reached stage. Stage 1 means "while the event is on at all".
//
// Syntax: $EVENTSTAGE event stage
//
// Moves a running event forward to stage, announcing what it unlocks. It
// never moves an event back, and does nothing for players while the event
// is only a preview.
//
// Syntax: $EVENTCOUNT event target stage
//
// Event scripts: while the running event is short of stage, counts the item
// toward it and moves the event to stage once target have been counted. A
// counted item skips any $WEAKEN after it, so an offering can do research
// first and weaken a boss later. From stage on it does nothing and the
// script carries on.
func init() {
	addHandler(scriptIfStage{},
		"",
		permissions.Anyone,
		"$IFSTAGE")
	addHandler(scriptEventStage{},
		"",
		permissions.Anyone,
		"$EVENTSTAGE")
	addHandler(scriptEventCount{},
		"",
		permissions.Anyone,
		"$EVENTCOUNT")
}

type scriptIfStage cmd

func (scriptIfStage) process(s *state) {
	name, stage, ok := stageArgs(s, "$IFSTAGE")
	if ok {
		if current, live := objects.QuestEventStage(name, s.actor); live && current >= stage {
			s.ok = true
			return
		}
	}
	if s.event != nil {
		s.event.refuse(s, "Nothing happens.")
	}
}

type scriptEventStage cmd

func (scriptEventStage) process(s *state) {
	name, stage, ok := stageArgs(s, "$EVENTSTAGE")
	if !ok {
		return
	}
	// Advancing never locks rooms, so it is safe from inside a command.
	if _, err := objects.AdvanceQuestEvent(name, stage, s.actor, false); err != nil {
		log.Println("$EVENTSTAGE " + name + ": " + err.Error())
		return
	}
	s.ok = true
}

type scriptEventCount cmd

func (scriptEventCount) process(s *state) {
	refuse := func(why string) {
		log.Println("$EVENTCOUNT " + why)
		if s.event != nil {
			s.event.refuse(s, "Nothing happens.")
		}
	}
	if len(s.words) < 3 {
		refuse("needs an event name, a target and a stage")
		return
	}
	target, err1 := strconv.Atoi(s.words[1])
	stage, err2 := strconv.Atoi(s.words[2])
	if err1 != nil || err2 != nil || target < 1 || stage < 2 {
		refuse("needs a target of 1 or more and a stage from 2")
		return
	}
	counted, tally, err := objects.CountQuestEvent(s.words[0], target, stage, s.actor)
	if err != nil {
		refuse(s.words[0] + ": " + err.Error())
		return
	}
	if counted {
		if s.event != nil {
			s.event.counted = true
		}
		s.msg.Actor.SendInfo("(" + strconv.Itoa(min(tally, target)) + " of " + strconv.Itoa(target) + ")")
	}
	s.ok = true
}

func stageArgs(s *state, verb string) (string, int, bool) {
	if len(s.words) < 2 {
		log.Println(verb + " needs an event name and a stage")
		return "", 0, false
	}
	stage, err := strconv.Atoi(s.words[1])
	if err != nil || stage < 1 {
		log.Println(verb + " stage " + s.words[1] + " is not a number from 1")
		return "", 0, false
	}
	return s.words[0], stage, true
}
