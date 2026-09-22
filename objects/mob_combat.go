package objects

import (
	"log"
	"math"
	"strconv"

	"github.com/ArcCS/Nevermore/config"
	"github.com/ArcCS/Nevermore/data"
	"github.com/ArcCS/Nevermore/text"
	"github.com/ArcCS/Nevermore/utils"
)

// AttackStyle is the special-hit outcome of a mob strike against a character.
type AttackStyle int

const (
	StyleNormal AttackStyle = iota
	StyleVital
	StyleCritical
	StyleDouble
)

// StrikeOpts carries the per-caller flavor for ApplyStrike: metric naming,
// message prefix, and the death message used if the strike kills the target.
type StrikeOpts struct {
	// Metric is the base combat-metric name ("melee", "range", "follow", ...).
	// Vital strikes record as Metric+"_vital", reflects as Metric+"_player_reflect".
	Metric string
	// Mode is passed straight through to data.StoreCombatMetric.
	Mode int
	// HitPrefix is prepended to the hit line, e.g. "Thwwip!! " for ranged.
	HitPrefix string
	// DeathMsg is handed to the target's death check.
	DeathMsg string
}

// RollSpecial decides whether a landed hit is a vital, critical, or double
// damage strike, and returns the damage multiplier for it. Vital and critical
// multipliers are reduced by the target's dex. A mob flagged no_specials
// always strikes normally.
func (m *Mob) RollSpecial(target *Character) (AttackStyle, float64) {
	if m.Flags["no_specials"] {
		return StyleNormal, 1
	}
	if utils.Roll(10, 1, 0) > 1 {
		return StyleNormal, 1
	}
	dex := float64(target.GetStat("dex"))
	styleRoll := utils.Roll(10, 1, 0)
	switch {
	case styleRoll <= config.MobVital:
		return StyleVital, 2 - (dex / 100)
	case styleRoll <= config.MobCritical:
		return StyleCritical, 4 - (dex / 50)
	case styleRoll <= config.MobDouble:
		return StyleDouble, 2
	}
	return StyleNormal, 1
}

// RollMiss rolls the mob's chance to miss the target based on level
// difference and the target's dex. On a miss it tells the player and records
// a metricPrefix+"-miss" metric, returning true.
func (m *Mob) RollMiss(target *Character, metricPrefix string) bool {
	missChance := 0
	lvlDiff := target.Tier - m.Level
	if lvlDiff >= 1 {
		missChance += lvlDiff * config.MissPerLevel
	}
	missChance += target.GetStat("dex") * config.HitPerDex
	if utils.Roll(100, 1, 0) > missChance {
		return false
	}
	target.writeCombat(text.Green + m.Name + " missed you!!" + "\n" + text.Reset)
	data.StoreCombatMetric(metricPrefix+"-miss", 0, 1, 0, 0, 0, 1, m.MobId, m.Level, 0, target.CharId)
	return true
}

// ApplyStrike resolves a hit that has already been rolled: it scales
// baseDamage by mult, applies it to the target (vitality only for vital
// strikes), records the metric, prints the banner and hit line, reflects
// damage back to the mob if the target has reflection, and runs the target's
// death check. It returns true if the target died.
//
// Callers own the "attacked" hook and any touch effects; run those first.
func (m *Mob) ApplyStrike(target *Character, baseDamage int, style AttackStyle, mult float64, opts StrikeOpts) bool {
	totalDamage := int(math.Ceil(float64(baseDamage) * mult))
	stamDamage, vitDamage, resisted := 0, 0, 0
	if style == StyleVital {
		vitDamage, resisted = target.ReceiveVitalDamage(totalDamage)
		data.StoreCombatMetric(opts.Metric+"_vital", 0, opts.Mode, totalDamage, resisted, vitDamage, 1, m.MobId, m.Level, 0, target.CharId)
	} else {
		stamDamage, vitDamage, resisted = target.ReceiveDamage(totalDamage)
		data.StoreCombatMetric(opts.Metric, 0, opts.Mode, totalDamage, resisted, stamDamage+vitDamage, 1, m.MobId, m.Level, 0, target.CharId)
	}

	if stamDamage == 0 && vitDamage == 0 {
		target.writeCombat(text.Red + m.Name + " attacks bounces off of you for no damage!" + "\n" + text.Reset)
	} else {
		switch style {
		case StyleVital:
			target.writeCombat(text.Red + "Vital Strike!!!\n" + text.Reset)
		case StyleCritical:
			target.writeCombat(text.Red + "Critical Strike!!!\n" + text.Reset)
		case StyleDouble:
			target.writeCombat(text.Red + "Double Damage!!!\n" + text.Reset)
		}
		buildString := ""
		if stamDamage != 0 {
			buildString += strconv.Itoa(stamDamage) + " stamina"
		}
		if stamDamage != 0 && vitDamage != 0 {
			buildString += " and "
		}
		if vitDamage != 0 {
			buildString += strconv.Itoa(vitDamage) + " vitality"
		}
		target.writeCombat(text.Red + opts.HitPrefix + m.Name + " attacks you for " + buildString + " points of damage!" + "\n" + text.Reset)
	}

	if target.CheckFlag("reflection") {
		reflectDamage := int(float64(baseDamage) * (float64(target.GetStat("int")) * config.ReflectDamagePerInt))
		mobFin, _, mobResisted := m.ReceiveDamage(reflectDamage)
		data.StoreCombatMetric(opts.Metric+"_player_reflect", 0, opts.Mode, reflectDamage, mobResisted, mobFin, 0, target.CharId, target.Tier, 1, m.MobId)
		target.writeCombat(text.Cyan + "You reflect " + strconv.Itoa(reflectDamage) + " damage back to " + m.Name + "!\n" + text.Reset)
		m.DeathCheck(target)
	}

	return target.DeathCheckBool(opts.DeathMsg)
}

// writeCombat writes a combat line to the player, logging any write error.
func (c *Character) writeCombat(line string) {
	if _, err := c.Write([]byte(line)); err != nil {
		log.Println("Error writing to player:", err)
	}
}
