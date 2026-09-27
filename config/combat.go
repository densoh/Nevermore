package config

import "github.com/ArcCS/Nevermore/utils"

var CombatModifiers = map[string]int{
	// Attack Modifiers
	"critical": 5,
	"double":   2,

	// Bash
	"thunk":    100,
	"crushing": 10,
	"thwomp":   2,

	// Sneaky Types
	"backstab": 5,
	"snipe":    4,
}

// MultiAttackMultipliers is the damage multiplier for each landed hit of a
// fighter or ranger multi attack, keyed by weapon skill level.  Multipliers
// are applied to landed hits in order, so the first hit that connects always
// takes the 1.0 slot.  Levels not present get a single swing.
var MultiAttackMultipliers = map[int][]float64{
	5:  {1, 0.2},
	6:  {1, 0.4},
	7:  {1, 0.4, 0.1},
	8:  {1, 0.5, 0.2},
	9:  {1, 0.5, 0.3, 0.2},
	10: {1, 0.5, 0.3, 0.2, 0.2},
}

// MultiAttackMissPenaltyFor returns the extra miss chance, in percentage
// points, applied to every swing after the first in a multi attack.  It starts
// at MultiAttackMissPenalty and drops by MultiAttackMissReduction for each
// weapon skill level above MultiAttackMissReductionLevel.
func MultiAttackMissPenaltyFor(skillLevel int) int {
	penalty := MultiAttackMissPenalty - MultiAttackMissReduction*(skillLevel-MultiAttackMissReductionLevel)
	if penalty > MultiAttackMissPenalty {
		penalty = MultiAttackMissPenalty
	}
	if penalty < 0 {
		penalty = 0
	}
	return penalty
}

// MajorAbilityTier gates the tier 10 class abilities (berserk, flurry, touch
// of death, the long leap, the meditate save). A const so other config
// constants can build on it.
const MajorAbilityTier = 10

