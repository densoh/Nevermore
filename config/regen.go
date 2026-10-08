package config

import "math"

// Passive regeneration, applied once per character tick (8s).
//
// Mana regenerates from ManaRegenMod * (RegenStat + tier), where RegenStat is
// the larger of piety and the average of piety and intelligence, so an
// int-heavy caster is not punished for skipping piety. Health regenerates
// from ConHealRegenMod * con. Both are cut to RegenCombatMod while the
// character has attacked or been attacked within CombatRegenWindowSeconds.
// Rooms flagged heal_fast double both. Bless (or, for paladins, the seal of
// faith) adds a fifth to both and softens the combat cut to
// BlessRegenCombatMod; the room and bless multipliers apply before the
// combat cut.
var (
	ManaRegenMod             = .3
	RegenCombatMod           = .35 // share of regen kept while in combat
	BlessRegenCombatMod      = .65 // share kept while in combat and blessed
	CombatRegenWindowSeconds = 8
	HealFastRoomRegenMod     = 2.0
	BlessRegenMod            = 1.2
)

// RegenStat is the stat mana regen scales from: piety, or the average of
// piety and intelligence when that is higher.
func RegenStat(pie int, intel int) float64 {
	return math.Max(float64(pie), float64(pie+intel)/2)
}

func regenAmount(base float64, roomMod float64, inCombat bool, blessed bool) int {
	amount := base * roomMod
	if blessed {
		amount *= BlessRegenMod
	}
	if inCombat {
		if blessed {
			amount *= BlessRegenCombatMod
		} else {
			amount *= RegenCombatMod
		}
	}
	return int(math.Ceil(amount))
}

// ManaRegen is the mana restored per tick.
func ManaRegen(pie int, intel int, tier int, roomMod float64, inCombat bool, blessed bool) int {
	return regenAmount((RegenStat(pie, intel)+float64(tier))*ManaRegenMod, roomMod, inCombat, blessed)
}

// HealthRegen is the stamina and vitality restored per tick.
func HealthRegen(con int, roomMod float64, inCombat bool, blessed bool) int {
	return regenAmount(float64(con)*ConHealRegenMod, roomMod, inCombat, blessed)
}
