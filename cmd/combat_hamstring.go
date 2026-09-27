package cmd

import (
	"strconv"

	"github.com/ArcCS/Nevermore/config"
	"github.com/ArcCS/Nevermore/data"
	"github.com/ArcCS/Nevermore/objects"
	"github.com/ArcCS/Nevermore/permissions"
	"github.com/ArcCS/Nevermore/text"
	"github.com/ArcCS/Nevermore/utils"
)

func init() {
	addHandler(hamstring{},
		"Usage:  hamstring target # \n\n A half-damage swing that taunts the target and cripples it, making it less likely to follow you or block your way.",
		permissions.Fighter,
		"hamstring", "ham")
}

type hamstring cmd

func (hamstring) process(s *state) {
	if len(s.input) < 1 {
		s.msg.Actor.SendBad("Hamstring what exactly?")
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

	// Check some timers
	ready, msg := s.actor.TimerReady("combat_hamstring")
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
		// Try to snag a number off the list
		if val, err := strconv.Atoi(s.words[1]); err == nil {
			nameNum = val
		}
	}

	whatMob := s.where.Mobs.Search(name, nameNum, s.actor)
	if whatMob == nil {
		s.msg.Actor.SendInfo("Hamstring what?")
		s.ok = true
		return
	}

	s.actor.Victim = whatMob
	s.actor.RunHook("combat")

	// Shortcut a missing weapon:
	if s.actor.Equipment.Main == (*objects.Item)(nil) {
		s.msg.Actor.SendBad("You have no weapon to attack with.")
		return
	}

	// Hamstring is a cut to the leg, so it needs a melee weapon within reach.
	if s.actor.Equipment.Main.ItemType == 4 {
		s.msg.Actor.SendBad("You cannot hamstring with a ranged weapon.")
		return
	}
	if inRange, rangeMsg := weaponInRange(s, s.actor.Equipment.Main, whatMob); !inRange {
		s.msg.Actor.SendBad(rangeMsg)
		return
	}

	// Check for a miss. A miss costs the combat round but not the cooldown.
	if utils.Roll(100, 1, 0) <= DetermineMissChance(s, whatMob.Level-s.actor.Tier) {
		s.msg.Actor.SendBad("You missed!!")
		s.msg.Observers.SendBad(s.actor.Name + " fails to hamstring " + whatMob.Name)
		data.StoreCombatMetric("hamstring-miss", 0, 0, 0, 0, 0, 0, s.actor.CharId, s.actor.Tier, 1, whatMob.MobId)
		whatMob.AddThreatDamage(1, s.actor)
		whatMob.CurrentTarget = s.actor.Name
		s.actor.SetTimer("combat", config.CombatCooldown)
		return
	}

	actualDamage, _, resisted := whatMob.ReceiveDamage(config.HamstringDamage(s.actor.InflictDamage()))
	data.StoreCombatMetric("hamstring", 0, 0, actualDamage+resisted, resisted, actualDamage, 0, s.actor.CharId, s.actor.Tier, 1, whatMob.MobId)
	// The taunt: a threat bump on top of the damage.
	whatMob.AddThreatDamage(actualDamage+config.ThreatPercent(whatMob.Stam.Max, config.TauntThreatPercent), s.actor)
	whatMob.CurrentTarget = s.actor.Name
	// Mobs carry no timed states, so a crippled mob stays crippled until it dies.
	whatMob.FlagOn("crippled", "hamstring")
	s.actor.AdvanceSkillExp((float64(actualDamage) / float64(whatMob.Stam.Max) * float64(whatMob.Experience)))
	s.msg.Actor.SendInfo("You hamstring the " + whatMob.Name + " for " + strconv.Itoa(actualDamage) + " damage, crippling it!" + text.Reset)
	s.msg.Observers.SendInfo(s.actor.Name + " hamstrings " + whatMob.Name)
	if whatMob.CheckFlag("reflection") {
		reflectDamage := int(float64(actualDamage) * config.ReflectDamageFromMob)
		stamDamage, vitDamage, resisted := s.actor.ReceiveDamage(reflectDamage)
		data.StoreCombatMetric("hamstring_mob_reflect", 0, 0, stamDamage+vitDamage+resisted, resisted, stamDamage+vitDamage, 1, whatMob.MobId, whatMob.Level, 0, s.actor.CharId)
		s.msg.Actor.Send("The " + whatMob.Name + " reflects " + strconv.Itoa(reflectDamage) + " damage back at you!")
		s.actor.DeathCheck(" was killed by reflection!")
	}
	DeathCheck(s, whatMob)
	s.actor.SetTimer("combat_hamstring", config.HamstringTimer)
	s.actor.SetTimer("combat", config.CombatCooldown)
}
