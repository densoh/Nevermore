package cmd

import (
	"strconv"

	"github.com/ArcCS/Nevermore/config"
	"github.com/ArcCS/Nevermore/objects"
	"github.com/ArcCS/Nevermore/permissions"
)

func init() {
	addHandler(flurry{},
		"Usage:  flurry \n\n Toggle a flurry stance.  While it lasts your attacks rain extra strikes, more as your hand-to-hand mastery grows, but the stance draws on your chi with every round and your blows build none.  It drops on its own when your chi runs dry.",
		permissions.Monk,
		"flurry")
}

type flurry cmd

func (flurry) process(s *state) {
	if s.actor.CheckFlag("flurry") {
		s.actor.RemoveEffect("flurry")
		s.msg.Actor.SendInfo("You let the flurry subside.")
		s.ok = true
		return
	}

	if s.actor.Stam.Current <= 0 {
		s.msg.Actor.SendBad("You are far too tired to do that.")
		return
	}

	if s.actor.Tier < config.MonkFlurryTier {
		s.msg.Actor.SendBad("You must be at least tier " + strconv.Itoa(config.MonkFlurryTier) + " to use this skill.")
		return
	}

	cost := config.FlurryChiCost(attackSkillLevel(s))
	if s.actor.Mana.Current < cost {
		s.msg.Actor.SendBad("You need at least " + strconv.Itoa(cost) + " chi to begin a flurry.")
		return
	}

	objects.Effects["flurry"](s.actor, s.actor, 0)
	s.msg.Observers.SendInfo(s.actor.Name + "'s hands blur into a flurry of blows.")
	s.ok = true
}
