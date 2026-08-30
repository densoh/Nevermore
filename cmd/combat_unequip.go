package cmd

import (
	"strings"

	"github.com/ArcCS/Nevermore/config"
	"github.com/ArcCS/Nevermore/objects"
	"github.com/ArcCS/Nevermore/permissions"
)

func init() {
	addHandler(unequip{},
		"Usage:  unequip [<item>|all]  # \n\n Try to unequip something you're wearing, if you use 'all' it will remove everything as long as you are not singing.",
		permissions.Player,
		"unequip", "remove", "rem")
}

type unequip cmd

func (unequip) process(s *state) {
	if len(s.words) == 0 {
		s.msg.Actor.SendBad("What did you want to equip?")
		return
	}

	if s.actor.Stam.Current <= 0 {
		s.msg.Actor.SendBad("You are far too tired to do that.")
		return
	}

	// Check some timers
	ready, msg := s.actor.TimerReady("combat")
	if !ready {
		s.msg.Actor.SendBad(msg)
		return
	}

	name := s.input[0]
	s.actor.RunHook("combat")

	if s.actor.CheckFlag("singing") {
		if s.actor.Equipment.FindLocation(name) != "main" || name == "all" {
			s.msg.Actor.SendBad("You may only remove your main hand weapon while performing.")
			s.ok = true
			return
		}
	}

	if name == "all" {
		for _, item := range s.actor.Equipment.UnequipAll() {
			s.actor.Inventory.Add(item)
			s.msg.Actor.SendGood("You unequip " + item.Name)
			s.msg.Observer.SendInfo(s.actor.Name + " unequips " + item.Name)
		}
		s.actor.SetTimer("combat", config.UnequipCooldown)
		s.ok = true
		return
	}

	_, what := s.actor.Equipment.Unequip(name)
	if what != nil {
		s.actor.Inventory.Add(what)
		s.msg.Actor.SendGood("You unequip " + what.Name)
		s.msg.Observer.SendInfo(s.actor.Name + " unequips " + what.Name)
		s.actor.SetTimer("combat", config.UnequipCooldown)
		s.ok = true
		return
	}

	// Nothing worn or wielded matched, but a weapon prepared for a quickdraw can be
	// removed too, which stows it back in the pack the same as an empty prepare.
	if prepared := s.actor.Equipment.Prepared; prepared != (*objects.Item)(nil) &&
		strings.Contains(strings.ToLower(prepared.Name), strings.ToLower(name)) {
		stowed := s.actor.Equipment.Unprepare()
		s.actor.Inventory.Add(stowed)
		s.msg.Actor.SendGood("You put a " + stowed.DisplayName() + " back in your pack.")
		s.msg.Observers.SendInfo(s.actor.Name + " puts a " + stowed.DisplayName() + " away.")
		s.actor.SetTimer("combat", config.UnequipCooldown)
		s.ok = true
		return
	}

	s.msg.Actor.SendInfo("What did you want to unequip?")
	s.ok = true
}
