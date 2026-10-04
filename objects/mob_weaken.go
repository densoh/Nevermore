package objects

import "math"

// ValidWeaken reports whether pct and maxSteps describe a weakening that
// never drives the mob's stats to zero or below.
func ValidWeaken(pct, maxSteps int) bool {
	return pct > 0 && maxSteps > 0 && pct*maxSteps < 100
}

// Weaken applies one offering step to m, which should be the mob's template
// in Mobs, and reports false without changing anything once maxSteps have
// been applied. Each step takes pct% of the original armor, positive
// resistances, max stamina and damage dice, so after n steps those stats sit
// at (100 - n*pct)% of what the builder set.
//
// The original values aren't kept anywhere: step n rescales the current
// values by (100-n*pct)/(100-(n-1)*pct), which lands in the same place give
// or take rounding. The caller saves the template and syncs live instances
// with SyncWeakened.
func (m *Mob) Weaken(pct, maxSteps int) bool {
	if !ValidWeaken(pct, maxSteps) || m.WeakenSteps >= maxSteps {
		return false
	}
	before := 100 - m.WeakenSteps*pct
	m.WeakenSteps++
	after := 100 - m.WeakenSteps*pct

	scale := func(v int) int {
		return int(math.Round(float64(v) * float64(after) / float64(before)))
	}
	// Negative resistances are weaknesses; shrinking them would make the
	// mob tougher, so only positive ones come down.
	scaleResist := func(v int) int {
		if v <= 0 {
			return v
		}
		return scale(v)
	}

	m.Armor = scale(m.Armor)
	m.WaterResistance = scaleResist(m.WaterResistance)
	m.AirResistance = scaleResist(m.AirResistance)
	m.FireResistance = scaleResist(m.FireResistance)
	m.EarthResistance = scaleResist(m.EarthResistance)
	m.Stam.Max = max(scale(m.Stam.Max), 1)
	m.Stam.Current = m.Stam.Max
	m.SidesDice = max(scale(m.SidesDice), 1)
	m.PlusDice = scale(m.PlusDice)
	return true
}

// SyncWeakened copies the stats Weaken changes from the template onto a live
// instance. Current stamina keeps the same fraction of max it had, so a
// wounded boss stays just as wounded.
func (m *Mob) SyncWeakened(template *Mob) {
	if m.Stam.Max > 0 {
		m.Stam.Current = int(math.Round(float64(m.Stam.Current) * float64(template.Stam.Max) / float64(m.Stam.Max)))
	}
	m.Stam.Max = template.Stam.Max
	m.Armor = template.Armor
	m.WaterResistance = template.WaterResistance
	m.AirResistance = template.AirResistance
	m.FireResistance = template.FireResistance
	m.EarthResistance = template.EarthResistance
	m.SidesDice = template.SidesDice
	m.PlusDice = template.PlusDice
	m.WeakenSteps = template.WeakenSteps
}
