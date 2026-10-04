package cmd

import (
	"strings"

	"github.com/ArcCS/Nevermore/permissions"
)

// Syntax: $RUN script
//
// Runs a whole script, steps and all, as the actor. It is how code outside a
// command - a mob's @DEATH, say - hands a stored script to objects.Script.
func init() {
	addHandler(scriptRun{},
		"",
		permissions.Anyone,
		"$RUN")
}

type scriptRun cmd

func (scriptRun) process(s *state) {
	if len(s.input) == 0 {
		return
	}
	s.runTrigger(false, strings.Join(s.input, " "))
	s.ok = true
}