var (
	MultiAttackMissPenalty        = 25 // Percentage points added to follow-up swing miss chance
	MultiAttackMissReduction      = 5  // Points removed from the penalty per skill level past the threshold
	MultiAttackMissReductionLevel = 7  // Skill level at which the penalty starts falling off: 30% follow-up miss through Expert, then 25/20/15 at Specialist/Master/Grandmaster
	BaselineStatValue             = 10
	ProximityChance               = 80
	ProximityStep                 = 10

	CombatCooldown  = 8
	UnequipCooldown = 2

	// ResumeGraceSeconds is how long after resuming a session a death still
	// counts as a lag death, so a fight that ran while the player was
	// link-dead cannot kill them for real before they have seen the screen.
	ResumeGraceSeconds = 3

	RoomClearTimer            = 3  // Seconds
	RoomEffectInvocation      = 18 // Seconds
	RoomDefaultEncounterSpeed = 10 // Seconds
	RoomMaxJigger             = 4
	// Crowding: once a room holds more hostile mobs than characters, double
	// spawns stop and the spawn chance falls linearly, reaching half at
	// chars + RoomEncCrowdHalfAt mobs and staying there.
	RoomEncCrowdHalfAt = 4

	BaseDevicePiety = 8.0

	IntMajorPenalty = 9
	PieMajorPenalty = 5

	MobAugmentPerCharacter = 3

	FreeDeathTier = 4

	SpecialAbilityTier = 7
	MinorAbilityTier   = 5
	SealJusticeTier    = 10 // Courage and faith open at MinorAbilityTier, justice here

	MobVital       = 3
	MobCritical    = 4
	MobDouble      = 10
	MobFollowVital = 40
	MobFollMult    = 3

	BindCost   = 75000
	RenameCost = 150000

	MissPerLevel = 8 // This is a percentage
	SearchPerInt = 3 // This is a percentage

	SurgeExtraDamage     = .15
	SurgeDamageBonus     = .20 // Percentage added when using surge
	InertialDamageIgnore = .20 // Percentage ignored when using inertial barrier
	ReflectDamagePerInt  = .02 // Percentage of damage reflected per int point
	ReflectDamageFromMob = .15 // Percentage of damage reflected from mob
	LethalSkillFloor     = .25 // Minimum fraction of a mob's exp granted as weapon skill on a lethal

	DodgeDamagePerDex     = .01
	FullDodgeChancePerDex = 1.0

	PeekCD                      = 8
	StealCD                     = 8
	HideChance                  = 20
	SneakChance                 = 20
	SneakBonus                  = 10
	StealChance                 = 20
	StealChancePerSkillLevel    = 4
	BackStabChance              = 20 // starting hit chance before stealth, dex and level
	BackstabDamageSkillModifier = .15
	BackStabChancePerSkillLevel = 3
	BackstabMissPerLevel        = 8 // points of miss per level the mob is above the thief, from level 2
	SnipeChance                 = 15
	HideChancePerPoint          = 3
	SneakChancePerPoint         = 1
	SneakChancePerTier          = 1
	StealChancePerPoint         = 1
	SnipeChancePerPoint         = 1
	SnipeChancePerLevel         = 5
	SnipeFumbleChance           = 20
	MobStealRevengeVitalChance  = 15
	MobBSRevengeVitalChance     = 25
	VitalStrikeScale            = 2
	BackstabCooldown            = 30
	QuickdrawCooldown           = 30
	TrackCooldown               = 16
	TrackChance                 = 20
	TrackChancePerLevel         = 5
	TrackChancePerPoint         = 1

	TurnMax            = 50
	TurnScaleDown      = 10
	DisintegrateChance = 5
	TurnTimer          = 60
	SlamTimer          = 30
	ShieldStun         = .4

	// Paladin seals persist until dropped, so the effect gets a duration it
	// will never reach.
	SealDuration             = 10 * 365 * 24 * 60 * 60 // Seconds
	SealTimer                = 10
	SealCourageArmorBase     = 5
	SealCourageArmorPerLevel = 1
	// RescueWindow is how long attacks aimed at a rescued ally are drawn onto
	// the paladin; RescueTimer is the cooldown between rescues.
	RescueWindow = 8  // Seconds
	RescueTimer  = 30 // Seconds

	ScalePerPiety  = 1
	DurationPerCon = 10

	ParryStuns  = 2
	CircleStuns = 1
	CircleTimer = 16
	BashStuns   = 16
	BashTimer   = 45

	// Threat bumps, as a percent of the mob's max stamina, added on top of any
	// damage the ability dealt. Taunts (circle, feint, hamstring, shield slam)
	// use TauntThreatPercent; bash is a lighter taunt. A failed backstab hands
	// the thief threat; a failed turn hands over the mob's current stamina,
	// capped at FailedTurnThreatCapPercent.
	TauntThreatPercent          = 50
	BashThreatPercent           = 25
	FailedBackstabThreatPercent = 25
	FailedTurnThreatCapPercent  = 50
	// Heal threat is the amount healed split evenly across the mobs
	// attacking the recipient (plus the healer's own target), but no single
	// mob takes more than HealThreatCapPercent of the heal.
	HealThreatCapPercent = 50

	// MobThreatSwitchChance is the percent chance per mob tick (about 8s)
	// that a mob turns on whoever tops its threat table when that is not its
	// current target. Taunts set the target directly and skip this roll.
	MobThreatSwitchChance = 10

	MobBlock          = 35
	MobBlockPerLevel  = 15
	MobFollow         = 40
	MobFollowPerLevel = 2
	MobTakeChance     = 20 // Percent

	StrCarryMod     = 10 // Per Point
	BaseCarryWeight = 40
	StatDamageMod   = .01 // Per Point

	ReduceSickCon     = 1
	SickConBonus      = 2
	ConBonusHealthDiv = 5
	ConHealRegenMod   = .10
	ConFallDamageMod  = 1
	ConArmorMod       = .005

	HitPerDex        = 1
	MissPerDex       = 1
	BlessHitBonus    = 5 // percentage points off the miss chance while blessed
	DexFallDamageMod = 1

	FallDamage = .20

	IntResistMagicBase     = 10
	IntResistMagicPerPoint = 1
	IntManaPool            = 2  // Number of points of mana to add
	IntManaPoolDiv         = 5  // Number to divide by
	IntSpellEffectDuration = 30 // Seconds to add
	IntBroad               = 1  // Number of broadcasts per int point
	IntEvalDivInt          = 3  //Divide int by this number to get eval
	BaseEvals              = 1
	BaseBroads             = 5
	FizzleSave             = 25 // chance to fizzle per int below 9

	PieHealMod           = .7 // Per point
	MinorPieHealMod      = .6 // Per point, vigor/mend only
	MinorHealDivinityMod = .3 // Vigor/mend get 30% of the divinity bonus of detraumatize/renewal
	PaladinDivinityMod   = .5 // Paladins get this share of their divinity bonus without the seal of faith

	// Seal of justice adds (base + pie/pieDiv + weaponLevel/weaponDiv) percent damage.
	// Seal of justice damage bonus, in percent: base + tier/TierDiv +
	// pie/PieDiv + weaponLevel/WeaponDiv. About 15% at tier 10 and 27% at 20.
	SealJusticeBase      = 5.0
	SealJusticeTierDiv   = 2.0
	SealJusticePieDiv    = 4.0
	SealJusticeWeaponDiv = 2.0
	MinorHealTierDiv     = 4 // Vigor/mend gain +1 base per this many tiers
	MajorHealTierDiv     = 2 // Detraumatize/renewal gain +1 base per this many tiers
	MajorHealBaseCut     = 5 // Subtracted from detraumatize/renewal base to offset the tier bonus

	ArmorReduction         = .007
	ArmorReductionPoints   = 10
	ArmorReductionConstant = 1100

	MobArmorReduction = .5

	ExperienceReduction = map[int]float64{
		1: .9,
		2: .7,
		3: .6,
		4: .5,
		5: .45,
	}
	AttractionChance = 80
)

