package cmd

import (
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/ArcCS/Nevermore/config"
	"github.com/ArcCS/Nevermore/objects"
	"github.com/ArcCS/Nevermore/permissions"
	"github.com/ArcCS/Nevermore/text"
	"github.com/ArcCS/Nevermore/utils"
	"github.com/jinzhu/copier"
)

func init() {
	addHandler(scriptDeath{},
		"",
		permissions.Anyone,
		"$DEATH")
}

type scriptDeath cmd

// charInRoom reports whether this exact character is already listed in the
// room. CharInventory.Add appends unconditionally, so without this a character
// relocated twice ends up in the room's list twice.
func charInRoom(r *objects.Room, c *objects.Character) bool {
	for _, other := range r.Chars.Contents {
		if other == c {
			return true
		}
	}
	return false
}

func (scriptDeath) process(s *state) {

	healingHand := objects.Rooms[config.HealingHand]
	if !utils.IntIn(healingHand.RoomId, s.rLocks) {
		s.AddLocks(healingHand.RoomId)
		s.ok = false
		return
	}

	if time.Now().Sub(objects.GetLastActivity(s.actor.Name)).Seconds() < 60 {
		deathString := "### " + s.actor.Name + " has died."
		if len(s.words[0]) > 0 {
			deathString = "### " + s.actor.Name + " " + strings.Join(s.input[0:], " ")
		}

		objects.ActiveCharacters.MessageAll("### An otherworldly bell sounds once, the note echoing in your soul", config.BroadcastChannel)
		objects.ActiveCharacters.MessageAll(deathString, config.BroadcastChannel)

		harshdeath := false
		if s.actor.Tier > config.FreeDeathTier {

			// End the bards song before processing their death
			if s.actor.CheckFlag("singing") {
				s.actor.RemoveEffect("sing")
			}
			// A weapon prepared at their side goes back in the pack first, so that it
			// dies with the rest of their inventory rather than with their equipment.
			if stowed := s.actor.Equipment.Unprepare(); stowed != (*objects.Item)(nil) {
				s.actor.Inventory.Add(stowed)
			}
			equipment := s.actor.Equipment.UnequipAll()

			var tempStore []*objects.Item
			for _, item := range s.actor.Inventory.Contents {
				tempStore = append(tempStore, item)
			}

			newItem := objects.Item{}
			if err := copier.CopyWithOption(&newItem, objects.Items[1], copier.Option{DeepCopy: true}); err != nil {
				log.Println("Error copying item: ", err)
			}
			newItem.Name = "corpse of " + s.actor.Name
			newItem.Description = "It's the corpse of " + s.actor.Name + "."
			newItem.Placement = s.actor.Placement
			if len(tempStore) != 0 {
				for _, item := range tempStore {
					if !item.Flags["permanent"] {
						if err := s.actor.Inventory.Remove(item); err != nil {
							log.Println("Error removing item: ", err)
						}
						newItem.Storage.Add(item)
					}
				}
			}
			if len(equipment) != 0 {
				for _, item := range equipment {
					if !item.Flags["permanent"] {
						newItem.Storage.Add(item)
					}
				}
			}
			// Gold cost of death
			deathcost := 0
			deathcost = config.HealingHandCost[s.actor.Tier]
			if s.actor.Gold.CanSubtract(deathcost) {
				s.actor.Gold.SubIfCan(deathcost)
				s.msg.Actor.Send(text.Green + "The healing hand takes " + strconv.Itoa(deathcost) + " gold marks from your pockets as payment for the resurrection.\n\n" + text.Reset)
			} else {
				paid := 0
				paid = s.actor.Gold.Value
				s.actor.Gold.Value = 0
				rem := deathcost - paid
				if s.actor.BankGold.CanSubtract(rem) {
					s.actor.BankGold.SubIfCan(rem)
					s.msg.Actor.Send(text.Green + "The healing hand takes " + strconv.Itoa(paid) + " from your pockets and " + strconv.Itoa(rem) + " from your bank account as payment for the resurrection.\n\n" + text.Reset)
					paid += rem
				} else {
					s.msg.Actor.Send(text.Green + "You could not cover the cost of the healing hand and results in a more painful resurrection. (20% xp loss) \n\n" + text.Reset)
					log.Println(s.actor.Name + " has died with 20% loss due to insufficient funds.")
					harshdeath = true
				}
			}

			if s.actor.Gold.Value > 0 {
				newGold := objects.Item{}
				if err := copier.CopyWithOption(&newGold, objects.Items[3456], copier.Option{DeepCopy: true}); err != nil {
					log.Println("Error copying item: ", err)
				}
				newGold.Name = strconv.Itoa(s.actor.Gold.Value) + " gold marks"
				newGold.Value = s.actor.Gold.Value
				newItem.Storage.Add(&newGold)
				s.actor.Gold.Value = 0
			}

			s.msg.Observers.SendBad("The lifeless body of " + s.actor.Name + " falls to the ground.\n\n")
			s.where.Items.Add(&newItem)
		} else {
			s.msg.Actor.Send(text.Green + "An apprentice aura protects you from the worst of this death and ferries you and your gear safely to the healing hand...")
		}

		s.where.Chars.Remove(s.actor)
		if !charInRoom(healingHand, s.actor) {
			healingHand.Chars.Add(s.actor)
		}
		s.actor.Placement = 3
		s.actor.ParentId = healingHand.RoomId
		// The actor has moved, so where has to follow them - see the note on the
		// state struct. The LOOK below is dispatched through this same state and
		// would otherwise consult the command stack of the room they died in.
		s.where = healingHand

		s.actor.RemoveEffect("blind")
		s.actor.RemoveEffect("poison")
		s.actor.RemoveEffect("disease")
		s.actor.Stam.Current = s.actor.Stam.Max
		s.actor.Vit.Current = s.actor.Vit.Max
		s.actor.Mana.Current = s.actor.Mana.Max

		totalExpNeeded := config.MaxLoss(s.actor.Tier)
		finalMin := config.TierExpLevels[s.actor.Tier] - int(float64(totalExpNeeded))

		if config.QuestMode == true {
			finalMin = config.TierExpLevels[s.actor.Tier]
		}
		// Determine the death penalty
		if s.actor.Tier > config.FreeDeathTier {
			xpLoss := 0.10
			if harshdeath {
				xpLoss = 0.20
			} else {
				s.msg.Actor.Send(text.Green + "You've pass through this death with minimal effects. (10% xp loss) \n\n" + text.Reset)
				log.Println(s.actor.Name + " has died with 10% loss.")
			}
			s.actor.Experience.SubMax(int(float64(totalExpNeeded)*xpLoss), finalMin)
		} else {
			s.msg.Actor.Send(text.Green + "The healing hand is able to restore you completely and you suffer no experience loss.\n\n" + text.Reset)
		}

		s.actor.DeathInProgress = false
		s.scriptActor("LOOK")

	} else {
		deathString := "### " + s.actor.Name + " died a lag death."

		objects.ActiveCharacters.MessageAll("### An otherworldly bell attempts to ring but is abruptly muffled.", config.BroadcastChannel)
		objects.ActiveCharacters.MessageAll(deathString, config.BroadcastChannel)

		// lag death carries no cost or penalty

		// This has to happen inline: both rooms are locked by this state, and a
		// goroutine would run after they are released, shuffling the character
		// between rooms with nothing held and racing whatever ran next.
		log.Println("Lag Death: Clean Room")
		s.where.Chars.Remove(s.actor)
		if !charInRoom(healingHand, s.actor) {
			healingHand.Chars.Add(s.actor)
		}
		s.actor.RemoveEffect("blind")
		s.actor.RemoveEffect("poison")
		s.actor.RemoveEffect("disease")
		s.actor.Stam.Current = s.actor.Stam.Max
		s.actor.Vit.Current = s.actor.Vit.Max
		s.actor.Mana.Current = s.actor.Mana.Max
		s.actor.Placement = 3
		s.actor.ParentId = healingHand.RoomId
		s.where = healingHand

		s.actor.DeathInProgress = false
	}

}
