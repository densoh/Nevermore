package cmd

import (
	"strconv"

	"github.com/ArcCS/Nevermore/config"
	"github.com/ArcCS/Nevermore/data"
	"github.com/ArcCS/Nevermore/permissions"
	"github.com/ArcCS/Nevermore/utils"
)

func init() {
	addHandler(sweep{},
		"Usage:  sweep target # \n\n Strike low to knock a target off its feet, stunning it if the sweep takes.  Throws extra strikes while flurrying, though a flurried sweep builds no chi.",
		permissions.Monk,
		"sweep")
}

type sweep cmd

func (sweep) process(s *state) {
	if len(s.input) < 1 {
		s.msg.Actor.SendBad("Sweep what exactly?")
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

	if s.actor.Tier < config.MonkSweepTier {
		s.msg.Actor.SendBad("You must be at least tier " + strconv.Itoa(config.MonkSweepTier) + " to use this skill.")
		return
	}

	ready, msg := s.actor.TimerReady("combat_sweep")
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
		s.msg.Actor.SendInfo("Sweep what?")
		s.ok = true
		return
	}

	if s.actor.Placement != whatMob.Placement {
		s.msg.Actor.SendBad("You are too far away to sweep them.")
		return
	}

	s.actor.Victim = whatMob
	s.actor.RunHook("combat")

	skillLevel := attackSkillLevel(s)
	attacks, flurried := monkAttackPlan(s, skillLevel)
	result := resolveHits(s, whatMob, attacks, skillLevel, "sweep")
	if result.hits == 0 {
		s.msg.Observers.SendBad(s.actor.Name + " fails to sweep " + whatMob.Name)
		s.actor.SetTimer("combat", config.CombatCooldown)
		return
	}

	if !flurried {
		gainChiFromHits(s, result.hits)
	}

	levelOver := whatMob.Level - s.actor.Tier
	if levelOver < 0 {
		levelOver = 0
	}
	stunChance := config.SweepStunChance(skillLevel, s.actor.GetStat("dex"), levelOver)
	rolls := 1
	if config.SweepStunEveryStrike {
		rolls = result.hits
	}
	stunned := false
	for i := 0; i < rolls && !stunned; i++ {
		if utils.Roll(100, 1, 0) <= stunChance {
			stunned = true
		}
	}
	if stunned {
		whatMob.Stun(config.SweepStuns)
		data.StoreCombatMetric("sweep-stun", 0, 0, 0, 0, 0, 0, s.actor.CharId, s.actor.Tier, 1, whatMob.MobId)
		s.msg.Actor.SendGood("You sweep " + whatMob.Name + " off its feet!")
		s.msg.Observers.SendInfo(s.actor.Name + " sweeps " + whatMob.Name + " off its feet!")
	} else {
		s.msg.Actor.SendInfo(whatMob.Name + " keeps its footing.")
		s.msg.Observers.SendInfo(s.actor.Name + " sweeps at " + whatMob.Name)
	}
	whatMob.CurrentTarget = s.actor.Name

	DeathCheck(s, whatMob)
	s.actor.SetTimer("combat_sweep", config.SweepTimer)
	s.actor.SetTimer("combat", config.CombatCooldown)
}
