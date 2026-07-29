package cmd

import (
	"log"
	"math"
	"strconv"
	"strings"

	"github.com/ArcCS/Nevermore/config"
	"github.com/ArcCS/Nevermore/data"
	"github.com/ArcCS/Nevermore/objects"
	"github.com/ArcCS/Nevermore/permissions"
	"github.com/ArcCS/Nevermore/text"
	"github.com/ArcCS/Nevermore/utils"
)

func init() {
	addHandler(godir{},
		"Usage:  go direction # \n \n Proceed to the specified exit.   The cardinal directions can also be used without the use of go",
		permissions.Player,
		"GO", "N", "NE", "E", "SE", "S", "SW", "W", "NW", "U", "D",
		"NORTH", "NORTHEAST", "EAST", "SOUTHEAST",
		"SOUTH", "SOUTHWEST", "WEST", "NORTHWEST",
		"UP", "DOWN", "OUT", "O")
}

var (
	directionals = []string{"N", "NE", "E", "SE", "S", "SW", "W", "NW", "U", "D", "NORTH", "NORTHEAST",
		"EAST", "SOUTHEAST", "SOUTH", "SOUTHWEST", "WEST", "NORTHWEST", "UP", "DOWN", "OUT", "O"}

	directionIndex = map[string]string{
		"N":         "NORTH",
		"NORTH":     "NORTH",
		"NE":        "NORTHEAST",
		"NORTHEAST": "NORTHEAST",
		"E":         "EAST",
		"EAST":      "EAST",
		"SE":        "SOUTHEAST",
		"SOUTHEAST": "SOUTHEAST",
		"S":         "SOUTH",
		"SOUTH":     "SOUTH",
		"SW":        "SOUTHWEST",
		"SOUTHWEST": "SOUTHWEST",
		"W":         "WEST",
		"WEST":      "WEST",
		"NW":        "NORTHWEST",
		"NORTHWEST": "NORTHWEST",
		"U":         "UP",
		"UP":        "UP",
		"D":         "DOWN",
		"DOWN":      "DOWN",
		"OUT":       "OUT",
		"O":         "OUT",
	}
)

type godir cmd

