package cmd

import (
	"github.com/ArcCS/Nevermore/config"
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

// rollSneakAttack returns the sneak attack damage for this hit, or 0 when it
// doesn't apply or misses its chance.
func rollSneakAttack(s *state, whatMob *objects.Mob) int {
	ok, chance := sneakAttackOdds(s.actor, whatMob)
	if !ok || (chance < 100 && utils.Roll(100, 1, 0) > chance) {
		return 0
	}
	dice := config.SneakAttackDice(s.actor.Tier, config.StealthLevel(s.actor.Skills[11].Value))
	if dice <= 0 {
		return 0
	}
	return utils.Roll(config.SneakAttackSides, dice, 0)
}
