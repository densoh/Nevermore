package cmd

import (
	"strconv"
	"strings"

	"github.com/ArcCS/Nevermore/config"
	"github.com/ArcCS/Nevermore/objects"
	"github.com/ArcCS/Nevermore/permissions"
)

func init() {
	addHandler(meditate{},
		"Usage:  meditate \n\n Center yourself, restoring some of your health, stamina and chi and purging any poison, disease, blindness or drink, then enter a trance during which your chi does not fade and every gain is heightened.  A seasoned monk in a trance may shrug off afflictions that reach them.  Usable in or out of combat.",
		permissions.Monk,
		"meditate")
}

type meditate cmd

func (meditate) process(s *state) {
	ready, msg := s.actor.TimerReady("combat_meditate")
	if !ready {
		s.msg.Actor.SendBad(msg)
		return
	}
	ready, msg = s.actor.TimerReady("combat")
	if !ready {
		s.msg.Actor.SendBad(msg)
		return
	}

	pct := config.MeditateRestorePercent(s.actor.Tier, s.actor.GetStat("pie"))
	vit := s.actor.HealVital(s.actor.Vit.Max * pct / 100)
	stam := s.actor.HealStam(s.actor.Stam.Max * pct / 100)
	chiBefore := s.actor.Mana.Current
	s.actor.RestoreMana(s.actor.Mana.Max * pct / 100)
	chi := s.actor.Mana.Current - chiBefore

	if cleared := s.actor.ClearAfflictions(); len(cleared) > 0 {
		s.msg.Actor.SendGood("Your discipline purges the " + strings.Join(cleared, ", ") + " from your body.")
	}
	objects.Effects["meditate"](s.actor, s.actor, 0)
	s.msg.Actor.SendGood("You slow your breathing and center yourself, restoring " +
		strconv.Itoa(vit) + " health, " + strconv.Itoa(stam) + " stamina and " + strconv.Itoa(chi) + " chi.")
	s.msg.Observers.SendInfo(s.actor.Name + " meditates!")
	s.actor.SetTimer("combat_meditate", config.MeditateTime)
	s.actor.SetTimer("combat", config.CombatCooldown)

	s.ok = true
}
