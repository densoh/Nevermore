package objects

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/ArcCS/Nevermore/config"
	"github.com/ArcCS/Nevermore/data"
	"github.com/ArcCS/Nevermore/text"
	"github.com/ArcCS/Nevermore/utils"
)

// A hit callback is a one-shot function a hit can fire on a mob: a spell
// landing today, a melee hit later. It can do anything; applyDoT builds the
// ones that put a damage-over-time effect on the mob.

// HitContext is what a hit callback, and any effect it applies, knows about
// where it came from.
type HitContext struct {
	CasterName string
	Tier       int
	Bonus      float64 // the caster's spell damage multiplier when it fired
	Element    string
}

type HitCallback func(ctx *HitContext, target *Mob)

// OnHit attaches a named hit callback to a source, firing Chance percent of
// the time.
type OnHit struct {
	Name   string
	Chance int
}

var HitCallbacks = map[string]HitCallback{
	"ignite": applyDoT(burn),
}

// TriggerOnHit rolls each callback's chance and runs the ones that fire.
func TriggerOnHit(callbacks []OnHit, ctx *HitContext, target *Mob) {
	for _, cb := range callbacks {
		fn, ok := HitCallbacks[cb.Name]
		if !ok {
			log.Println("Unknown hit callback:", cb.Name)
			continue
		}
		if target.Stam.Current <= 0 {
			return
		}
		if cb.Chance < 100 && utils.Roll(100, 1, 0) > cb.Chance {
			continue
		}
		fn(ctx, target)
	}
}

// SpellHitCallbacks fires a spell's OnSpellHit callbacks after it lands on a
// mob. Call it while the caster's "casting" flag is still on so the bonus
// matches the hit's.
func SpellHitCallbacks(caster *Character, spell Spell, target *Mob) {
	if len(spell.OnSpellHit) == 0 || target.Stam.Current <= 0 {
		return
	}
	element := strings.TrimSuffix(spell.Effect, "-damage")
	ctx := &HitContext{
		CasterName: caster.Name,
		Tier:       caster.Tier,
		Bonus:      spellBonusMultiplier(caster, element),
		Element:    element,
	}
	TriggerOnHit(spell.OnSpellHit, ctx, target)
}

// A DoT is a damage-over-time effect a hit callback can put on a mob. Each
// tick rolls Damage and deals it as the hit's element; applying it again
// replaces the running one, so only one of each Effect runs on a mob. The
// messages take the mob's name as %s.
type DoT struct {
	Effect   string // the effect's name on the mob, also used in its combat metric
	Ticks    int
	Interval int // seconds between ticks
	Damage   func(ctx *HitContext) int
	OnMsg    string
	TickVerb string // "<mob> <verb> for N <element> damage."
	OffMsg   string
}

// burn is combust's ignite: 2d(caster tier/2 + 1)+2 fire damage a tick.
var burn = DoT{
	Effect:   "burn",
	Ticks:    config.BurnTicks,
	Interval: config.BurnInterval,
	Damage:   tierRollDamage,
	OnMsg:    "%s bursts into flame!",
	TickVerb: "burns",
	OffMsg:   "The flames on %s die out.",
}

// applyDoT makes a hit callback that puts dot on the mob.
func applyDoT(dot DoT) HitCallback {
	return func(ctx *HitContext, target *Mob) {
		room, ok := Rooms[target.ParentId]
		if !ok {
			return
		}
		target.ReplaceEffect(dot.Effect, strconv.Itoa(dot.Ticks*dot.Interval), dot.Interval, 0,
			func(triggers int) {
				dotTick(dot, ctx, target, triggers)
			},
			func() {
				room.MessageAll(text.Info + fmt.Sprintf(dot.OffMsg, target.Name) + "\n" + text.Reset)
			})
		room.MessageAll(text.Red + fmt.Sprintf(dot.OnMsg, target.Name) + "\n" + text.Reset)
	}
}

// dotTick deals one tick of a DoT. Tick 0 fires as the DoT is applied and
// does nothing; the hit that applied it already landed.
func dotTick(dot DoT, ctx *HitContext, target *Mob, triggers int) {
	if triggers == 0 || triggers > dot.Ticks || target.Stam.Current <= 0 {
		return
	}
	room, ok := Rooms[target.ParentId]
	if !ok {
		return
	}
	credit := dotCredit(ctx.CasterName, target, room)
	if credit == nil {
		target.RemoveEffect(dot.Effect)
		room.MessageAll(text.Info + fmt.Sprintf(dot.OffMsg, target.Name) + "\n" + text.Reset)
		return
	}

	damage, _, resisted := target.ReceiveMagicDamage(dot.Damage(ctx), ctx.Element)
	data.StoreCombatMetric(ctx.Element+"spell_"+dot.Effect, 0, 2, damage+resisted, resisted, damage, 0, credit.CharId, credit.Tier, 1, target.MobId)
	target.AddThreatDamage(damage, credit)
	if _, elemental := magicSkillMap[ctx.Element]; elemental && credit.Name == ctx.CasterName {
		credit.AdvanceElementalExp(int(float64(damage)/float64(target.Stam.Max)*float64(target.Experience)), ctx.Element, credit.Class)
	}
	room.MessageAll(text.Red + target.Name + " " + dot.TickVerb + " for " + strconv.Itoa(damage) + " " + ctx.Element + " damage.\n" + text.Reset)
	target.DeathCheck(credit)
}

// tierRollDamage rolls 2d(caster tier/2 + 1)+2 and scales it by the caster's
// spell bonus.
func tierRollDamage(ctx *HitContext) int {
	return int(float64(utils.Roll(ctx.Tier/2+1, 2, 2)) * ctx.Bonus)
}

// dotCredit is who a DoT tick counts for: the caster if they're in the
// room, else the character in the room with the most threat on the mob, else
// nobody.
func dotCredit(casterName string, target *Mob, room *Room) *Character {
	var top *Character
	topThreat := 0
	for _, c := range room.Chars.Contents {
		if strings.EqualFold(c.Name, casterName) {
			return c
		}
		if threat, ok := target.ThreatTable[c.Name]; ok && (top == nil || threat > topThreat) {
			top, topThreat = c, threat
		}
	}
	return top
}
