package cmd

import (
	"math"
	"strconv"

	"github.com/ArcCS/Nevermore/config"
	"github.com/ArcCS/Nevermore/objects"
	"github.com/ArcCS/Nevermore/permissions"
	"github.com/ArcCS/Nevermore/text"
)

func init() {
	addHandler(berserk{},
		"Usage:  berserk \n\n Begin an uncontrollable rage with enhanced strength. From tier "+strconv.Itoa(config.BloodlustTier)+", every kill while berserk extends the rage and restores stamina.",
		permissions.Barbarian,
		"berserk", "rage", "berz")
}

type berserk cmd

func (berserk) process(s *state) {
	// Check some timers
	if !s.requireStamina() {
		return
	}

	if !s.requireTier(config.MajorAbilityTier) {
		return
	}

	if s.actor.CheckFlag("berserk") {
		s.msg.Actor.SendBad("You're already in the grips of the red rage!")
		return
	}

	ready, msg := s.actor.TimerReady("combat_berserk")
	if !ready {
		s.msg.Actor.SendBad(msg)
		return
	}
	ready, msg = s.actor.TimerReady("combat")
	if !ready {
		s.msg.Actor.SendBad(msg)
		return
	}

	s.actor.RunHook("combat")
	objects.Effects["berserk"](s.actor, s.actor, 0)
	s.msg.Observers.SendInfo(s.actor.Name + " goes berserk!")
	s.actor.SetTimer("combat_berserk", config.BerserkCooldown)

	s.ok = true
}

// bloodlust rewards a berserk barbarian for a kill: the rage runs longer, up
// to its cap, and a share of max stamina comes back. Only the killer benefits.
func bloodlust(s *state, slain string) {
	if s.actor.Class != config.BARBARIAN || s.actor.Tier < config.BloodlustTier || !s.actor.CheckFlag("berserk") {
		return
	}
	s.actor.ExtendEffect("berserk", config.BloodlustExtendSeconds, config.BerserkMaxDurationFor(s.actor.Tier))
	healed := s.actor.HealStam(int(math.Ceil(float64(s.actor.Stam.Max) * float64(config.BloodlustStamPercent) / 100)))
	msg := text.Red + "Your rage is fueled by " + slain + "'s death!"
	if healed > 0 {
		msg += " You recover " + strconv.Itoa(healed) + " stamina."
	}
	s.msg.Actor.Send(msg + text.Reset)
}
