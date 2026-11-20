package cmd

import (
	"github.com/ArcCS/Nevermore/config"
	"github.com/ArcCS/Nevermore/data"
	"github.com/ArcCS/Nevermore/objects"
	"github.com/ArcCS/Nevermore/utils"
)

// sneakAttackOdds reports whether a thief's regular hit on whatMob can sneak
// attack, and the percent chance it does. It never can below
// config.SneakAttackTier, with a ranged weapon, or when whatMob itself is
// targeting the thief; otherwise each other living mob in the room targeting
// the thief lowers the chance.
func sneakAttackOdds(actor *objects.Character, whatMob *objects.Mob) (bool, int) {
	if actor.Class != config.THIEF || actor.Tier < config.SneakAttackTier {
		return false, 0
	}
	main := actor.Equipment.Main
	if main == nil || main.ItemType == 4 {
		return false, 0
	}
	if whatMob.CurrentTarget == actor.Name {
		return false, 0
	}
	others := 0
	if room, ok := objects.Rooms[actor.ParentId]; ok && room.Mobs != nil {
		for _, m := range room.Mobs.Contents {
			if m != whatMob && m.Stam.Current > 0 && m.CurrentTarget == actor.Name {
				others++
			}
		}
	}
	chance := config.SneakAttackChance(others)
	return chance > 0, chance
}

// sneakAttackLands rolls whether this hit is a sneak attack.
func sneakAttackLands(s *state, whatMob *objects.Mob) bool {
	ok, chance := sneakAttackOdds(s.actor, whatMob)
	return ok && (chance >= 100 || utils.Roll(100, 1, 0) <= chance)
}

// sneakAttackDamage rolls the sneak attack dice for the actor.
func sneakAttackDamage(s *state) int {
	dice := config.SneakAttackDice(s.actor.Tier, config.StealthLevel(s.actor.Skills[11].Value))
	if dice <= 0 {
		return 0
	}
	return utils.Roll(config.SneakAttackSides, dice, 0)
}

// deviousFor picks what a sneak attack hit on whatMob delivers: the queued
// devious strike, or "" for the sneak attack dice. A queued expose weakness
// on a mob that's already exposed waits for the next mob, and this hit
// sneak attacks instead.
func deviousFor(actor *objects.Character, whatMob *objects.Mob) string {
	if actor.DeviousStrike == config.DeviousExpose && whatMob.CheckFlag("exposed") {
		return ""
	}
	return actor.DeviousStrike
}

// deviousCallbacks maps each strike to its hit callback.
var deviousCallbacks = map[string]string{
	config.DeviousPoison: "devious-poison",
	config.DeviousExpose: "devious-expose",
}

// fireDevious delivers a devious strike that rode in on a hit, once the hit's
// damage is in, through its hit callback. It clears the queue only when the
// strike landed: a mob the hit killed takes nothing, and the strike stays
// queued.
func fireDevious(s *state, whatMob *objects.Mob, strike string) {
	callback, ok := deviousCallbacks[strike]
	if !ok || whatMob.Stam.Current <= 0 {
		return
	}
	if strike == config.DeviousExpose && whatMob.CheckFlag("exposed") {
		return
	}
	objects.DeviousHit(s.actor, callback, whatMob)
	switch strike {
	case config.DeviousPoison:
		s.msg.Actor.SendGood("You slip poison into the " + whatMob.Name + "'s wound!")
		s.msg.Observers.SendInfo(s.actor.Name + " poisons the " + whatMob.Name + "!")
	case config.DeviousExpose:
		s.msg.Actor.SendGood("You expose a weakness in the " + whatMob.Name + "'s defenses!")
		s.msg.Observers.SendInfo(s.actor.Name + " exposes a weakness in the " + whatMob.Name + "'s defenses!")
	}
	data.StoreCombatMetric("devious-"+strike, 0, 0, 0, 0, 0, 0, s.actor.CharId, s.actor.Tier, 1, whatMob.MobId)
	s.actor.DeviousStrike = ""
}
