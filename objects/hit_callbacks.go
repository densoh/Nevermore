package objects

import (
	"log"
	"strconv"
	"strings"

	"github.com/ArcCS/Nevermore/config"
	"github.com/ArcCS/Nevermore/data"
	"github.com/ArcCS/Nevermore/text"
	"github.com/ArcCS/Nevermore/utils"
)

// A hit callback is a one-shot function a hit can fire on a mob: a spell
// landing today, a melee hit later. It can do anything; ignite, for one,
// puts a ticking burn effect on the mob.

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
	"ignite": ignite,
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

// ignite sets the mob burning. Recasting replaces the burn, so only one runs.
func ignite(ctx *HitContext, target *Mob) {
	room, ok := Rooms[target.ParentId]
	if !ok {
		return
	}
	target.ReplaceEffect("burn", strconv.Itoa(config.BurnTicks*config.BurnInterval), config.BurnInterval, 0,
		func(triggers int) {
			burnTick(ctx, target, triggers)
		},
		func() {
			room.MessageAll(text.Info + "The flames on " + target.Name + " die out.\n" + text.Reset)
		})
	room.MessageAll(text.Red + target.Name + " bursts into flame!\n" + text.Reset)
}

// burnTick deals one tick of burn damage. Tick 0 fires as the burn is applied
// and does nothing; the hit that lit it already landed.
func burnTick(ctx *HitContext, target *Mob, triggers int) {
	if triggers == 0 || triggers > config.BurnTicks || target.Stam.Current <= 0 {
		return
	}
	room, ok := Rooms[target.ParentId]
	if !ok {
		return
	}
	credit := burnCredit(ctx.CasterName, target, room)
	if credit == nil {
		target.RemoveEffect("burn")
		room.MessageAll(text.Info + "The flames on " + target.Name + " die out.\n" + text.Reset)
		return
	}

	damage, _, resisted := target.ReceiveMagicDamage(burnDamage(ctx), ctx.Element)
	data.StoreCombatMetric(ctx.Element+"spell_burn", 0, 2, damage+resisted, resisted, damage, 0, credit.CharId, credit.Tier, 1, target.MobId)
	target.AddThreatDamage(damage, credit)
	if _, elemental := magicSkillMap[ctx.Element]; elemental && credit.Name == ctx.CasterName {
		credit.AdvanceElementalExp(int(float64(damage)/float64(target.Stam.Max)*float64(target.Experience)), ctx.Element, credit.Class)
	}
	room.MessageAll(text.Red + target.Name + " burns for " + strconv.Itoa(damage) + " " + ctx.Element + " damage.\n" + text.Reset)
	target.DeathCheck(credit)
}

// burnDamage rolls 1d(caster tier) and scales it by the caster's spell bonus.
func burnDamage(ctx *HitContext) int {
	return int(float64(utils.Roll(ctx.Tier, 1, 0)) * ctx.Bonus)
}

// burnCredit is who a burn tick counts for: the caster if they're in the
// room, else the character in the room with the most threat on the mob, else
// nobody.
func burnCredit(casterName string, target *Mob, room *Room) *Character {
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
