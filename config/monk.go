package config

import "math"

// Monk tuning. Everything the monk kit reads lives here so the class can be
// rebalanced from one file.
//
// Chi is the monk's resource. It reuses the mana meter (persisted as
// manacur/manamax) but follows different rules: it is generated only by
// landing attacks, dodging, and meditating; it never regenerates passively;
// and it bleeds away once the monk has been out of combat for a while.
const (
	// Tier gates.
	MonkLeapTier     = 5
	MonkSweepTier    = 5
	MonkFlurryTier   = 10
	MonkTodTier      = 10
	MonkLongLeapTier = 10
	MonkDodgeTier    = 15
	MonkFeintTier    = 15
	// From this tier a monk's discipline blunts vital and critical strikes.
	MonkIronBodyTier = 15
	// From this tier, afflictions aimed at a meditating monk must beat a save.
	MonkMeditateSaveTier = 10

	// Chi meter: max = tier*MonkChiPerTier + pie*MonkChiPerPie.
	MonkChiPerTier       = 4
	MonkChiPerPie        = 2
	MonkSpellCostDivisor = 2 // monk spells cost mana/2, minimum 1
	ChiPerHit            = 2 // flat chi per landed hit ...
	ChiPerHitPieDiv      = 8 // ... plus pie/8, never less than ChiPerHitPieMin
	ChiPerHitPieMin      = 1
	DodgeChiGain         = 3 // chi per successful passive dodge
	// Character ticks are 8s apart. The grace is one full tick: the first tick
	// after the last hit is a reprieve, the bleed starts on the second.
	ChiDecayGraceSeconds = 8
	ChiDecayPercent      = 10 // of max chi, per tick, once out of combat

	// Meditate: instant restore of (base + pie/div + tier*perTier)% of each pool,
	// then a trance that halts chi decay and boosts chi generation.
	MeditateTime            = 300 // cooldown seconds
	MeditateBasePercent     = 20
	MeditatePieDiv          = 3
	MeditatePercentPerTier  = 1
	MeditateBaseDuration    = 60 // seconds
	MeditateDurationPerTier = 5
	MeditateChiMultiplier   = 1.5
	// Meditate also clears poison, disease, blindness and drunkenness on cast.
	// While the trance holds, a monk of MonkMeditateSaveTier or higher rolls
	// to shrug off a new affliction:
	//   base + (tier - attacker level)*perLevelDiff + (tier - saveTier) + con/conDiv, capped.
	MeditateSaveBase         = 50
	MeditateSavePerLevelDiff = 2
	MeditateSaveConDiv       = 2
	MeditateSaveCap          = 90

	// Flurry: a stance that spends chi every attack round for extra swings.
	// A round costs 5 chi below Expert and 7 from Expert up.
	FlurryChiCostBase   = 5
	FlurryChiCostExpert = 7
	FlurryExpertSkill   = 7
	FlurryMaxDuration   = 3600 // safety expiry for the stance, seconds

	// Leap strike.
	LeapFreeDistance = 2  // squares closable at no chi cost
	LongLeapChiCost  = 10 // chi for a leap beyond LeapFreeDistance
	LeapTimer        = 30 // cooldown seconds; 0 means none beyond the combat round

	// Sweep.
	SweepTimer              = 30
	SweepStuns              = 12 // seconds
	SweepChance             = 40
	SweepChancePerSkill     = 3
	SweepChanceDexDiv       = 2
	SweepChancePerLevelOver = 4
	SweepChanceCap          = 90
	SweepStunEveryStrike    = false // when flurried, roll the stun on every strike instead of just the first

	// Touch of death. The cooldown starts at TodTimerBase and drops
	// TodTimerPerTier seconds for every tier past TodTimerScaleTier, never
	// below TodTimerMin.
	TodTimerBase      = 600
	TodTimerScaleTier = 15
	TodTimerPerTier   = 60
	TodTimerMin       = 300
	// R, the chi commitment the touch is measured against: base + tier*perTier.
	TodReferenceBase    = 20
	TodReferencePerTier = 4
	// Kill chance = scale * (c - h)^2. The scale starts at TodChanceScale and
	// grows TodChanceScalePerTier for every tier past MonkTodTier, so the touch
	// matures with the monk: the cap sits near a quarter health at tier 10 and
	// near 38% at tier 25.
	TodChanceScale        = 1.5
	TodChanceScalePerTier = 0.05
	TodVulnerableChance   = 0.5 // the monk is told a target looks vulnerable once the odds reach this
	TodMaxChance          = 0.85
	// A touch that lands but fails to kill still tears out a share of what's
	// left plus a few unarmed blows, and it ignores the target's armor.
	TodFailMinPercent = 15
	TodFailMaxPercent = 20
	TodFailHits       = 1
	TodHitPieDiv      = 4   // miss chance -= pie/4
	TodMissRefund     = 0.5 // share of the committed chi returned when the touch whiffs

	// Passive dodge and feint. Feint is a full attack round (flurried if the
	// stance is up, at the normal flurry chi cost) with every swing scaled by
	// FeintDamageMultiplier. It costs no chi of its own; an unflurried feint
	// builds chi like any landed hit.
	MonkDodgeBase         = 5
	MonkDodgePerDex       = 0.5
	MonkDodgePerSkill     = 2
	MonkDodgeCap          = 60
	BreathDodgeMultiplier = 0.5 // a dodged breath weapon still lands, at this share of its damage
	FeintDamageMultiplier = 0.5
	FeintThreatFraction   = 0.5 // of the mob's max stamina, added as threat on a landed feint
	FeintDodgeBonus       = 25
	FeintCharges          = 2
	FeintTimer            = 16
	FeintSafetyDuration   = 60 // seconds before unused feint charges lapse

	// Natural armor.
	MonkArmorPerLevel = 15
	ConMonkArmor      = 2

	// Iron body: vital and critical multipliers against a monk drop by a
	// share of the monk's tier. At tier 15 a vital goes 1.8 -> 1.65 and a
	// critical 3.6 -> 3.3; at tier 25, 1.55 and 3.1. Doubles are untouched.
	MonkVitalReductionPerTier    = 0.01
	MonkCriticalReductionPerTier = 0.02

	// Unarmed damage: base + ceil(str/45 * base) + MonkDamageDice d(base/2),
	// where base is the tier's max weapon damage over MonkDamageDivisor.
	// Fitted against what active players actually wield: ~85-90% of a real
	// weapon from tier 8 up, with flurry carrying it to parity. Two dice keep
	// the roll bell-shaped, so the spread feels tighter than its range.
	MonkDamageDivisor = 2
	MonkDamageDice    = 2
)

