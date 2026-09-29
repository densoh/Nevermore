package cmd

import (
	"log"
	"math"
	"strconv"

	"github.com/ArcCS/Nevermore/data"

	"github.com/ArcCS/Nevermore/config"
	"github.com/ArcCS/Nevermore/objects"
	"github.com/ArcCS/Nevermore/permissions"
	"github.com/ArcCS/Nevermore/text"
	"github.com/ArcCS/Nevermore/utils"
)

func init() {
	addHandler(kill{},
		"Usage:  kill target # \n\n Try to attack something! Can also use attack, or shorthand k",
		permissions.Player,
		"kill", "k")
}

type kill cmd

func (kill) process(s *state) {
	if len(s.input) < 1 && s.actor.Victim == nil {
		s.msg.Actor.SendBad("Attack what exactly?")
		return
	}

	if s.actor.CheckFlag("blind") {
		s.msg.Actor.SendBad("You can't see anything!")
		return
	}

	if s.actor.Stam.Current <= 0 {
		s.msg.Actor.SendBad("You are far too tired to do that.")
		return
	}

	// Check some timers
	ready, msg := s.actor.TimerReady("combat")
	if !ready {
		s.msg.Actor.SendBad(msg)
		return
	}

	name := ""
	nameNum := 1
	if len(s.words) < 1 && s.actor.Victim != nil {
		switch s.actor.Victim.(type) {
		case *objects.Character:
			name = s.actor.Victim.(*objects.Character).Name
		case *objects.Mob:
			name = s.actor.Victim.(*objects.Mob).Name
			nameNum = s.where.Mobs.GetNumber(s.actor.Victim.(*objects.Mob))
		}
	} else {
		name = s.input[0]
	}

	if len(s.words) > 1 {
		// Try to snag a number off the list
		if val, err := strconv.Atoi(s.words[1]); err == nil {
			nameNum = val
		}
	}

	var whatMob *objects.Mob
	whatMob = s.where.Mobs.Search(name, nameNum, s.actor)
	if whatMob != nil {
		s.actor.Victim = whatMob
		s.actor.MarkCombat()

		// This is an override for a GM to delete a mob
		if s.actor.Permission.HasAnyFlags(permissions.Builder, permissions.Dungeonmaster, permissions.Gamemaster) {
			s.msg.Actor.SendInfo("You smashed ", whatMob.Name, " out of existence.")
			s.actor.Victim = nil
			objects.Rooms[whatMob.ParentId].Mobs.Remove(whatMob)
			whatMob = nil
			return
		}

		s.actor.RunHook("combat")

		// Shortcut a missing weapon:
		if s.actor.Equipment.Main == (*objects.Item)(nil) && s.actor.Class != config.MONK {
			s.msg.Actor.SendBad("You have no weapon to attack with.")
			return
		}

		if _, err := whatMob.ThreatTable[s.actor.Name]; !err {
			s.msg.Actor.Send(text.White + "You engaged " + whatMob.Name + " #" + strconv.Itoa(s.where.Mobs.GetNumber(whatMob)) + " in combat.")
			s.msg.Observers.Send(text.White + s.actor.Name + " attacks " + whatMob.Name)
			whatMob.AddThreatDamage(0, s.actor)
		}

		// Shortcut target not being in the right location, check if it's a missile weapon, or that they are placed right.
		weapon := s.actor.Equipment.Main
		if s.actor.Class == config.MONK {
			weapon = (*objects.Item)(nil)
		}
		if inRange, rangeMsg := weaponInRange(s, weapon, whatMob); !inRange {
			s.msg.Actor.SendBad(rangeMsg)
			return
		}

		performAttack(s, whatMob)
		return

	}

	s.msg.Actor.SendInfo("Attack what?")
	s.ok = true
}

// weaponInRange reports whether the mob is at a distance the weapon can reach, and
// why not when it isn't.  A nil weapon is an unarmed attack.
func weaponInRange(s *state, weapon *objects.Item, whatMob *objects.Mob) (bool, string) {
	if weapon == (*objects.Item)(nil) {
		if s.actor.Placement != whatMob.Placement {
			return false, "You are too far away to attack."
		}
		return true, ""
	}

	if (weapon.ItemType != 4 && weapon.ItemType != 3) && (s.actor.Placement != whatMob.Placement) {
		return false, "You are too far away to attack."
	} else if weapon.ItemType == 4 && (s.actor.Placement == whatMob.Placement) {
		return false, "You are too close to attack."
	} else if weapon.ItemType == 3 && (s.actor.Placement == whatMob.Placement) {
		return false, "You are too close to attack."
	} else if weapon.ItemType == 3 && (int(math.Abs(float64(s.actor.Placement-whatMob.Placement))) > 1) {
		return false, "You are too far away to attack."
	}

	return true, ""
}

