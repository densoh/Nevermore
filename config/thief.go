package config

import "strings"

// Sneak attack: from SneakAttackTier, a thief's regular melee hit on a mob that
// isn't targeting them adds SneakAttackDice d SneakAttackSides damage. It goes
// on after the crit/double multiplier (crits don't multiply it) and before mob
// armor. Backstab never triggers it.
const (
	SneakAttackTier  = 10
	SneakAttackSides = 6
	// Each other mob targeting the thief takes this many points off the
	// chance, so four or more attackers shut it off.
	SneakAttackPenaltyPerAttacker = 25
)

// SneakAttackDice is how many dice a sneak attack rolls:
// floor((tier-10)/3 + 1 + stealthRank/2), with the fractions added together
// before rounding down. 0 below SneakAttackTier.
func SneakAttackDice(tier, stealthRank int) int {
	if tier < SneakAttackTier {
		return 0
	}
	// Common denominator of 6 so the thirds and halves sum before truncating.
	return (2*(tier-SneakAttackTier) + 6 + 3*stealthRank) / 6
}

// SneakAttackChance is the percent chance a sneak attack lands when the hit
// mob isn't targeting the thief but otherAttackers other mobs are.
func SneakAttackChance(otherAttackers int) int {
	return max(0, 100-SneakAttackPenaltyPerAttacker*otherAttackers)
}

// Devious strikes: a thief queues one with "devious <strike>" and it fires in
// place of the sneak attack on their next hit that would sneak attack. Only
// one is queued at a time; it waits through target switches and is dropped
// when no mob left in the room has the thief on its threat table.
const (
	DeviousPoison = "poison"
	DeviousExpose = "expose-weakness"
)

// DeviousStrikeTiers is each strike and the tier it unlocks at.
var DeviousStrikeTiers = map[string]int{
	DeviousPoison: 10,
	DeviousExpose: 10,
}

// DeviousCooldown is the seconds between changes of readied devious strike.
// Readying isn't a combat action, so nothing else (stun, the global timer,
// haste) touches it.
const DeviousCooldown = 1

// DeviousStrikeOrder lists the strikes for help and status output.
var DeviousStrikeOrder = []string{DeviousPoison, DeviousExpose}

// A devious strike is learned from a scroll whose spell is DeviousSpellPrefix
// plus the strike's name ("devious-poison"); the thief keeps it in their
// spell list under that name, the way bards keep songs.
const DeviousSpellPrefix = "devious-"

// DeviousSpell is the spell-list name a strike is learned under.
func DeviousSpell(strike string) string {
	return DeviousSpellPrefix + strike
}

// DeviousFromSpell reports the strike a scroll or spell-list name teaches.
func DeviousFromSpell(spell string) (string, bool) {
	strike, ok := strings.CutPrefix(spell, DeviousSpellPrefix)
	if !ok {
		return "", false
	}
	_, ok = DeviousStrikeTiers[strike]
	return strike, ok
}

const (
	// The poison strike ticks PoisonStrikeTicks times, PoisonStrikeInterval
	// seconds apart, for PoisonStrikeDice d PoisonStrikeSides each tick.
	PoisonStrikeTicks    = 4
	PoisonStrikeInterval = 8
	PoisonStrikeSides    = 4

	// Expose weakness takes this much armor off the mob until it dies (or,
	// for a permanent mob, until its room resets). It doesn't stack.
	ExposeWeaknessArmor = 20
)

// PoisonStrikeDice is floor(1 + stealthRank/4 + (tier-10)/5), the fractions
// summed before rounding down; never below 1.
func PoisonStrikeDice(tier, stealthRank int) int {
	// Common denominator of 20 so the quarters and fifths sum first.
	return max(1, (20+5*stealthRank+4*(tier-SneakAttackTier))/20)
}
