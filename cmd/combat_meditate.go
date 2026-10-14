package cmd

import (
	"math"
	"strconv"
	"strings"

	"github.com/ArcCS/Nevermore/config"
	"github.com/ArcCS/Nevermore/objects"
	"github.com/ArcCS/Nevermore/permissions"
)

func init() {
	addHandler(meditate{},
		"Usage:  meditate \n\n Center yourself, restoring some of your health, stamina and chi and purging any poison, disease, blindness or drink, then enter a trance during which your chi slowly regenerates instead of fading and every gain is heightened.  A seasoned monk in a trance may shrug off afflictions that reach them.  Usable in or out of combat.\n\n Your center returns over five minutes.  Meditating sooner restores less and holds the trance for less time, but always purges afflictions; you can meditate again once your trance has ended and at least 30 seconds have passed.",
		permissions.Monk,
		"meditate")
}

type meditate cmd

func (meditate) process(s *state) {
	// Like a restorative spell, meditate waits only on its own charge and on
	// stun; the global and combat timers do not gate it.
	if ready, msg := s.actor.TimerReadyIgnoreGlobal("stun"); !ready {
		s.msg.Actor.SendBad(msg)
		return
	}
	if s.actor.CheckFlag("meditate") {
		s.msg.Actor.SendBad("You are already deep in a trance.")
		return
	}
	// The combat_meditate timer runs the full recharge from the last cast, so
	// what has passed of it is how centered the monk is now.
	elapsed := float64(config.MeditateChargeTime) - s.actor.TimerRemaining("combat_meditate")
	if elapsed < config.MeditateMinCharge {
		wait := int(math.Ceil(config.MeditateMinCharge - elapsed))
		s.msg.Actor.SendBad("You cannot find your center yet. (" + strconv.Itoa(wait) + " seconds)")
		return
	}
	charge := config.MeditateCharge(elapsed)

	pct := float64(config.MeditateRestorePercent(s.actor.Tier, s.actor.GetStat("pie"))) * charge
	vit := s.actor.HealVital(int(float64(s.actor.Vit.Max) * pct / 100))
	stam := s.actor.HealStam(int(float64(s.actor.Stam.Max) * pct / 100))
	chiBefore := s.actor.Mana.Current
	s.actor.RestoreMana(int(float64(s.actor.Mana.Max) * pct / 100))
	chi := s.actor.Mana.Current - chiBefore

	if cleared := s.actor.ClearAfflictions(); len(cleared) > 0 {
		s.msg.Actor.SendGood("Your discipline purges the " + strings.Join(cleared, ", ") + " from your body.")
	}
	objects.Effects["meditate"](s.actor, s.actor, config.MeditateTranceFor(s.actor.Tier, charge))
	centered := ""
	if charge < 1 {
		centered = " (" + strconv.Itoa(int(charge*100)) + "% centered)"
	}
	s.msg.Actor.SendGood("You slow your breathing and center yourself, restoring " +
		strconv.Itoa(vit) + " health, " + strconv.Itoa(stam) + " stamina and " + strconv.Itoa(chi) + " chi." + centered)
	s.msg.Observers.SendInfo(s.actor.Name + " meditates!")
	// Meditate is a free action: it recharges on its own timer but does not
	// consume the combat round.
	s.actor.SetTimer("combat_meditate", config.MeditateChargeTime)

	s.ok = true
}