// performAttack runs a standard weapon attack against the mob with whatever is in the
// main hand: the fighter lethal and ranger snipe rolls, the multi attack chain, mob
// reflection, the death check, weapon wear and the attack timer.
//
// The multi attack resolves in two passes.  Every swing rolls to hit first, with the
// first swing at the normal miss chance and follow-ups at a penalised one.  The
// damage multipliers are then handed out to the landed hits in order, so whichever
// swing connects first always takes the full damage slot and any critical or double
// roll rides on it.
func performAttack(s *state, whatMob *objects.Mob) {
	recklessAttack(s)
	// use a list of attacks,  so we can expand this later if other classes get multi style attacks
	attacks := []float64{
		1.0,
	}

	skillLevel := attackSkillLevel(s)
	// Kill is really the fighters realm for specialty.
	if s.actor.Permission.HasAnyFlags(permissions.Fighter) && s.actor.Equipment.Main.ItemType != 4 {
		// mob lethal? A tier 15 expert executes wounded mobs more often.
		execute := config.ExecuteMultiplier(s.actor.Tier, skillLevel, float64(whatMob.Stam.Current)/float64(whatMob.Stam.Max))
		if config.RollLethalScaled(skillLevel, execute) {
			// Sure did.  Kill this fool and bail.
			s.msg.Actor.SendInfo("You landed a lethal blow on the " + whatMob.Name)
			s.msg.Observers.SendInfo(s.actor.Name + " landed a lethal blow on " + whatMob.Name)
			s.actor.Equipment.DamageWeapon("main", 1)
			data.StoreCombatMetric("lethal", 0, 0, whatMob.Stam.Current, 0, 0, 0, s.actor.CharId, s.actor.Tier, 1, whatMob.MobId)
			whatMob.AddThreatDamage(whatMob.Stam.Current, s.actor)
			lethalFraction := math.Max(float64(whatMob.Stam.Current)/float64(whatMob.Stam.Max), config.LethalSkillFloor)
			s.actor.AdvanceSkillExp(lethalFraction * float64(whatMob.Experience))
			whatMob.Stam.Current = 0
			DeathCheck(s, whatMob)
			s.actor.SetTimer("combat", config.CombatCooldown)
			return
		}
	}

	if s.actor.Permission.HasAnyFlags(permissions.Ranger) && s.actor.Equipment.Main.ItemType == 4 && s.actor.Tier > 7 {
		// Sniper
		if utils.Roll(1000, 1, 0) == 1 {
			// Throw a snipe
			whatMob.AddThreatDamage(whatMob.Stam.Max/10, s.actor)
			actualDamage, _, resisted := whatMob.ReceiveDamage(int(math.Ceil(float64(s.actor.InflictDamage()) * config.CombatModifiers["snipe"])))
			data.StoreCombatMetric("snipe", 0, 0, actualDamage+resisted, resisted, actualDamage, 0, s.actor.CharId, s.actor.Tier, 1, whatMob.MobId)
			s.msg.Actor.SendInfo("You sniped the " + whatMob.Name + " for " + strconv.Itoa(actualDamage) + " damage!" + text.Reset)
			s.actor.AdvanceSkillExp((float64(actualDamage) / float64(whatMob.Stam.Max) * float64(whatMob.Experience)))
			s.msg.Observers.SendInfo(s.actor.Name + " snipes " + whatMob.Name)
			if whatMob.CheckFlag("reflection") {
				reflectDamage := int(float64(actualDamage) * config.ReflectDamageFromMob)
				stamDamage, vitDamage, resisted := s.actor.ReceiveDamage(reflectDamage)
				data.StoreCombatMetric("snipe_mob_reflect", 0, 0, stamDamage+vitDamage+resisted, resisted, stamDamage+vitDamage, 1, whatMob.MobId, whatMob.Level, 0, s.actor.CharId)
				s.msg.Actor.Send("The " + whatMob.Name + " reflects " + strconv.Itoa(reflectDamage) + " damage back at you!")
				s.actor.DeathCheck(" was killed by reflection!")
			}
			DeathCheck(s, whatMob)
			s.actor.SetTimer("combat", config.CombatCooldown)
			return
		}
	}

	flurried := false
	if s.actor.Permission.HasAnyFlags(permissions.Fighter) || (s.actor.Permission.HasAnyFlags(permissions.Ranger) && s.actor.Equipment.Main.ItemType == 4) {
		if mults, ok := config.MultiAttackMultipliers[skillLevel]; ok {
			attacks = mults
		}
	} else if s.actor.Class == config.MONK {
		attacks, flurried = monkAttackPlan(s, skillLevel)
	}

	result := resolveHits(s, whatMob, attacks, skillLevel, "kill")
	if s.actor.Class == config.MONK && result.hits > 0 && !flurried {
		gainChiFromHits(s, result.chiHits())
	}
	DeathCheck(s, whatMob)
	if s.actor.Class != config.MONK {
		weapMsg := s.actor.Equipment.DamageWeapon("main", result.weaponDamage)
		if weapMsg != "" {
			s.msg.Actor.SendInfo(weapMsg)
		}
	}

	s.actor.SetTimer("combat", config.CombatCooldown)
}

