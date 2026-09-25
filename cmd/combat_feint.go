package cmd

import (
	"strconv"

	"github.com/ArcCS/Nevermore/config"
	"github.com/ArcCS/Nevermore/objects"
	"github.com/ArcCS/Nevermore/permissions"
	"github.com/ArcCS/Nevermore/text"
)

func init() {
	addHandler(feint{},
		"Usage:  feint target # \n\n Strike at a target with a light, baiting blow that draws its attention onto you if it lands, and sets you to slip its next few attacks.  Spends your attack.  Throws extra strikes while flurrying, though a flurried feint builds no chi.",
		permissions.Monk,
		"feint")
}

type feint cmd

func (feint) process(s *state) {
	if len(s.input) < 1 {
		s.msg.Actor.SendBad("Feint at what exactly?")
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

	if s.actor.Tier < config.MonkFeintTier {
		s.msg.Actor.SendBad("You must be at least tier " + strconv.Itoa(config.MonkFeintTier) + " to use this skill.")
		return
	}

	ready, msg := s.actor.TimerReady("combat_feint")
	if !ready {
		s.msg.Actor.SendBad(msg)
		return
	}
	ready, msg = s.actor.TimerReady("combat")
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
		s.msg.Actor.SendInfo("Feint at what?")
		s.ok = true
		return
	}

	if s.actor.Placement != whatMob.Placement {
		s.msg.Actor.SendBad("You are too far away to feint at them.")
		return
	}

	s.actor.Victim = whatMob
	s.actor.RunHook("combat")

	if _, engaged := whatMob.ThreatTable[s.actor.Name]; !engaged {
		s.msg.Actor.Send(text.White + "You engaged " + whatMob.Name + " #" + strconv.Itoa(s.where.Mobs.GetNumber(whatMob)) + " in combat.")
		whatMob.AddThreatDamage(0, s.actor)
	}

	// The stance is the monk's own movement, so it takes whether or not the
	// blow lands.  The round is spent either way.
	objects.Effects["feint"](s.actor, s.actor, config.FeintCharges)
	s.actor.SetTimer("combat_feint", config.FeintTimer)
	s.actor.SetTimer("combat", config.CombatCooldown)

	skillLevel := attackSkillLevel(s)
	plan, flurried := monkAttackPlan(s, skillLevel)
	// Copy before scaling: a flurried plan is the shared multiplier table.
	attacks := make([]float64, len(plan))
	for i, mult := range plan {
		attacks[i] = mult * config.FeintDamageMultiplier
	}
	result := resolveHits(s, whatMob, attacks, skillLevel, "feint")
	if result.hits == 0 {
		s.msg.Observers.SendBad(s.actor.Name + " feints at " + whatMob.Name + ", but it isn't fooled.")
		return
	}

	if !flurried {
		gainChiFromHits(s, result.hits)
	}

	whatMob.AddThreatDamage(int(float64(whatMob.Stam.Max)*config.FeintThreatFraction), s.actor)
	whatMob.CurrentTarget = s.actor.Name
	s.msg.Actor.SendInfo(whatMob.Name + " turns its attention to you.")
	s.msg.Observers.SendInfo(s.actor.Name + " feints at " + whatMob.Name + ", drawing its attention.")
	DeathCheck(s, whatMob)
}
