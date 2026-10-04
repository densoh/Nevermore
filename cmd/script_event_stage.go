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
func init() {
	addHandler(scriptIfStage{},
		"",
		permissions.Anyone,
		"$IFSTAGE")
	addHandler(scriptEventStage{},
		"",
		permissions.Anyone,
		"$EVENTSTAGE")
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
