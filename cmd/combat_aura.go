package cmd

import (
	"strconv"

	"github.com/ArcCS/Nevermore/config"
	"github.com/ArcCS/Nevermore/objects"
	"github.com/ArcCS/Nevermore/permissions"
)

func init() {
	addHandler(aura{},
		"Usage:  aura (type) \n\n Channel your focus and devotion outward, giving off one of several possible auras.",
		permissions.Paladin,
		"aura")
}

type aura cmd

func (aura) process(s *state) {
	if s.actor.Tier < config.MinorAbilityTier {
		s.msg.Actor.SendBad("You must be at least tier " + strconv.Itoa(config.MinorAbilityTier) + " to use this skill.")
		return
	}
	if len(s.input) < 1 {
		s.msg.Actor.SendBad("What aura?")
	} else if s.input[0] == "courage" {
		courage, ok := s.actor.Flags["aura-courage"]
		if ok {
			if courage {
				s.msg.Actor.SendBad("You already are inspiring courage.")
				return
			}
		}
	} else if s.input[0] == "faith" {
		faith, ok := s.actor.Flags["aura-faith"]
		if ok {
			if faith {
				s.msg.Actor.SendBad("You already are inspiring faith.")
				return
			}
		}
	} else if s.input[0] == "judgement" {
		judgement, ok := s.actor.Flags["aura-judgement"]
		if ok {
			if judgement {
				s.msg.Actor.SendBad("You already have an aura of judgement.")
			}
		}
	} else {
		s.msg.Actor.SendBad("I don't know that aura.")
		return
	}

	ready, msg := s.actor.TimerReady("combat_aura")
	if !ready {
		s.msg.Actor.SendBad(msg)
		return
	}
	ready, msg = s.actor.TimerReady("combat")
	if !ready {
		s.msg.Actor.SendBad(msg)
		return
	}

	objects.Effects["aura_"+s.input[0]](s.actor, s.actor, 0)
	s.msg.Observers.SendInfo(s.actor.Name + " has an aura of " + s.input[0] + ".")
	s.actor.SetTimer("combat_aura", 10)
	s.actor.SetTimer("combat", 3)

	s.ok = true
}
