package cmd

import (
	"strconv"

	"github.com/ArcCS/Nevermore/config"
	"github.com/ArcCS/Nevermore/objects"
	"github.com/ArcCS/Nevermore/permissions"
	"github.com/ArcCS/Nevermore/text"
)

func init() {
	addHandler(quickdraw{},
		"Usage:  quickdraw target # \n\n Draw the weapon you have prepared at your side and attack with it in one motion.  The weapon it replaces is left ready at your side.",
		permissions.Ranger,
		"quickdraw", "qd")
}

type quickdraw cmd

func (quickdraw) process(s *state) {
	if s.actor.Tier < config.SpecialAbilityTier {
		s.msg.Actor.SendBad("You must be at least tier " + strconv.Itoa(config.SpecialAbilityTier) + " to use this skill.")
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

	drawWeapon := s.actor.Equipment.Prepared
	if drawWeapon == (*objects.Item)(nil) {
		s.msg.Actor.SendBad("You have nothing prepared to draw.")
		return
	}

	if drawWeapon.MaxUses <= 0 {
		s.msg.Actor.SendBad("The " + drawWeapon.DisplayName() + " is broken.")
		return
	}

	if len(s.input) < 1 && s.actor.Victim == nil {
		s.msg.Actor.SendBad("Quickdraw on what exactly?")
		return
	}

	// Check some timers
	ready, msg := s.actor.TimerReady("combat_quickdraw")
	if !ready {
		s.msg.Actor.SendBad(msg)
		return
	}

	ready, msg = s.actor.TimerReady("combat")
	if !ready {
		s.msg.Actor.SendBad(msg)
		return
	}

	name := ""
	nameNum := 1
	if len(s.words) < 1 && s.actor.Victim != nil {
		switch s.actor.Victim.(type) {
		case *objects.Character:
			name = s.actor.Victim.(*objects.Character).Name
		case *objects.Mob:
			name = s.actor.Victim.(*objects.Mob).Name
			nameNum = s.where.Mobs.GetNumber(s.actor.Victim.(*objects.Mob))
		}
	} else {
		name = s.input[0]
	}

	if len(s.words) > 1 {
		// Try to snag a number off the list
		if val, err := strconv.Atoi(s.words[1]); err == nil {
			nameNum = val
		}
	}

	var whatMob *objects.Mob
	whatMob = s.where.Mobs.Search(name, nameNum, s.actor)
	if whatMob == nil {
		s.msg.Actor.SendInfo("Quickdraw on what?")
		s.ok = true
		return
	}

	// A draw the prepared weapon couldn't reach with costs nothing at all, the weapon
	// stays at their side and no timers are spent.
	if inRange, rangeMsg := weaponInRange(s, drawWeapon, whatMob); !inRange {
		s.msg.Actor.SendBad(rangeMsg)
		return
	}

	s.actor.Victim = whatMob
	s.actor.RunHook("combat")

	drawn, stowed := s.actor.Equipment.SwapPrepared(s.actor.Class)
	if drawn == (*objects.Item)(nil) {
		s.msg.Actor.SendBad("You fumble your draw and come up empty handed.")
		return
	}

	s.msg.Actor.SendGood("You draw a " + drawn.DisplayName() + " in one smooth motion!")
	s.msg.Observers.SendInfo(s.actor.Name + " draws a " + drawn.DisplayName() + " in one smooth motion!")
	if stowed != (*objects.Item)(nil) {
		s.msg.Actor.SendInfo("You have a " + stowed.DisplayName() + " ready at your side.")
	}

	if _, engaged := whatMob.ThreatTable[s.actor.Name]; !engaged {
		s.msg.Actor.Send(text.White + "You engaged " + whatMob.Name + " #" + strconv.Itoa(s.where.Mobs.GetNumber(whatMob)) + " in combat.")
		s.msg.Observers.Send(text.White + s.actor.Name + " attacks " + whatMob.Name)
		whatMob.AddThreatDamage(0, s.actor)
	}

	s.actor.SetTimer("combat_quickdraw", config.QuickdrawCooldown)
	performAttack(s, whatMob)
	s.ok = true
}
