package cmd

import (
	"time"

	"github.com/ArcCS/Nevermore/config"
	"github.com/ArcCS/Nevermore/permissions"
)

func init() {
	addHandler(rescue{},
		"Usage:  rescue target \n\n Draw the attacks aimed at an ally onto yourself for a few moments. Melee attackers charge you and ranged attackers shoot you instead. Requires the seal of courage.",
		permissions.Paladin,
		"rescue")
}

type rescue cmd

func (rescue) process(s *state) {
	if len(s.input) < 1 {
		s.msg.Actor.SendBad("Rescue who?")
		return
	}
	if s.actor.CheckFlag("blind") {
		s.msg.Actor.SendBad("You can't see anything!")
		return
	}
	if s.actor.Stam.Current <= 0 {
		s.msg.Actor.SendBad("You are far too tired to do that.")
		return
	}
	if !s.actor.CheckFlag("seal-courage") {
		s.msg.Actor.SendBad("You must bear the seal of courage to rescue someone.")
		return
	}

	ready, msg := s.actor.TimerReady("combat_rescue")
	if !ready {
		s.msg.Actor.SendBad(msg)
		return
	}
	ready, msg = s.actor.TimerReady("combat")
	if !ready {
		s.msg.Actor.SendBad(msg)
		return
	}

	who := s.where.Chars.Search(s.input[0], s.actor)
	if who == nil {
		s.msg.Actor.SendInfo("Rescue who?")
		s.ok = true
		return
	}
	if who == s.actor {
		s.msg.Actor.SendBad("You can't rescue yourself.")
		return
	}

	s.actor.RunHook("combat")
	who.RescuedBy = s.actor.Name
	who.RescueUntil = time.Now().Add(time.Duration(config.RescueWindow) * time.Second)
	s.participant = who
	s.msg.Actor.SendInfo("You move to shield " + who.Name + "!")
	s.msg.Participant.SendInfo(s.actor.Name + " moves to shield you!")
	s.msg.Observers.SendInfo(s.actor.Name + " moves to shield " + who.Name + "!")
	s.actor.SetTimer("combat_rescue", config.RescueTimer)
	s.actor.SetTimer("combat", config.CombatCooldown)
	s.ok = true
}
