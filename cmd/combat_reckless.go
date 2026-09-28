package cmd

import (
	"strconv"

	"github.com/ArcCS/Nevermore/config"
	"github.com/ArcCS/Nevermore/objects"
	"github.com/ArcCS/Nevermore/permissions"
)

func init() {
	addHandler(reckless{},
		"Usage:  reckless \n\n Toggle a reckless stance.  While it lasts every melee attack you throw lands more often and hits harder, but each one costs stamina.  It drops on its own when you run too low to keep it up or draw a ranged weapon.",
		permissions.Barbarian,
		"reckless", "reck")
}

type reckless cmd

func (reckless) process(s *state) {
	if s.actor.CheckFlag("reckless") {
		s.actor.RemoveEffect("reckless")
		s.ok = true
		return
	}

	if s.actor.Tier < config.RecklessTier {
		s.msg.Actor.SendBad("You must be at least tier " + strconv.Itoa(config.RecklessTier) + " to use this skill.")
		return
	}
	if s.actor.Equipment.Main == (*objects.Item)(nil) {
		s.msg.Actor.SendBad("You have no weapon to swing.")
		return
	}
	if s.actor.Equipment.Main.ItemType == 4 {
		s.msg.Actor.SendBad("You can only swing recklessly with a melee weapon.")
		return
	}
	cost := config.RecklessStamCost(s.actor.Stam.Max)
	if s.actor.Stam.Current < cost {
		s.msg.Actor.SendBad("You are far too tired to do that.")
		return
	}

	objects.Effects["reckless"](s.actor, s.actor, 0)
	s.msg.Observers.SendInfo(s.actor.Name + " throws caution aside and swings with everything they have.")
	s.ok = true
}

// recklessAttack charges the reckless stance for one attack. Every attack
// that rolls to hit while the stance is up pays RecklessStamCost first, so
// the miss and damage rolls that read the flag never get the bonus for free.
// The stance drops on its own when the barbarian cannot cover the cost or is
// holding a ranged weapon; the flag is then already down before the roll.
func recklessAttack(s *state) {
	if !s.actor.CheckFlag("reckless") {
		return
	}
	if s.actor.Equipment.Main != (*objects.Item)(nil) && s.actor.Equipment.Main.ItemType == 4 {
		s.actor.RemoveEffect("reckless")
		s.msg.Actor.SendBad("You can't swing recklessly with a ranged weapon.")
		return
	}
	cost := config.RecklessStamCost(s.actor.Stam.Max)
	if s.actor.Stam.Current < cost {
		s.actor.RemoveEffect("reckless")
		s.msg.Actor.SendBad("You are too tired to keep swinging recklessly.")
		return
	}
	s.actor.Stam.Subtract(cost)
}