// canCrush reports whether the actor's regular attacks can land a crushing
// blow: barbarians wielding a blunt or two-handed melee weapon.
func canCrush(s *state) bool {
	main := s.actor.Equipment.Main
	if s.actor.Class != config.BARBARIAN || main == nil || main.ItemType == 4 {
		return false
	}
	return main.ItemType == 2 || main.IsTwoHanded()
}

// attackSkillLevel is the weapon skill level behind the actor's attacks:
// hand-to-hand for monks, otherwise the wielded weapon's skill.
func attackSkillLevel(s *state) int {
	if s.actor.Class == config.MONK {
		return config.WeaponLevel(s.actor.Skills[5].Value, s.actor.Class, 5)
	}
	return config.WeaponLevel(s.actor.Skills[s.actor.Equipment.Main.ItemType].Value, s.actor.Class, s.actor.Equipment.Main.ItemType)
}

// monkAttackPlan returns the swing multipliers for a monk's attack round and
// whether the flurry stance paid for them. Flurry costs chi every round; when
// the monk can't cover it the stance drops and the round is a single swing.
func monkAttackPlan(s *state, skillLevel int) ([]float64, bool) {
	single := []float64{1.0}
	if !s.actor.CheckFlag("flurry") {
		return single, false
	}
	cost := config.FlurryChiCost(skillLevel)
	if s.actor.Mana.Current < cost {
		s.actor.RemoveEffect("flurry")
		s.msg.Actor.SendBad("You are too drained to keep up the flurry.")
		return single, false
	}
	s.actor.Mana.Subtract(cost)
	return config.MonkFlurryFor(skillLevel), true
}

// gainChiFromHits awards a monk chi for landed hits and tells them about it.
func gainChiFromHits(s *state, hits int) {
	gained := s.actor.GainChi(config.ChiPerHitFor(s.actor.GetStat("pie"))*hits, false)
	if gained > 0 {
		s.msg.Actor.Send(text.Cyan + "Your chi rises by " + strconv.Itoa(gained) + "." + text.Reset)
	}
}

// hitResult is what resolveHits reports back: how many swings landed, the
// damage dealt, how much wear the weapon took (10 on a critical), and whether
// the round landed a critical strike.
type hitResult struct {
	hits         int
	totalDamage  int
	weaponDamage int
	critical     bool
}

// chiHits is the number of hits' worth of chi the round builds: a critical
// strike counts config.CritChiMultiplier times.
func (r hitResult) chiHits() int {
	if r.critical {
		return r.hits + config.CritChiMultiplier - 1
	}
	return r.hits
}

