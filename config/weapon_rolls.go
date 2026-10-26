package config

// Standards for drop weapon rolls, quoted against the real cap (the highest
// max roll a tier can wield, one under WeaponDamageCap since CanWield is strict).

// WeaponRollMaxBelowCap is the range a standard drop's max roll sits under the
// real cap, inclusive.  A lower max with the same average narrows the roll.
const WeaponRollMaxBelowCapMin = 2
const WeaponRollMaxBelowCapMax = 4

// WeaponRollMinPercent is the least a standard drop's min roll may be, as a
// percent of the real cap.
const WeaponRollMinPercent = 50

// WeaponRollTargetPercent is the share of the real cap a tier's drops should
// average (tier median and mean).
func WeaponRollTargetPercent(tier int) int {
	switch {
	case tier >= 20:
		return 84
	case tier >= 15:
		return 82
	case tier >= 10:
		return 80
	default:
		return 78
	}
}

// WeaponRollSpecialBonusPercent is how many points of the cap a special item
// averages above the tier's target, a rule of thumb.  Special items roll max at cap.
const WeaponRollSpecialBonusPercent = 8

// RealWeaponCap is the highest max roll a non-fighter at the tier can wield.
func RealWeaponCap(tier int, twoHanded bool) int {
	cap := MaxWeaponDamage[tier]
	if twoHanded {
		cap += TwoHandedCapBonus(cap)
	}
	return cap - 1
}

type WeaponRoll struct {
	Dice, Sides, Plus int
}

func (r WeaponRoll) Min() int { return r.Dice + r.Plus }
func (r WeaponRoll) Max() int { return r.Dice*r.Sides + r.Plus }

// Avg2 is twice the average roll, so half points stay whole.
func (r WeaponRoll) Avg2() int { return r.Min() + r.Max() }

// WeaponRollSingleDieMaxSides is the biggest die a lone-die roll may use.
const WeaponRollSingleDieMaxSides = 20

// WeaponRollMaxDice is the most dice a roll may use for a die size: 8d2, 10d3, 12d4...
func WeaponRollMaxDice(sides int) int {
	return 4 + 2*sides
}

// StandardDiceShape reports whether a dice count and die size are allowed for
// a drop: a lone die up to d20, no more than WeaponRollMaxDice, no d1.
func StandardDiceShape(dice int, sides int) bool {
	if dice < 1 || sides < 2 || dice > WeaponRollMaxDice(sides) {
		return false
	}
	return dice > 1 || sides <= WeaponRollSingleDieMaxSides
}

// StandardWeaponRolls lists every roll that meets the drop standards for a
// real cap: max within cap-2..cap-4, min at least WeaponRollMinPercent of the
// cap, and a StandardDiceShape.  Sorted by max (high first), then average
// (high first), then dice count.
func StandardWeaponRolls(realCap int) []WeaponRoll {
	var rolls []WeaponRoll
	for max := realCap - WeaponRollMaxBelowCapMin; max >= realCap-WeaponRollMaxBelowCapMax; max-- {
		// The spread n*(s-1) only shrinks the min, so walk it from tightest up.
		for spread := 1; ; spread++ {
			min := max - spread
			if min*100 < realCap*WeaponRollMinPercent {
				break
			}
			for dice := 1; dice <= spread; dice++ {
				if spread%dice != 0 {
					continue
				}
				r := WeaponRoll{dice, spread/dice + 1, min - dice}
				if r.Plus >= 0 && StandardDiceShape(r.Dice, r.Sides) {
					rolls = append(rolls, r)
				}
			}
		}
	}
	return rolls
}

// WeaponTier is the lowest tier a non-fighter can wield a weapon at, or 0 if
// its max roll is past every cap.
func WeaponTier(max int, twoHanded bool) int {
	for tier := 1; tier <= len(MaxWeaponDamage); tier++ {
		if max <= RealWeaponCap(tier, twoHanded) {
			return tier
		}
	}
	return 0
}
