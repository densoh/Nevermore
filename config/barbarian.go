package config

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

	// Reckless: a single melee swing that trades stamina for accuracy and
	// damage. The damage bonus scales the weapon roll only, before the
	// berserk flat bonus is added, so it is worth the same in and out of rage.
	RecklessTier          = 5
	RecklessStamPercent   = 5  // of max stamina, per swing, minimum 1
	RecklessDamagePercent = 25 // added to the weapon roll
	RecklessMissReduction = 10 // percentage points off the miss chance
)

// RecklessStamCost is the stamina one reckless swing costs at the given max.
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
