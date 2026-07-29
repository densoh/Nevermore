package cmd

import (
	"strconv"

	"github.com/ArcCS/Nevermore/config"
	"github.com/ArcCS/Nevermore/objects"
	"github.com/ArcCS/Nevermore/permissions"
)

func init() {
	addHandler(swap{},
		"Usage:  swap \n\n Trade the weapon in your hand for the one prepared at your side.  Can only be done out of combat.",
		permissions.Ranger,
		"swap")
}

type swap cmd

func (swap) process(s *state) {
	if s.actor.Tier < config.SpecialAbilityTier {
		s.msg.Actor.SendBad("You must be at least tier " + strconv.Itoa(config.SpecialAbilityTier) + " to use this skill.")
		return
	}

	if s.actor.Stam.Current <= 0 {
		s.msg.Actor.SendBad("You are far too tired to do that.")
		return
	}

	if s.actor.Equipment.Prepared == (*objects.Item)(nil) {
		s.msg.Actor.SendBad("You have nothing prepared to swap to.")
		return
	}

	if s.actor.Equipment.Prepared.MaxUses <= 0 {
		s.msg.Actor.SendBad("The " + s.actor.Equipment.Prepared.DisplayName() + " is broken.")
		return
	}

	for _, mob := range s.where.Mobs.Contents {
		if mob.CheckThreatTable(s.actor.Name) {
			s.msg.Actor.SendBad("You can't do that while in combat!")
			return
		}
	}

	// Check some timers
	ready, msg := s.actor.TimerReady("combat")
	if !ready {
		s.msg.Actor.SendBad(msg)
		return
	}

	s.actor.RunHook("combat")

	drawn, stowed := s.actor.Equipment.SwapPrepared(s.actor.Class)
	if drawn == (*objects.Item)(nil) {
		s.msg.Actor.SendBad("You cannot swap to that.")
		return
	}

	s.msg.Actor.SendGood("You swap to a " + drawn.DisplayName() + ".")
	s.msg.Observers.SendInfo(s.actor.Name + " swaps to a " + drawn.DisplayName() + ".")
	if stowed != (*objects.Item)(nil) {
		s.msg.Actor.SendInfo("You have a " + stowed.DisplayName() + " ready at your side.")
	}

	s.actor.SetTimer("combat", config.CombatCooldown)
	s.ok = true
}
