package cmd

import (
	"math"
	"strconv"

	"github.com/ArcCS/Nevermore/config"
	"github.com/ArcCS/Nevermore/data"
	"github.com/ArcCS/Nevermore/objects"
	"github.com/ArcCS/Nevermore/permissions"
	"github.com/ArcCS/Nevermore/text"
	"github.com/ArcCS/Nevermore/utils"
)

func init() {
	addHandler(tod{},
		"Usage:  tod target # \n\n Commit your chi to a touch of death.  The more chi you bring and the weaker the target, the likelier it dies outright; a touch that fails to kill still wounds, but one that misses entirely leaves you open to a savage blow.",
		permissions.Monk,
		"tod", "touch-of-death")
}

type tod cmd

func (tod) process(s *state) {
	if s.actor.CheckFlag("blind") {
		s.msg.Actor.SendBad("You can't see anything!")
		return
	}

	if len(s.input) < 1 {
		s.msg.Actor.SendBad("Touch of Death what exactly?")
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
		s.msg.Actor.SendInfo("Touch what?")
		s.ok = true
		return
	}

	reference := config.TodReference(s.actor.Tier)

	if s.actor.Stam.Current <= 0 {
		s.msg.Actor.SendBad("You are far too tired to do that.")
		return
	}

	if s.actor.Tier < config.MonkTodTier {
		s.msg.Actor.SendBad("You must be at least tier " + strconv.Itoa(config.MonkTodTier) + " to use this skill.")
		return
	}

	ready, msg := s.actor.TimerReady("combat_tod")
	if !ready {
		s.msg.Actor.SendBad(msg)
		return
	}
	ready, msg = s.actor.TimerReady("combat")
	if !ready {
		s.msg.Actor.SendBad(msg)
		return
	}

	if whatMob.CheckFlag("undead") {
		s.msg.Actor.SendBad("Your target is undead and unaffected by your chi!")
		return
	}
	if whatMob.CheckFlag("no_touch") {
		s.msg.Actor.SendBad("Your chi finds no purchase on " + whatMob.Name + ".")
		return
	}

	if s.actor.Placement != whatMob.Placement {
		s.msg.Actor.SendBad("You are too far away to perform a touch of death on them.")
		return
	}

	if s.actor.Mana.Current <= 0 {
		s.msg.Actor.SendBad("You have no chi to focus.")
		return
	}

	s.actor.Victim = whatMob
	s.actor.RunHook("combat")

	spent := s.actor.Mana.Current
	if spent > reference {
		spent = reference
	}
	s.actor.Mana.Subtract(spent)
	commitment := float64(spent) / float64(reference)
	remaining := hpFraction(whatMob)

	// The touch has to land before it can do anything.
	missChance := DetermineMissChance(s, whatMob.Level-s.actor.Tier) - s.actor.GetStat("pie")/config.TodHitPieDiv
	if missChance < 5 {
		missChance = 5
	}
	if utils.Roll(100, 1, 0) <= missChance {
		s.msg.Actor.SendBad("Your touch finds nothing vital, leaving you open to a savage blow from " + whatMob.Name + "!")
		s.msg.Observers.SendInfo(s.actor.Name + " reaches for " + whatMob.Name + " and misses, leaving an opening.")
		// Nothing connected, so part of the chi flows back.
		if refund := int(float64(spent) * config.TodMissRefund); refund > 0 {
			s.actor.Mana.Add(refund)
		}
		whatMob.AddThreatDamage(1, s.actor)
		data.StoreCombatMetric("tod-miss", 0, 0, 0, 0, 0, 0, s.actor.CharId, s.actor.Tier, 1, whatMob.MobId)
		s.actor.SetTimer("combat", config.CombatCooldown)
		// A botched touch hands the mob a free swing at twice its damage,
		// resolved as a normal-style strike so the player sees the
		// vulnerability text rather than a double banner.
		whatMob.ApplyStrike(s.actor, whatMob.InflictDamage(), objects.StyleNormal, float64(config.CombatModifiers["double"]), objects.StrikeOpts{
			Metric:    "tod_fail_retaliate",
			Mode:      0,
			HitPrefix: "Exposed!! ",
			DeathMsg:  "was slain while attempting a touch of death on a " + utils.Title(whatMob.Name),
		})
		return
	}

	s.actor.SetTimer("combat_tod", config.TodCooldown(s.actor.Tier))
	s.actor.SetTimer("combat", config.CombatCooldown)

	killChance := config.TodKillChance(s.actor.Tier, commitment, remaining)
	if utils.Roll(100, 1, 0) <= int(math.Round(killChance*100)) {
		whatMob.AddThreatDamage(whatMob.Stam.Current, s.actor)
		s.actor.AdvanceSkillExp(float64(whatMob.Experience))
		whatMob.Stam.Current = 0
		data.StoreCombatMetric("tod_whole", 0, 0, whatMob.Stam.Max, 0, whatMob.Stam.Max, 0, s.actor.CharId, s.actor.Tier, 1, whatMob.MobId)
		todKillMessage(s, whatMob)
		DeathCheck(s, whatMob)
		return
	}

	// The kill didn't take: the touch still tears at what's left. Chi goes
	// straight through armor.
	percent := utils.Roll(config.TodFailMaxPercent-config.TodFailMinPercent+1, 1, config.TodFailMinPercent-1)
	damage := whatMob.Stam.Current * percent / 100
	for i := 0; i < config.TodFailHits; i++ {
		damage += s.actor.InflictDamage()
	}
	actualDamage, _ := whatMob.ReceiveDamageNoArmor(damage)
	resisted := 0
	whatMob.AddThreatDamage(actualDamage, s.actor)
	s.actor.AdvanceSkillExp((float64(actualDamage) / float64(whatMob.Stam.Max) * float64(whatMob.Experience)))
	if whatMob.CheckFlag("reflection") {
		reflectDamage := int(float64(actualDamage) * config.ReflectDamageFromMob)
		stamDamage, vitDamage, reflResisted := s.actor.ReceiveDamage(reflectDamage)
		data.StoreCombatMetric("tod_mob_reflect", 0, 0, stamDamage+vitDamage+reflResisted, reflResisted, stamDamage+vitDamage, 1, whatMob.MobId, whatMob.Level, 0, s.actor.CharId)
		s.msg.Actor.Send("The " + whatMob.Name + " reflects " + strconv.Itoa(reflectDamage) + " damage back at you!")
		s.actor.DeathCheck(" was killed by reflection!")
	}
	if whatMob.Stam.Current <= 0 {
		// Lethal all the same; count it as a full touch.
		data.StoreCombatMetric("tod_whole", 0, 0, actualDamage+resisted, resisted, actualDamage, 0, s.actor.CharId, s.actor.Tier, 1, whatMob.MobId)
		todKillMessage(s, whatMob)
	} else {
		data.StoreCombatMetric("tod_partial", 0, 0, actualDamage+resisted, resisted, actualDamage, 0, s.actor.CharId, s.actor.Tier, 1, whatMob.MobId)
		s.msg.Actor.SendInfo("Your touch lands imperfectly, wounding " + whatMob.Name + " for " + strconv.Itoa(actualDamage) + " damage.")
		s.msg.Observers.SendInfo(s.actor.Name + " touches " + whatMob.Name + ", and it staggers.")
	}
	DeathCheck(s, whatMob)
}