// resolveHits runs the swings of one attack round against the mob and reports
// the outcome. metric names the combat metric family ("kill", "leap", ...).
//
// It resolves in two passes. Every swing rolls to hit first, with the first
// swing at the normal miss chance and follow-ups at a penalised one (a monk's
// extra swings are always a flurry, which has its own penalty). The damage
// multipliers are then handed out to the landed hits in order, so whichever
// swing connects first always takes the full damage slot and any critical or
// double roll rides on it. Damage and reflection are totalled and reported once.
func resolveHits(s *state, whatMob *objects.Mob, attacks []float64, skillLevel int, metric string) hitResult {
	result := hitResult{weaponDamage: 1}
	baseMiss := DetermineMissChance(s, whatMob.Level-s.actor.Tier)
	followUpMiss := baseMiss + followUpMissPenalty(s, skillLevel)
	if followUpMiss > 95 {
		followUpMiss = 95
	}
	for swing := range attacks {
		missChance := baseMiss
		if swing > 0 {
			missChance = followUpMiss
		}
		if utils.Roll(100, 1, 0) <= missChance {
			data.StoreCombatMetric(metric+"-miss", 0, 0, 0, 0, 0, 0, s.actor.CharId, s.actor.Tier, 1, whatMob.MobId)
			whatMob.AddThreatDamage(1, s.actor)
			continue
		}
		result.hits++
	}
	if result.hits == 0 {
		s.msg.Actor.SendBad("You missed!!")
		return result
	}

	alwaysCrit := false
	if s.actor.Class != config.MONK {
		alwaysCrit = s.actor.Equipment.Main.Flags["always_crit"]
	}
	crushing := false
	totalReflect := 0
	for hit := 0; hit < result.hits; hit++ {
		mult := attacks[hit]
		action := metric
		if hit == 0 {
			if canCrush(s) && config.RollCrushing(skillLevel) {
				crushing = true
				mult *= config.CombatModifiers["crushing"]
				s.msg.Actor.SendGood("Craaackk!! A crushing blow!")
				action = metric + "-crushing"
			} else if config.RollCritical(skillLevel) || alwaysCrit {
				mult *= config.CombatModifiers["critical"]
				s.msg.Actor.SendGood("Critical Strike!")
				result.weaponDamage = 10
				result.critical = true
				action = metric + "-critical"
			} else if config.RollDouble(skillLevel) {
				mult *= config.CombatModifiers["double"]
				s.msg.Actor.SendGood("Double Damage!")
				action = metric + "-double"
			}
		}

		actualDamage, _, resisted := whatMob.ReceiveDamage(int(math.Ceil(float64(s.actor.InflictDamage()) * mult)))
		data.StoreCombatMetric(action, 0, 0, actualDamage+resisted, resisted, actualDamage, 0, s.actor.CharId, s.actor.Tier, 1, whatMob.MobId)
		whatMob.AddThreatDamage(actualDamage, s.actor)
		s.actor.AdvanceSkillExp((float64(actualDamage) / float64(whatMob.Stam.Max) * float64(whatMob.Experience)))
		result.totalDamage += actualDamage
		if whatMob.CheckFlag("reflection") {
			reflectDamage := int(float64(actualDamage) * config.ReflectDamageFromMob)
			stamDamage, vitDamage, resisted := s.actor.ReceiveDamage(reflectDamage)
			data.StoreCombatMetric(metric+"_mob_reflect", 0, 0, stamDamage+vitDamage+resisted, resisted, stamDamage+vitDamage, 1, whatMob.MobId, whatMob.Level, 0, s.actor.CharId)
			totalReflect += reflectDamage
		}
	}
	if result.hits == 1 {
		s.msg.Actor.SendInfo("You hit the " + whatMob.Name + " for " + strconv.Itoa(result.totalDamage) + " damage!" + text.Reset)
	} else {
		s.msg.Actor.SendInfo("You hit the " + whatMob.Name + " " + strconv.Itoa(result.hits) + " times for " + strconv.Itoa(result.totalDamage) + " damage!" + text.Reset)
	}
	if crushing {
		whatMob.Stun(config.CrushingStuns)
		s.msg.Observers.SendInfo(s.actor.Name + " lands a crushing blow on " + whatMob.Name + "!")
	}
	if totalReflect > 0 {
		s.msg.Actor.Send("The " + whatMob.Name + " reflects " + strconv.Itoa(totalReflect) + " damage back at you!")
		s.actor.DeathCheck(" was killed by reflection!")
	}
	warnTouchVulnerable(s, whatMob)
	return result
}

// followUpMissPenalty is the extra miss chance on every swing after the first.
func followUpMissPenalty(s *state, skillLevel int) int {
	if s.actor.Class == config.MONK {
		return config.FlurryMissPenaltyFor(skillLevel)
	}
	return config.MultiAttackMissPenaltyFor(skillLevel)
}

