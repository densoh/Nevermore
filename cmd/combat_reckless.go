package cmd

import (
	"strconv"

	"github.com/ArcCS/Nevermore/config"
	"github.com/ArcCS/Nevermore/objects"
	"github.com/ArcCS/Nevermore/permissions"
	"github.com/ArcCS/Nevermore/text"
)

func init() {
	addHandler(reckless{},
		"Usage:  reckless target # \n\n Throw everything into one swing: it lands more often and hits harder, but costs stamina.",
		permissions.Barbarian,
		"reckless", "reck")
}

type reckless cmd

func (reckless) process(s *state) {
	if len(s.input) < 1 {
		s.msg.Actor.SendBad("Swing recklessly at what exactly?")
		return
	}
	if s.actor.CheckFlag("blind") {
		s.msg.Actor.SendBad("You can't see anything!")
		return
	}
	if s.actor.Tier < config.RecklessTier {
		s.msg.Actor.SendBad("You must be at least tier " + strconv.Itoa(config.RecklessTier) + " to use this skill.")
		return
	}
	cost := config.RecklessStamCost(s.actor.Stam.Max)
	if s.actor.Stam.Current <= 0 || s.actor.Stam.Current < cost {
		s.msg.Actor.SendBad("You are far too tired to do that.")
		return
	}
	ready, msg := s.actor.TimerReady("combat")
	if !ready {
		s.msg.Actor.SendBad(msg)
		return
	}

	name := s.input[0]
	nameNum := 1
	if len(s.words) > 1 {
		if val, err := strconv.Atoi(s.words[1]); err == nil {
			nameNum = val
		}
	}

	whatMob := s.where.Mobs.Search(name, nameNum, s.actor)
	if whatMob == nil {
		s.msg.Actor.SendInfo("Swing recklessly at what?")
		s.ok = true
		return
	}

	if s.actor.Equipment.Main == (*objects.Item)(nil) {
		s.msg.Actor.SendBad("You have no weapon to attack with.")
		return
	}
	if s.actor.Equipment.Main.ItemType == 4 {
		s.msg.Actor.SendBad("You can only swing recklessly with a melee weapon.")
		return
	}
	if inRange, rangeMsg := weaponInRange(s, s.actor.Equipment.Main, whatMob); !inRange {
		s.msg.Actor.SendBad(rangeMsg)
		return
	}

	s.actor.Victim = whatMob
	s.actor.RunHook("combat")
	if _, engaged := whatMob.ThreatTable[s.actor.Name]; !engaged {
		s.msg.Actor.Send(text.White + "You engaged " + whatMob.Name + " #" + strconv.Itoa(s.where.Mobs.GetNumber(whatMob)) + " in combat.")
		whatMob.AddThreatDamage(0, s.actor)
	}

	s.actor.Stam.Subtract(cost)
	s.msg.Actor.SendInfo("You throw everything into the swing!")

	// The flag is only up for this one swing; the damage roll and the miss
	// roll both read it.
	s.actor.FlagOn("reckless", "reckless")
	defer s.actor.FlagOff("reckless", "reckless")
	performAttack(s, whatMob)
}
