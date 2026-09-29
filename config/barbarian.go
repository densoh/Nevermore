package config

import "math"

// Barbarian tuning. Everything the berserk kit reads lives here so the class
// can be rebalanced from one file.
const (
	// Berserk opens at MajorAbilityTier. It is a rage that lasts
	// BerserkDuration + tier*BerserkDurationPerTier seconds and adds a flat
	// bonus to every hit of str*BerserkDamagePerStr + tier*BerserkDamagePerTier,
	// where str already includes BerserkStrBonus. Recomputed once when the
	// rage starts and removed exactly when it ends.
	BerserkStrBonus        = 5
	BerserkDamagePerStr    = 3
	BerserkDamagePerTier   = 2
	BerserkDuration        = 60 // base seconds
	BerserkDurationPerTier = 1
	BerserkCooldown        = 300 // seconds between rages
	// Bloodlust: from BloodlustTier, every mob the barbarian kills while
	// berserk extends the rage and restores stamina. Extensions can add at
	// most BerserkMaxExtension seconds on top of the rage's base length, so
	// trash kills cannot sustain it forever.
	BloodlustTier          = 15
	BloodlustExtendSeconds = 15
	BloodlustStamPercent   = 10 // of max stamina, per kill
	BerserkMaxExtension    = 60

	// Reckless: a toggled melee stance that trades stamina for accuracy and
	// damage. Every attack thrown while it is up costs RecklessStamPercent of
	// max stamina; the stance drops on its own when the barbarian cannot cover
	// the cost or draws a ranged weapon. The damage bonus scales the weapon
	// roll only, before the berserk flat bonus is added, so it is worth the
	// same in and out of rage.
	RecklessTier          = 5
	RecklessStamPercent   = 5    // of max stamina, per attack, minimum 1
	RecklessDamagePercent = 25   // added to the weapon roll
	RecklessMissReduction = 10   // percentage points off the miss chance
	RecklessMaxDuration   = 3600 // safety expiry for the stance, seconds

	// Bash adds BashDamagePerTier per tier on top of its hit, after the
	// special-roll multiplier so a Thunk does not amplify it.
	BashDamagePerTier = 3

	// A Thunk multiplies the bash hit by ThunkBaseMultiplier through
	// ThunkBaseTier, then ThunkMultiplierPerTier more for every tier above it.
	ThunkBaseMultiplier    = 5.0
	ThunkBaseTier          = 5
	ThunkMultiplierPerTier = 0.5
)

// ThunkMultiplier is the damage multiplier on a Thunk bash at the given tier.
func ThunkMultiplier(tier int) float64 {
	if tier <= ThunkBaseTier {
		return ThunkBaseMultiplier
	}
	return ThunkBaseMultiplier + float64(tier-ThunkBaseTier)*ThunkMultiplierPerTier
}

// BashDamage is a bash's damage before mob armor: the weapon hit times the
// special-roll multiplier, plus the flat tier bonus.
func BashDamage(hit int, multiplier float64, tier int) int {
	return int(math.Ceil(float64(hit)*multiplier)) + tier*BashDamagePerTier
}

// RecklessStamCost is the stamina one reckless attack costs at the given max.
func RecklessStamCost(maxStam int) int {
	cost := (maxStam*RecklessStamPercent + 99) / 100
	if cost < 1 {
		cost = 1
	}
	return cost
}

// BerserkDurationFor is how long a rage lasts at the given tier before any
// bloodlust extensions.
func BerserkDurationFor(tier int) int {
	return BerserkDuration + tier*BerserkDurationPerTier
}

// BerserkMaxDurationFor is the longest a rage can run at the given tier once
// bloodlust extensions are counted.
func BerserkMaxDurationFor(tier int) int {
	return BerserkDurationFor(tier) + BerserkMaxExtension
}

// BerserkDamageBonus is the flat damage a berserk barbarian adds to each hit.
func BerserkDamageBonus(str int, tier int) int {
	return str*BerserkDamagePerStr + tier*BerserkDamagePerTier
}