// MonkVitalReduction is how much a monk of the given tier shaves off an
// incoming vital strike multiplier; zero below MonkIronBodyTier.
func MonkVitalReduction(tier int) float64 {
	if tier < MonkIronBodyTier {
		return 0
	}
	return float64(tier) * MonkVitalReductionPerTier
}

// MonkCriticalReduction is the same for critical strikes.
func MonkCriticalReduction(tier int) float64 {
	if tier < MonkIronBodyTier {
		return 0
	}
	return float64(tier) * MonkCriticalReductionPerTier
}

// MonkNaturalArmor is a monk's armor before spell modifiers. The tier and
// flat con terms form the base, and con then adds the same percentage bonus
// on top that every other class gets on worn armor.
func MonkNaturalArmor(tier int, con int) int {
	base := tier*MonkArmorPerLevel + con*ConMonkArmor
	return base + int(float64(con)*ConArmorMod*float64(base))
}

// MonkUnarmedBase is the fixed portion of a monk's hit at a tier; the rolled
// portion is MonkDamageDice dice of half that.
func MonkUnarmedBase(tier int) int {
	return MaxWeaponDamage[tier] / MonkDamageDivisor
}

// MonkUnarmedRange is the lowest and highest hit a monk can land at a tier
// and strength, before surge or damage modifiers.
func MonkUnarmedRange(tier int, str int) (int, int) {
	base := MonkUnarmedBase(tier)
	fixed := base + int(math.Ceil(float64(str)/45*float64(base)))
	return fixed + MonkDamageDice, fixed + MonkDamageDice*(base/2)
}

// FlurryMinSkill is the unarmed skill level a flurry swings at when the monk's
// own skill is lower: Ace, the first row of the fighter multi-attack table.
const FlurryMinSkill = 5

// FlurryChiCost is the chi a flurried attack round costs at an unarmed skill level.
func FlurryChiCost(skill int) int {
	if skill >= FlurryExpertSkill {
		return FlurryChiCostExpert
	}
	return FlurryChiCostBase
}

