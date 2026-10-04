package cmd

import (
	"log"
	"strconv"

	"github.com/ArcCS/Nevermore/permissions"
)

// Syntax: $GIVEGOLD amount
//
// Puts amount gold in the actor's pouch. Flavour text is the builder's job,
// with $ECHO; this only says how much arrived.
func init() {
	addHandler(scriptGiveGold{},
		"",
		permissions.Anyone,
		"$GIVEGOLD")
}

type scriptGiveGold cmd

func (scriptGiveGold) process(s *state) {
	if len(s.words) < 1 {
		log.Println("$GIVEGOLD needs an amount")
		return
	}
	amount, err := strconv.Atoi(s.words[0])
	if err != nil || amount <= 0 {
		log.Println("$GIVEGOLD amount " + s.words[0] + " is not a positive number")
		return
	}
	s.actor.Gold.Add(amount)
	s.msg.Actor.SendGood("You receive " + strconv.Itoa(amount) + " gold marks.")
	s.ok = true
}