func MaxWeight(str int) int {
	return BaseCarryWeight + (str * StrCarryMod)
}

func CalcHealth(tier int, con int, class int) int {
	if class >= 99 {
		return 800
	}
	return (tier * Classes[AvailableClasses[class]].Health) + int(float64(tier)*(float64(con)/float64(ConBonusHealthDiv)))
}

func CalcStamina(tier int, con int, class int) int {
	if class >= 99 {
		return 800
	}
	return (tier * Classes[AvailableClasses[class]].Stamina) + int(float64(tier)*(float64(con)/float64(ConBonusHealthDiv)))
}

// CalcMana is the mana pool at a tier. Monks carry chi instead, which scales
// on piety rather than intelligence.
func CalcMana(tier int, intel int, pie int, class int) int {
	if class >= 99 {
		return 800
	}
	if class == MONK {
		return MonkMaxChi(tier, pie)
	}
	return (tier * Classes[AvailableClasses[class]].Mana) + int(float64(tier)*(float64(intel)/float64(IntManaPoolDiv))*float64(IntManaPool))
}

func CalcHaste(tier int) int {
	if tier < 10 {
		return 2
	} else if tier >= 10 && tier < 15 {
		return 3
	} else if tier > 15 {
		return 4
	}
	return 0
}

var DoubleDamage = []int{
	0,
	1,
	2,
	4,
	6,
	8,
	10,
	12,
	15,
	20,
	25,
}

func RollDouble(skill int) bool {
	if skill > 0 {
		dRoll := utils.Roll(100, 1, 0)
		if dRoll <= DoubleDamage[skill] {
			return true
		}
	}
	return false
}

var CriticalDamage = []int{
	0,
	1,
	2,
	3,
	4,
	5,
	6,
	7,
	8,
	10,
	12,
}

func RollCritical(skill int) bool {
	if skill > 0 {
		dRoll := utils.Roll(1000, 1, 0)
		if dRoll <= CriticalDamage[skill] {
			return true
		}
	}
	return false
}

var LethalDamage = []int{ // Lethals are 1000000 chance rolls.
	0,
	125,
	250,
	500,
	750,
	1000,
	1250,
	1500,
	1875,
	2500,
	3125,
}