// MonkFlurryFor returns the swing multipliers for a flurried round. Flurry
// follows the fighter multi-attack table exactly, except that a monk below
// Ace still swings as an Ace.
func MonkFlurryFor(skill int) []float64 {
	if skill < FlurryMinSkill {
		skill = FlurryMinSkill
	}
	if mults, ok := MultiAttackMultipliers[skill]; ok {
		return mults
	}
	return []float64{1}
}

// MonkMaxChi is the monk's chi pool at a tier and piety.
func MonkMaxChi(tier int, pie int) int {
	return tier*MonkChiPerTier + pie*MonkChiPerPie
}

// ChiPerHitFor is the chi a monk with the given piety earns per landed hit.
func ChiPerHitFor(pie int) int {
	bonus := pie / ChiPerHitPieDiv
	if bonus < ChiPerHitPieMin {
		bonus = ChiPerHitPieMin
	}
	return ChiPerHit + bonus
}

// ChiDecayAmount is how much chi bleeds away per tick out of combat.
func ChiDecayAmount(max int) int {
	amount := max * ChiDecayPercent / 100
	if amount < 1 {
		return 1
	}
	return amount
}

// MeditateRestorePercent is the share of each pool meditate restores on cast.
func MeditateRestorePercent(tier int, pie int) int {
	return MeditateBasePercent + pie/MeditatePieDiv + tier*MeditatePercentPerTier
}

// MeditateSaveChance is the percent chance a meditating monk shrugs off an
// affliction from an attacker of the given level.
func MeditateSaveChance(tier int, attackerLevel int, con int) int {
	chance := MeditateSaveBase + (tier-attackerLevel)*MeditateSavePerLevelDiff + (tier - MonkMeditateSaveTier) + con/MeditateSaveConDiv
	if chance > MeditateSaveCap {
		return MeditateSaveCap
	}
	if chance < 0 {
		return 0
	}
	return chance
}

// MeditateDuration is the trance length in seconds.
func MeditateDuration(tier int) int {
	return MeditateBaseDuration + tier*MeditateDurationPerTier
}

// TodReference is R: the chi commitment a touch of death is measured against.
func TodReference(tier int) int {
	return TodReferenceBase + tier*TodReferencePerTier
}

// TodCooldown is the touch of death cooldown in seconds at a tier.
func TodCooldown(tier int) int {
	over := tier - TodTimerScaleTier
	if over < 0 {
		over = 0
	}
	cooldown := TodTimerBase - TodTimerPerTier*over
	if cooldown < TodTimerMin {
		return TodTimerMin
	}
	return cooldown
}

// TodChanceScaleFor is the kill chance coefficient at a tier.
func TodChanceScaleFor(tier int) float64 {
	over := tier - MonkTodTier
	if over < 0 {
		over = 0
	}
	return TodChanceScale + float64(over)*TodChanceScalePerTier
}

// TodKillChance is the probability (0..TodMaxChance) that a touch kills,
// given c = chi spent / R and h = the mob's remaining HP fraction.
func TodKillChance(tier int, c float64, h float64) float64 {
	gap := c - h
	if gap <= 0 {
		return 0
	}
	chance := gap * gap * TodChanceScaleFor(tier)
	if chance > TodMaxChance {
		return TodMaxChance
	}
	return chance
}

// SweepStunChance is the percent chance a landed sweep stuns.
func SweepStunChance(skill int, dex int, levelOver int) int {
	chance := SweepChance + skill*SweepChancePerSkill + dex/SweepChanceDexDiv - levelOver*SweepChancePerLevelOver
	if chance > SweepChanceCap {
		return SweepChanceCap
	}
	if chance < 0 {
		return 0
	}
	return chance
}

// MonkDodgeChance is the percent chance a monk's passive dodge evades an
// attack that has already passed the normal miss roll.
func MonkDodgeChance(dex int, skill int, feinted bool) int {
	chance := MonkDodgeBase + int(math.Floor(float64(dex)*MonkDodgePerDex)) + skill*MonkDodgePerSkill
	if feinted {
		chance += FeintDodgeBonus
	}
	if chance > MonkDodgeCap {
		return MonkDodgeCap
	}
	return chance
}
