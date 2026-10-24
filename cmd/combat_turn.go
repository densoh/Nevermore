package cmd

import (
	"github.com/ArcCS/Nevermore/config"
	"github.com/ArcCS/Nevermore/data"
	"github.com/ArcCS/Nevermore/objects"
	"github.com/ArcCS/Nevermore/permissions"
	"github.com/ArcCS/Nevermore/utils"
	"strconv"
)

func init() {
	addHandler(turn{},
		"Usage:  turn target # \n\n Channel the power of your faith into an undead target, instilling "+
			"fear and potentially destroying them completely.  Should your faith falter, the creature will charge you and you will be open to a savage blow.",
		permissions.Cleric|permissions.Paladin,
		"turn")
}

type turn cmd

func (turn) process(s *state) {
	if len(s.input) < 1 {
		s.msg.Actor.SendBad("Turn what exactly?")
		return
	}
	if !s.requireSight() {
		return
	}

	if !s.requireStamina() {
		return
	}
	if !s.requireTier(config.SpecialAbilityTier) {
		return
	}

	// Check some timers
	ready, msg := s.actor.TimerReady("combat_turn")
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

	var whatMob *objects.Mob
	whatMob = s.where.Mobs.Search(name, nameNum, s.actor)
	if whatMob != nil {
		if whatMob.Flags["undead"] != true {
			s.msg.Actor.SendBad("Your target isn't undead!")
			return
		}

		if s.actor.Class == 6 && (s.actor.Placement != whatMob.Placement) {
			s.msg.Actor.SendBad("You are too far away to turn them.")
			return
		}
		s.actor.Victim = whatMob

		s.actor.RunHook("combat")
		s.actor.SetTimer("combat_turn", config.TurnTimer)
		s.actor.SetTimer("combat", config.CombatCooldown)
		divinityRank := config.WeaponLevel(s.actor.Skills[config.DivinitySkill].Value, s.actor.Class, config.DivinitySkill)
		curChance := config.TurnSuccessChance(s.actor.Tier, whatMob.Level, s.actor.GetStat("pie"), divinityRank)

		// One d1000: the bottom of the success band disintegrates, the rest of it turns.
		turnRoll := utils.Roll(1000, 1, 0)
		turned := curChance >= 100 || turnRoll <= curChance*10
		if turnRoll <= config.TurnDisintegratePerMille(curChance) {
			s.msg.Actor.SendInfo("Your faith overwhelms the " + whatMob.Name + " and utterly demolishes them.")
			s.msg.Observers.SendInfo(s.actor.Name + " disintegrates " + whatMob.Name)
			whatMob.AddThreatDamage(whatMob.Stam.Current, s.actor)
			s.actor.AdvanceDivinity(config.TurnDivinityExp(whatMob.Stam.Current, whatMob.Stam.Max, whatMob.Experience()), s.actor.Class)
			whatMob.Stam.Current = 0
			data.StoreCombatMetric("turn_whole", 0, 0, whatMob.Stam.Max, 0, whatMob.Stam.Max, 0, s.actor.CharId, s.actor.Tier, 1, whatMob.MobId)
			DeathCheck(s, whatMob)
			whatMob = nil
		} else if turned {
			if s.actor.Class == config.PALADIN {
				// Paladins turn from the front line, so the undead turns on them.
				s.msg.Actor.SendInfo("You stand fast before " + whatMob.Name + " and drive your faith into it, causing " + strconv.Itoa(whatMob.Stam.Current/2) + " damage!  It turns its fury on you.")
				whatMob.AddThreatDamage(whatMob.Stam.Current/2+config.ThreatPercent(whatMob.Stam.Max, config.LargeThreatPercent), s.actor)
				whatMob.CurrentTarget = s.actor.Name
			} else {
				s.msg.Actor.SendInfo("Your faith pours into " + whatMob.Name + " and causes " + strconv.Itoa(whatMob.Stam.Current/2) + " damage.")
				whatMob.AddThreatDamage(whatMob.Stam.Current/2, s.actor)
			}
			s.msg.Observers.SendInfo(s.actor.Name + " turned " + whatMob.Name)
			s.actor.AdvanceDivinity(config.TurnDivinityExp(whatMob.Stam.Current/2, whatMob.Stam.Max, whatMob.Experience()), s.actor.Class)
			whatMob.Stam.Subtract(whatMob.Stam.Current / 2)
			data.StoreCombatMetric("turn_half", 0, 0, whatMob.Stam.Current, 0, whatMob.Stam.Current, 0, s.actor.CharId, s.actor.Tier, 1, whatMob.MobId)
		} else {
			s.msg.Actor.SendBad("You fail to turn the " + whatMob.Name + ".  They charge you, and you are left open to a savage blow!")
			whatMob.CurrentTarget = s.actor.Name
			whatMob.Placement = s.actor.Placement
			whatMob.AddThreatDamage(config.FailedTurnThreat(whatMob.Stam.Current, whatMob.Stam.Max), s.actor)
			s.msg.Observers.SendInfo(s.actor.Name + " turn attempt fails and enrages " + whatMob.Name)
			// The enraged mob may get a free vital strike, scaled to the mob
			// rather than to the player.
			whatMob.RevengeVital(s.actor, config.MobTurnRevengeVitalChance, nil, objects.StrikeOpts{
				Metric:    "turn_fail_retaliate",
				Mode:      0,
				HitPrefix: "Exposed!! ",
				DeathMsg:  "was slain while trying to turn a " + utils.Title(whatMob.Name),
			})
		}
		return
	}

	s.msg.Actor.SendInfo("Attack what?")
	s.ok = true
}