// warnTouchVulnerable tells a monk, once per target, when the mob they are
// fighting has been worn down far enough that a touch of death is a good bet
// with the chi they hold right now. Only fires when the touch is available.
func warnTouchVulnerable(s *state, whatMob *objects.Mob) {
	if s.actor.Class != config.MONK || s.actor.Tier < config.MonkTodTier || whatMob.Stam.Current <= 0 {
		return
	}
	if s.actor.TodWarnedTarget == whatMob {
		return
	}
	if ready, _ := s.actor.TimerReady("combat_tod"); !ready {
		return
	}
	reference := config.TodReference(s.actor.Tier)
	spend := s.actor.Mana.Current
	if spend > reference {
		spend = reference
	}
	if config.TodKillChance(s.actor.Tier, float64(spend)/float64(reference), hpFraction(whatMob)) < config.TodVulnerableChance {
		return
	}
	s.actor.TodWarnedTarget = whatMob
	s.msg.Actor.SendInfo(text.Cyan + whatMob.Name + " is faltering; you sense an opening for a finishing blow." + text.Reset)
}

// hpFraction is the mob's remaining hit points as a share of its maximum.
func hpFraction(m *objects.Mob) float64 {
	if m.Stam.Max <= 0 {
		return 0
	}
	return float64(m.Stam.Current) / float64(m.Stam.Max)
}

func todKillMessage(s *state, whatMob *objects.Mob) {
	s.msg.Actor.SendInfo("Your chi flows through you and you perform a perfect touch of death on " + whatMob.Name + ", killing them.")
	s.msg.Observers.SendInfo(s.actor.Name + " touches " + whatMob.Name + " and kills them.")
}
