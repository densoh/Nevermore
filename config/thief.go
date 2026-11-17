package config

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
