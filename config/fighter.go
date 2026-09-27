package config

import "github.com/ArcCS/Nevermore/utils"

// Fighter tuning. Everything the fighter kit reads lives here so the class
// can be rebalanced from one file, the same way config/barbarian.go does for
// the barbarian.
const (
	// Hamstring is a half-damage swing that taunts the mob (threat equal to
	// its max stamina, on top of the damage dealt) and leaves it crippled. Mobs
	// carry no timed states, so the "crippled" flag lasts until the mob dies. A
	// crippled mob has its chance to follow a character through an exit, and
	// to block one, cut to CrippledChancePercent of normal. Cooldown is
	// HamstringTimer; a miss only costs the combat round, not the cooldown.
	HamstringTimer         = 30 // seconds between hamstrings
	HamstringDamagePercent = 50 // of the weapon roll
	CrippledChancePercent  = 50 // of the mob's normal follow / block chance

	// Circle stuns for CircleStuns seconds for everyone who has it. A fighter
	// adds one second per CircleFighterLevelsPerSecond weapon skill levels on
	// top, so a grandmaster fighter circles for a full combat round.
	CircleFighterLevelsPerSecond = 2

	// Execute: from ExecuteTier, a fighter with at least ExecuteWeaponLevel
	// (Expert) in the wielded weapon lands lethal blows more often against a
	// wounded mob. The lethal chance is unchanged at or above
	// ExecuteStartHealth of the mob's max stamina, climbs linearly as the mob
	// drops, and reaches ExecuteMaxMultiplier times the table chance at
	// ExecuteFullHealth and below.
	ExecuteTier          = 15
	ExecuteWeaponLevel   = 7 // Expert
	ExecuteStartHealth   = 0.5
	ExecuteFullHealth    = 0.1
	ExecuteMaxMultiplier = 2.0

	// Parry: a fighter parries ParryPerWeaponLevel percent per weapon skill
	// level, so 15% at grandmaster. A shield adds tier / ShieldParryTiersPerPoint
	// points on top, half a point per character level. Parries that only
	// landed because of the shield are blocks: they stop the hit but neither
	// stun nor riposte.
	ParryPerWeaponLevel      = 1.5
	ShieldParryTiersPerPoint = 2.0

	// Shield slam is shared with the paladin (see ShieldStun and SlamTimer in
	// combat.go). Damage is str*ShieldSlamStrMult + tier*ShieldSlamTierMult
	// plus half of a 2d(shield armor) roll, so a heavier shield hits harder.
	// The paladin's stun scales on piety; the fighter's scales on strength
	// with the same multiplier.
	ShieldSlamStrMult  = 2
	ShieldSlamTierMult = 1
)

// HamstringDamage is the damage a hamstring does from the given weapon roll.
// It rounds up so a weak roll still lands for at least one point.
func HamstringDamage(roll int) int {
	return (roll*HamstringDamagePercent + 99) / 100
}

// CrippledChance is a mob's follow or block chance once crippling is applied.
func CrippledChance(chance int) int {
	return chance * CrippledChancePercent / 100
}

// CircleStunFor is how long a circle stuns the mob for a character of the
// given class at the given weapon skill level.
func CircleStunFor(class int, weaponLevel int) int {
	stun := CircleStuns
	if class == FIGHTER {
		stun += weaponLevel / CircleFighterLevelsPerSecond
	}
	return stun
}

// ShieldSlamDamage is a shield slam's damage before mob armor: strength
// times ShieldSlamStrMult, tier times ShieldSlamTierMult, plus half the
// shield roll, which the caller makes with RollShieldArmor.
func ShieldSlamDamage(str int, tier int, shieldRoll int) int {
	return str*ShieldSlamStrMult + tier*ShieldSlamTierMult + shieldRoll/2
}

// RollShieldArmor rolls 2d(armor) for a shield's contribution to a slam. A
// shield with no armor value adds nothing.
func RollShieldArmor(armor int) int {
	if armor <= 0 {
		return 0
	}
	return utils.Roll(armor, 2, 0)
}

// ShieldSlamStun is how long a shield slam stuns the mob. Paladins scale on
// piety, fighters on strength.
func ShieldSlamStun(class int, str int, pie int) int {
	stat := pie
	if class == FIGHTER {
		stat = str
	}
	return int(ShieldStun * float64(stat))
}

// ExecuteMultiplier is the factor applied to a fighter's lethal chance against
// a mob at the given fraction of its max stamina. It is 1 for anyone who has
// not unlocked execute, or against a mob still above ExecuteStartHealth.
func ExecuteMultiplier(tier int, weaponLevel int, healthFraction float64) float64 {
	if tier < ExecuteTier || weaponLevel < ExecuteWeaponLevel || healthFraction >= ExecuteStartHealth {
		return 1
	}
	if healthFraction <= ExecuteFullHealth {
		return ExecuteMaxMultiplier
	}
	progress := (ExecuteStartHealth - healthFraction) / (ExecuteStartHealth - ExecuteFullHealth)
	return 1 + (ExecuteMaxMultiplier-1)*progress
}

// ParryChance is a fighter's percent chance to parry at the given weapon
// skill level and character tier, with or without a shield. Both parts can
// be a half point, so the chance is a float.
func ParryChance(weaponLevel int, tier int, shield bool) float64 {
	if weaponLevel <= 0 {
		return 0
	}
	chance := float64(weaponLevel) * ParryPerWeaponLevel
	if shield {
		chance += float64(tier) / ShieldParryTiersPerPoint
	}
	return chance
}

// RollFighterParry rolls one incoming attack against a fighter's parry. It
// reports whether the attack was parried and, if so, whether only the shield
// bonus caught it: a shield block, which can never be turned into a riposte.
// The roll is to a tenth of a percent so half-point bonuses count.
func RollFighterParry(weaponLevel int, tier int, shield bool) (parried bool, shieldBlock bool) {
	if weaponLevel <= 0 {
		return false, false
	}
	roll := float64(utils.Roll(1000, 1, 0)) / 10
	if roll <= ParryChance(weaponLevel, tier, false) {
		return true, false
	}
	if shield && roll <= ParryChance(weaponLevel, tier, true) {
		return true, true
	}
	return false, false
}