// BashChances Skill = Thunk, Crushing, Thwomp, Thump
var BashChances = map[int][]int{
	0: {0, 0, 0, 0},
	1: {125 * 4, 600 * 4, 1200 * 4, 2400 * 4},
	2: {250 * 4, 1000 * 4, 2000 * 4, 4000 * 4},
	3: {500 * 4, 2000 * 4, 4000 * 4, 8000 * 4},
	4: {750 * 4, 3000 * 4, 6000 * 4, 12000 * 4},
	5: {1000 * 4, 4000 * 4, 8000 * 4, 16000 * 4},
	6: {1250 * 4, 5000 * 4, 10000 * 4, 20000 * 4},
	7: {1500 * 4, 6000 * 4, 12000 * 4, 24000 * 4},
	8: {1875 * 4, 7500 * 4, 15000 * 4, 30000 * 4},
	9: {3000 * 4, 12000 * 4, 24000 * 4, 48000 * 4},
}

func RollBash(skill int) (damModifier int, stunModifier int, output string) {
	/*
		var ThumpRoll = 10
		var ThwompRoll = 50
		var CrushingRoll = 500
		var ThunkRoll = 1000

	*/
	damModifier = 1
	stunModifier = 1
	bashRoll := utils.Roll(1000000, 1, 0)
	if bashRoll <= BashChances[skill][0] { // Thunk
		damModifier = CombatModifiers["thunk"]
		output = "Thunk!!"
	} else if bashRoll <= BashChances[skill][1] { // Crushing
		damModifier = CombatModifiers["crushing"]
		output = "Craaackk!!"
	} else if bashRoll <= BashChances[skill][2] { // Thwomp
		damModifier = CombatModifiers["thwomp"]
		output = "Thwomp!!"
	} else if bashRoll <= BashChances[skill][3] { // Thump
		stunModifier = 3
		output = "Thump!!"
	}
	return
}

func RollLethal(skill int) bool {
	return RollLethalScaled(skill, 1)
}

// RollLethalScaled rolls a lethal blow with the table chance multiplied by
// multiplier (see ExecuteMultiplier for the fighter's execute scaling).
func RollLethalScaled(skill int, multiplier float64) bool {
	if skill > 0 {
		dRoll := utils.Roll(1000000, 1, 0)
		if float64(dRoll) <= float64(LethalDamage[skill])*multiplier {
			return true
		}
	}
	return false
}

func BreatheDamage(level int) int {
	switch {
	case level < 5:
		return 8
	case level < 10:
		return 20
	case level < 15:
		return 40
	case level < 20:
		return 90
	case level < 25:
		return 125
	default:
		return 8
	}
}

var XP_Modifiers = map[string]float64{
	"poisons":       .03,
	"fast_moving":   .03,
	"block_exit":    .06,
	"follows":       .06,
	"no_stun":       .03,
	"diseases":      .03,
	"spits_acid":    .03,
	"blinds":        .06,
	"no_steal":      .03,
	"ranged_attack": .03,
	"invisible":     .03,
	"steals":        .06,
}

// ThreatPercent is percent of a mob's max stamina, for the threat bumps above.
func ThreatPercent(maxStam int, percent int) int {
	return maxStam * percent / 100
}

// BackstabMissChance is the miss chance a backstab starts from, before dex,
// level difference and combat flags are applied: the inverse of the base hit
// chance plus the stealth bonus. Dex and the rest are shared with the
// weapon miss chance, which is why they are not here.
func BackstabMissChance(stealthLevel int) int {
	return 100 - BackStabChance - stealthLevel*BackStabChancePerSkillLevel
}

// FailedTurnThreat is the threat a failed turn hands the caster: the mob's
// current stamina, capped at FailedTurnThreatCapPercent of its max.
func FailedTurnThreat(current int, max int) int {
	cap := ThreatPercent(max, FailedTurnThreatCapPercent)
	if current > cap {
		return cap
	}
	return current
}

// MobStunBase is how long (seconds) a mob's stun spell holds a player, and
// MobStunMin is the floor once level difference is subtracted.
const (
	MobStunBase = 20
	MobStunMin  = 1
)

// MobStunDuration is how long a mob of mobLevel stuns a player of playerTier.
// A lower-level mob loses one second per level it is below the player, down
// to MobStunMin; a mob at or above the player's level stuns for the full duration.
func MobStunDuration(mobLevel int, playerTier int) int {
	duration := MobStunBase
	if mobLevel < playerTier {
		duration -= playerTier - mobLevel
	}
	if duration < MobStunMin {
		duration = MobStunMin
	}
	return duration
}