func (godir) process(s *state) {

	var exitName string
	from := s.where
	// Does this place even have exits?
	if len(from.Exits) == 0 {
		s.msg.Actor.SendInfo("You can't see anywhere to go from here.")
		return
	}

	if s.actor.Stam.Current <= 0 {
		s.msg.Actor.SendBad("You are far too tired to do that.")
		return
	}

	// Decide what exit we are going to
	if utils.StringIn(s.cmd, directionals) {
		exitName = directionIndex[s.cmd]
	} else {
		if len(s.words) > 0 {
			// Join the strings together for exits with spaces
			exitName = strings.Join(s.words, " ")
		} else {
			s.msg.Actor.SendBad("Go where?")
		}
	}

	// Test for partial exit names
	exitTxt := strings.ToLower(exitName)
	if !utils.StringIn(strings.ToUpper(exitTxt), directionals) {
		for txtE := range from.Exits {
			if strings.Contains(txtE, exitTxt) {
				exitTxt = txtE
			}
		}
	}
	if toE, ok := from.Exits[exitTxt]; ok {
		s.actor.RunHook("move")
		// Check that the room ID exists
		if to, ok := objects.Rooms[toE.ToId]; ok {
			// Apply a lock
			if !utils.IntIn(toE.ToId, s.rLocks) {
				s.AddLocks(toE.ToId)
				s.ok = false
				return
			} else {
				if !s.actor.Permission.HasAnyFlags(permissions.Builder, permissions.Dungeonmaster, permissions.Gamemaster) {

					allowed, evasive, chasers := canMove(actorMover(s), from, to, toE)
					if !allowed {
						return
					}

					moveChar(s.actor, from, to, evasive)

					// Broadcast leaving and arrival notifications
					if s.actor.Flags["invisible"] == false {
						s.msg.Observers[from.RoomId].SendInfo("You see ", s.actor.Name, " go to the ", strings.ToLower(toE.Name), ".")
						s.msg.Observers[to.RoomId].SendInfo(s.actor.Name, " just arrived.")
					}

					// The actor is out of the room, so anything hunting them can
					// come along. This is done here, inline, because this state
					// already holds both rooms: handing the mob a command to chase
					// them later would have it move itself with nothing locked,
					// after the actor may have moved on or died. FollowChar makes
					// its own roll and refuses rooms mobs have no business in.
					// Only a mob that actually arrives gets a swing, and only the
					// first one to land it does - a chased character pays for the
					// exit once, however big the pack behind them.
					// MOB-CHASE-SWITCH: delete this loop to stop mobs following the actor between rooms.
					vitalSpent := false
					for _, mob := range chasers {
						if !mob.FollowChar(s.actor, from, to) {
							continue
						}
						if vitalSpent {
							continue
						}
						landed, died := followVital(actorMover(s), mob)
						vitalSpent = landed
						if died {
							// The death script has them now, so there is no one
							// left here to lead the party or be shown the room.
							s.ok = true
							return
						}
					}

					// Do not invoke player state, just move them within this state
					// lock - a state of their own would deadlock on the rooms this
					// one is holding. Each follower is judged on their own merits,
					// so one who cannot come along does not strand the rest.
					for _, peo := range s.actor.PartyFollowers {
						follChar := from.Chars.SearchAll(peo)
						if follChar == nil {
							continue
						}

						follChar.RunHook("move")

						follAllowed, follEvasive, follChasers := true, 0, []*objects.Mob(nil)
						if !follChar.Permission.HasAnyFlags(permissions.Builder, permissions.Dungeonmaster, permissions.Gamemaster) {
							follAllowed, follEvasive, follChasers = canMove(followerMover(follChar), from, to, toE)
						}
						if !follAllowed {
							continue
						}

						moveChar(follChar, from, to, follEvasive)

					evasiveMan := 0
					// Check if anyone blocks.
					for _, mob := range s.where.Mobs.Contents {
						// Check if a mob blocks.
						if _, inList := mob.ThreatTable[s.actor.Name]; inList {
							if mob.CheckFlag("block_exit") && mob.Placement == s.actor.Placement && mob.MobStunned == 0 && !mob.CheckFlag("run_away") {
								evasiveMan = 2
								curChance := config.MobBlock - ((s.actor.Tier - mob.Level) * config.MobBlockPerLevel)
								if curChance > 85 {
									curChance = 85
								}
								if utils.Roll(100, 1, 0) <= curChance {
									s.msg.Actor.SendBad(mob.Name + " blocks your way.")
									s.actor.SetTimer("global", 8)
									return
								}
								break
							}
						}
					}
					for _, mob := range s.where.Mobs.Contents {
						// No one blocked, so check if anyone follows.
						if _, inList := mob.ThreatTable[s.actor.Name]; inList {
							if mob.CurrentTarget == s.actor.Name {
								// Now check if they follow.
								if mob.CheckFlag("follows") && !mob.CheckFlag("curious_canticle") {
									evasiveMan = 4
									if utils.Roll(100, 1, 0) <= config.MobFollowVital {
										vitDamage, resisted := s.actor.ReceiveVitalDamage(int(math.Ceil(float64(mob.InflictDamage() * config.MobFollMult))))
										data.StoreCombatMetric("follow_vital", 0, 1, vitDamage, resisted, vitDamage, 1, mob.MobId, mob.Level, 0, s.actor.CharId)

										if vitDamage == 0 {
											s.msg.Actor.SendInfo(text.Red + mob.Name + " attacks bounces off of you for no damage!" + "\n" + text.Reset)
										} else {
											s.msg.Actor.SendBad(text.Red + "Vital Strike!!!!\n" + text.Reset)
											s.msg.Actor.SendBad(text.Red + mob.Name + " attacks you for " + strconv.Itoa(vitDamage) + " points of vital damage!" + "\n" + text.Reset)
										}
										deathCheck := s.actor.DeathCheckBool("was slain by a " + mob.Name + ".")
										if deathCheck {
											return
										}
										break
									}
								}
							}
						}

					/*
						// Character has been removed, invoke any follows for them.  this should be fine as the mob should take over locks
						for _, mob := range followList {
							mobProc := mob
							go func() { mobProc.MobCommands <- "follow " + s.actor.Name }()
						}
					*/

					// Do not invoke player state, just move them within this state lock
					if len(s.actor.PartyFollowers) > 0 {
						for _, peo := range s.actor.PartyFollowers {
							follChar := s.where.Chars.SearchAll(peo)
							endFollProc := false
							if follChar != nil {
								// Check some timers
								if !follChar.Permission.HasAnyFlags(permissions.Builder, permissions.Dungeonmaster, permissions.Gamemaster) {
									ready, msg := follChar.TimerReady("evade")
									if !ready {
										if _, err := follChar.Write([]byte(text.Bad + msg)); err != nil {
											log.Println("Error writing to player: ", err)
										}
										break
									}

									if s.actor.Stam.Current <= 0 {
										if _, err := follChar.Write([]byte(text.Bad + "You are far too tired to follow.")); err != nil {
											log.Println("Error writing to player: ", err)
										}
										break
									}

									follChar.RunHook("move")

									evasiveMan = 0

									if !objects.Rooms[toE.ToId].Flags["active"] {
										if _, err := follChar.Write([]byte(text.Bad + "Go where?")); err != nil {
											log.Println("Error writing to player: ", err)
										}
										break
									}

									if toE.Flags["invisible"] && !follChar.CheckFlag("detect-invisible") {
										if _, err := follChar.Write([]byte(text.Bad + "Go where?")); err != nil {
											log.Println("Error writing to player: ", err)
										}
										break
									}

									if toE.Flags["placement_dependent"] && follChar.Placement != toE.Placement {
										if _, err := follChar.Write([]byte(text.Bad + "You must be next to the exit to use it.")); err != nil {
											log.Println("Error writing to player: ", err)
										}
										break
									}

									if follChar.Equipment.GetWeight() > follChar.MaxWeight() {
										if _, err := follChar.Write([]byte(text.Bad + "You are carrying too much to move.")); err != nil {
											log.Println("Error writing to player: ", err)
										}
										break
									}

									if objects.Rooms[toE.ToId].Crowded() {
										if _, err := follChar.Write([]byte("That area is crowded.")); err != nil {
											log.Println("Error writing to player: ", err)
										}
										s.ok = true
										return
									}

									hasRope := false
									if follChar.Equipment.Off != (*objects.Item)(nil) {
										if follChar.Equipment.Off.ItemId == 1463 {
											hasRope = true
										}
									}

									if toE.Flags["levitate"] && !follChar.CheckFlag("levitate") && !hasRope {
										chanceToPass := follChar.GetStat("dex")/45 + 10
										if utils.Roll(100, 1, 0) >= chanceToPass {
											fallDamageStam := int(config.FallDamage*float64(follChar.Stam.Max)) -
												(config.ConFallDamageMod * follChar.GetStat("con")) -
												(config.DexFallDamageMod * follChar.GetStat("dex"))
											fallDamageVit := int(config.FallDamage*float64(follChar.Stam.Max)) -
												(config.ConFallDamageMod * follChar.GetStat("con")) -
												(config.DexFallDamageMod * follChar.GetStat("dex"))
											totStam, totVit := 0, 0
											if fallDamageStam > 0 {
												totStam, totVit = follChar.ReceiveDamageNoArmor(fallDamageStam)
											}
											if fallDamageVit > 0 {
												totVit += follChar.ReceiveVitalDamageNoArmor(fallDamageVit)
											}
											buildStr := ""
											if totStam <= 0 && totVit <= 0 {
												buildStr = "You take no damage in the fall."
											} else {
												if totStam >= 1 {
													buildStr += "You take " + strconv.Itoa(totStam) + " points of stamina"
												}
												if totVit >= 1 {
													if totStam >= 1 {
														buildStr += " and "
													}
													buildStr += strconv.Itoa(totVit) + " points of vitality"
												}
												buildStr += " damage in the fall."
											}
											if _, err := follChar.Write([]byte(text.Bad + "You fall while trying to go that way! " + buildStr)); err != nil {
												log.Println("Error writing to player: ", err)
											}
											go follChar.DeathCheck("fell to their death.")
											break
										}
									}

									// Check if anyone blocks.
									for _, mob := range s.where.Mobs.Contents {
										// Check if a mob blocks.
										if _, inList := mob.ThreatTable[follChar.Name]; inList {
											if mob.CheckFlag("block_exit") && mob.Placement == follChar.Placement && mob.MobStunned == 0 && !mob.CheckFlag("run_away") {
												evasiveMan = 2
												curChance := config.MobBlock - ((follChar.Tier - mob.Level) * config.MobBlockPerLevel)
												if curChance > 85 {
													curChance = 85
												}
												if utils.Roll(100, 1, 0) <= curChance {
													if _, err := follChar.Write([]byte(mob.Name + " blocks you from following." + "\n")); err != nil {
														log.Println("Error writing to player: ", err)
													}
													follChar.SetTimer("global", 8)

												}
												endFollProc = true
												break
											}
										}
									}
									for _, mob := range s.where.Mobs.Contents {
										// Check if a follows
										if _, inList := mob.ThreatTable[follChar.Name]; inList {
											if mob.CurrentTarget == follChar.Name {
												// Now check if they follow.
												if mob.CheckFlag("follows") && !mob.CheckFlag("curious_canticle") {
													evasiveMan = 4
													if utils.Roll(100, 1, 0) <= config.MobFollowVital {
														vitDamage, resisted := follChar.ReceiveVitalDamage(int(math.Ceil(float64(mob.InflictDamage() * config.MobFollMult))))
														data.StoreCombatMetric("follow_vital", 0, 1, vitDamage, resisted, vitDamage, 1, mob.MobId, mob.Level, 0, follChar.CharId)

														if vitDamage == 0 {
															if _, err := follChar.Write([]byte(text.Red + mob.Name + " attacks bounces off of you for no damage!" + "\n" + text.Reset)); err != nil {
																log.Println("Error writing to player: ", err)
															}

														} else {
															if _, err := follChar.Write([]byte(text.Red + "Vital Strike!!!!\n" + text.Reset)); err != nil {
																log.Println("Error writing to player: ", err)
															}
															if _, err := follChar.Write([]byte(text.Red + mob.Name + " attacks you for " + strconv.Itoa(vitDamage) + " points of vital damage!" + "\n" + text.Reset)); err != nil {
																log.Println("Error writing to player: ", err)
															}
														}
														deathCheck := s.actor.DeathCheckBool("was slain by a " + mob.Name + ".")
														if deathCheck {
															endFollProc = true
														}
														break
													}
												}
											}
										}
									}
									if endFollProc {
										continue
									}
								}
								from.Chars.Remove(follChar)
								// If they were evasive, add a global timer
								follChar.SetTimer("evade", evasiveMan)
								to.Chars.Add(follChar)
								follChar.Victim = nil
								follChar.Placement = 3
								follChar.ParentId = toE.ToId

								if s.actor.CheckFlag("blind") {
									s.msg.Actor.SendBad("You can't see anything!")
									return
								} else {
									if _, err := follChar.Write([]byte(objects.Rooms[to.RoomId].Look(follChar))); err != nil {
										log.Println("Error writing to player: ", err)
									}
								}

								// Broadcast leaving and arrival notifications
								if follChar.Flags["invisible"] == false {
									s.msg.Observers[from.RoomId].SendInfo("You see ", follChar.Name, " follow "+s.actor.Name+" to the ", strings.ToLower(toE.Name), ".")
									s.msg.Observers[to.RoomId].SendInfo(follChar.Name, " just arrived.")
								}
							}
						}
					}

					s.scriptActor("LOOK")
					s.ok = true
					return
				} else {
					// Builders and staff walk through everything unchecked.
					// SetTimer is a no-op for them, so the evade timer here
					// costs them nothing.
					moveChar(s.actor, from, to, 0)

					// Broadcast leaving and arrival notifications
					if s.actor.Flags["invisible"] == false {
						s.msg.Observers[from.RoomId].SendInfo("You see ", s.actor.Name, " go to the ", strings.ToLower(toE.Name), ".")
						s.msg.Observers[to.RoomId].SendInfo(s.actor.Name, " just arrived.")
					}

					s.scriptActor("LOOK")
					s.ok = true
					return
				}
			}
		} else {
			s.msg.Actor.SendInfo("You can't go that direction.")
			s.ok = true
			return
		}
	} else {
		s.msg.Actor.SendInfo("You can't go that direction.")
		s.ok = true
		return
	}
}

// mover is the character being walked through an exit, together with how to
// talk to them. The actor's output goes through the state's buffers so it
// arrives with the rest of the command; a party follower is not this state's
// actor, so theirs has to be written straight to their connection. Everything
// else about moving the two is identical, which is the point.
type mover struct {
	char *objects.Character
	bad  func(string)
	info func(string)
}

func actorMover(s *state) mover {
	return mover{
		char: s.actor,
		bad:  func(msg string) { s.msg.Actor.SendBad(msg) },
		info: func(msg string) { s.msg.Actor.SendInfo(msg) },
	}
}

func followerMover(char *objects.Character) mover {
	return mover{
		char: char,
		bad:  func(msg string) { writeTo(char, text.Bad+msg) },
		info: func(msg string) { writeTo(char, text.Info+msg) },
	}
}

// writeTo sends text straight to a character, for output that cannot go
// through the state's buffers because the character is not this state's actor.
func writeTo(char *objects.Character, msg string) {
	if _, err := char.Write([]byte(msg + "\n")); err != nil {
		log.Println("Error writing to player: ", err)
	}
}

// canMove runs every per-character check for walking through toE and reports
// whether the character may go, the evade timer their exit earned them, and
// the mobs entitled to give chase once they are gone. One of the checks bites:
// a failed levitate roll damages the character before refusing, and can kill
// them.
//
// Nothing here ends the command or touches the state. The caller decides what
// a refusal means - the actor stops walking, a follower is simply left behind -
// which is what lets the actor and their party share one copy of the rules.
// Both rooms must already be locked by the caller.
func canMove(m mover, from *objects.Room, to *objects.Room, toE *objects.Exit) (allowed bool, evasive int, chasers []*objects.Mob) {
	char := m.char

	// Check some timers
	if ready, msg := char.TimerReady("evade"); !ready {
		m.bad(msg)
		return false, 0, nil
	}

	if char.Stam.Current <= 0 {
		m.bad("You are far too tired to move.")
		return false, 0, nil
	}

	if !to.Flags["active"] {
		m.bad("Go where?")
		return false, 0, nil
	}

	if toE.Flags["invisible"] && !char.CheckFlag("detect-invisible") {
		m.bad("Go where?")
		return false, 0, nil
	}

	if toE.Flags["placement_dependent"] && char.Placement != toE.Placement {
		m.bad("You must be next to the exit to use it.")
		return false, 0, nil
	}

	if toE.Flags["closed"] {
		m.bad("The way is closed.")
		return false, 0, nil
	}

	if toE.Flags["day_only"] && !objects.DayTime {
		m.bad("You can only go there at night.")
		return false, 0, nil
	}

	if toE.Flags["night_only"] && objects.DayTime {
		m.bad("You can only go there during the day.")
		return false, 0, nil
	}

	if char.Equipment.GetWeight() > char.MaxWeight() {
		m.bad("You are carrying too much to move.")
		return false, 0, nil
	}

	// Asked before the fall roll below: no reason to drop someone down a shaft
	// they were never going to be let into.
	if to.Crowded() {
		m.info("That area is crowded.")
		return false, 0, nil
	}

	hasRope := false
	if char.Equipment.Off != (*objects.Item)(nil) {
		if char.Equipment.Off.ItemId == 1463 {
			hasRope = true
		}
	}
	if toE.Flags["levitate"] && !char.CheckFlag("levitate") && !hasRope {
		chanceToPass := char.GetStat("dex")/45 + 10
		if utils.Roll(100, 1, 0) >= chanceToPass {
			m.bad("You fall while trying to go that way! " + takeFallDamage(char))
			go char.DeathCheck("fell to their death.")
			return false, 0, nil
		}
	}

	// Check if anyone blocks.
	for _, mob := range from.Mobs.Contents {
		if _, inList := mob.ThreatTable[char.Name]; !inList {
			continue
		}
		if mob.CheckFlag("block_exit") && mob.Placement == char.Placement && mob.MobStunned == 0 && !mob.CheckFlag("run_away") {
			evasive = 2
			curChance := config.MobBlock - ((char.Tier - mob.Level) * config.MobBlockPerLevel)
			if curChance > 85 {
				curChance = 85
			}
			if utils.Roll(100, 1, 0) <= curChance {
				m.bad(mob.Name + " blocks your way.")
				char.SetTimer("global", 8)
				return false, 0, nil
			}
			break
		}
	}

	// No one blocked, so work out who is entitled to give chase. Only the
	// entitlement is decided here - whether each one actually comes along is
	// FollowChar's roll, and the parting strike hangs off that, not off this.
	for _, mob := range from.Mobs.Contents {
		if _, inList := mob.ThreatTable[char.Name]; !inList {
			continue
		}
		if mob.CurrentTarget != char.Name {
			continue
		}
		if !mob.CheckFlag("follows") || mob.CheckFlag("curious_canticle") {
			continue
		}

		evasive = 4
		chasers = append(chasers, mob)
	}

	return true, evasive, chasers
}

// followVital rolls the parting strike a mob earns for chasing a character
// through an exit, and reports whether it landed and whether it killed them.
//
// This is only ever called for a mob that has already followed. The strike is
// the mob catching up with them on the other side, so one that lost its follow
// roll - or was never in any shape to give chase - has nothing to swing at. It
// lands in the destination room, after the move, for the same reason.
func followVital(m mover, mob *objects.Mob) (landed bool, died bool) {
	char := m.char

	if utils.Roll(100, 1, 0) > config.MobFollowVital-(char.GetStat("dex")/2) {
		return false, false
	}

	vitDamage, resisted := char.ReceiveVitalDamage(int(math.Ceil(float64(mob.InflictDamage() * config.MobFollMult))))
	data.StoreCombatMetric("follow_vital", 0, 1, vitDamage, resisted, vitDamage, 1, mob.MobId, mob.Level, 0, char.CharId)

	if vitDamage == 0 {
		m.info(text.Red + mob.Name + " attacks bounces off of you for no damage!" + "\n" + text.Reset)
	} else {
		m.bad(text.Red + "Vital Strike!!!!\n" + text.Reset)
		m.bad(text.Red + mob.Name + " attacks you for " + strconv.Itoa(vitDamage) + " points of vital damage!" + "\n" + text.Reset)
	}

	// Whoever took the hit is the one to death check.
	return true, char.DeathCheckBool("was slain by a " + mob.Name + ".")
}

// takeFallDamage applies the damage for a failed levitate roll and returns the
// line describing it.
func takeFallDamage(char *objects.Character) string {
	fallDamageStam := int(config.FallDamage*float64(char.Stam.Max)) -
		(config.ConFallDamageMod * char.GetStat("con")) -
		(config.DexFallDamageMod * char.GetStat("dex"))
	fallDamageVit := int(config.FallDamage*float64(char.Stam.Max)) -
		(config.ConFallDamageMod * char.GetStat("con")) -
		(config.DexFallDamageMod * char.GetStat("dex"))

	totStam, totVit := 0, 0
	if fallDamageStam > 0 {
		totStam, totVit = char.ReceiveDamageNoArmor(fallDamageStam)
	}
	if fallDamageVit > 0 {
		totVit += char.ReceiveVitalDamageNoArmor(fallDamageVit)
	}

	if totStam <= 0 && totVit <= 0 {
		return "You take no damage in the fall."
	}

	buildStr := ""
	if totStam >= 1 {
		buildStr += "You take " + strconv.Itoa(totStam) + " points of stamina"
	}
	if totVit >= 1 {
		if totStam >= 1 {
			buildStr += " and "
		}
		buildStr += strconv.Itoa(totVit) + " points of vitality"
	}
	return buildStr + " damage in the fall."
}

// moveChar takes the character out of one room and puts them in the other,
// with the evade timer their exit earned. Both rooms must be locked by the
// caller.
func moveChar(char *objects.Character, from *objects.Room, to *objects.Room, evasive int) {
	from.Chars.Remove(char)
	// If they were evasive, add a global timer
	char.SetTimer("evade", evasive)
	to.Chars.Add(char)
	char.Victim = nil
	char.Placement = 3
	char.ParentId = to.RoomId
}