// DeathCheck Universal death check for mobs on whatever the current state is
func DeathCheck(s *state, m *objects.Mob) {
	if m.Stam.Current <= 0 {
		s.msg.Actor.SendGood("You killed " + m.Name)
		s.msg.Observers.SendGood(s.actor.Name + " killed " + m.Name)
		bloodlust(s, m.Name)
		partyLead := s.actor
		if s.actor.PartyFollow != "" {
			partyLead = objects.ActiveCharacters.Find(s.actor.PartyFollow)
		}
		partyMembers := append(partyLead.PartyFollowers, partyLead.Name)
		highestTier := 0
		expReduce := len(s.where.Chars.Contents)
		//reusing this because we are already running through everyone in the room
		for _, gm := range s.where.Chars.Contents {
			if gm.Tier > highestTier && gm.Class <= 8 {
				highestTier = gm.Tier
			}
			if gm.Class > 8 {
				expReduce -= 1
			}
		}

		if expReduce > 5 {
			expReduce = 5
		}
		//debuging stuff
		//s.msg.Actor.SendGood("Highest Tier: " + strconv.Itoa(highestTier))
		//s.msg.Actor.SendGood(strconv.Itoa(tierLimit))
		experienceAwarded := 0
		if config.QuestMode {
			experienceAwarded = m.Experience
		} else if m.CheckFlag("hostile") {
			experienceAwarded = int(float64(m.Experience) * (config.ExperienceReduction[expReduce] + (float64(utils.Roll(10, 1, 0)) / 100)))
		} else {
			experienceAwarded = m.Experience / 10
		}

		for flag, modifier := range config.XP_Modifiers {
			if m.CheckFlag(flag) {
				experienceAwarded += int(float64(m.Experience) * modifier)
			}
		}

		// Check for elemental damage in the room and add bonus experience
		if s.where.Flags["earth"] || s.where.Flags["fire"] || s.where.Flags["water"] || s.where.Flags["air"] {
			experienceAwarded = int(float64(experienceAwarded) * 1.06) // 6% bonus for elemental rooms
		}

		for _, member := range s.where.Chars.Contents {
			buildActorString := ""
			charClean := s.where.Chars.SearchAll(member.Name)
			if charClean != nil {
				partyCheck := false
				if config.QuestMode == false {
					for _, name := range partyMembers {
						if charClean.Name == name {
							partyCheck = true
						}
					}
				}
				if config.QuestMode {
					buildActorString += text.Cyan + "You earn " + strconv.Itoa(experienceAwarded) + " experience for the defeat of the " + m.Name + "\n"
					charClean.GainExperience(experienceAwarded)
				} else if partyCheck || m.CheckThreatTable(charClean.Name) {
					if int(math.Ceil((float64(charClean.Tier+1))*1.2)) < highestTier {
						buildActorString += text.Cyan + "You learn nothing for the defeat of the " + m.Name + "\n"
					} else {
						buildActorString += text.Cyan + "You earn " + strconv.Itoa(experienceAwarded) + " experience for the defeat of the " + m.Name + "\n"
						charClean.GainExperience(experienceAwarded)
					}
				}
				if charClean == s.actor {
					buildActorString += text.Green + m.DropInventory() + "\n"
					s.msg.Actor.Send(buildActorString)
				} else {
					go func() {
						if _, err := charClean.Write([]byte(buildActorString + "\n" + text.Reset)); err != nil {
							log.Println("Error writing to player: ", err)
						}
					}()
				}
				if charClean.Victim == m {
					charClean.Victim = nil
				}
			}
		}

		s.where.Mobs.Remove(m)
	}
}

// DetermineMissChance is the miss chance of a weapon attack: the weapon skill
// sets the base and the shared modifiers do the rest.
func DetermineMissChance(s *state, lvlDiff int) int {
	skill := 5
	if s.actor.Class != config.MONK {
		skill = s.actor.Equipment.Main.ItemType
	}
	return missChance(s, config.WeaponMissChance(s.actor.Skills[skill].Value), config.MissPerLevel, lvlDiff)
}

// BackstabMissChance is the miss chance of a backstab: the stealth skill sets
// the base and the shared modifiers do the rest.
func BackstabMissChance(s *state, lvlDiff int) int {
	stealth := config.StealthLevel(s.actor.Skills[11].Value)
	return missChance(s, config.BackstabMissChance(stealth), config.BackstabMissPerLevel, lvlDiff)
}

// missChance builds a to-hit roll's miss chance from a skill-derived base.
// Only the base and the per-level scalar differ between kinds of attack;
// the level penalty (from two levels up, off in quest mode), dex, combat
// flags such as bless and reckless, and the 5-95 clamp are shared.
func missChance(s *state, baseMiss int, missPerLevel int, lvlDiff int) int {
	missChance := baseMiss
	if !config.QuestMode && lvlDiff >= 2 {
		missChance += lvlDiff * missPerLevel
	}
	missChance -= s.actor.GetStat("dex") * config.HitPerDex
	if s.actor.CheckFlag("reckless") {
		missChance -= config.RecklessMissReduction
	}
	// The seal of justice carries the to-hit half of bless for paladins, who
	// cannot cast bless themselves. The two do not stack.
	if s.actor.CheckFlag("bless") || s.actor.CheckFlag("seal-justice") {
		missChance -= config.BlessHitBonus
	}
	if missChance >= 100 {
		missChance = 95
	}
	if missChance < 5 {
		missChance = 5
	}
	return missChance
}
