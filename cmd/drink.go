package cmd

import (
	"github.com/ArcCS/Nevermore/objects"
	"github.com/ArcCS/Nevermore/permissions"
	"log"
	"strconv"
	"strings"
)

func init() {
	addHandler(drink{},
		"Usage:  drink beverage # \n\n Take a sip from a beverage in your inventory or hands",
		permissions.Player,
		"drink", "quaff")
}

type drink cmd

func (drink) process(s *state) {
	if len(s.words) < 1 {
		s.msg.Actor.SendInfo("What do you want to drink?")
		return
	}

	ready, msg := s.actor.TimerReady("use")
	if !ready {
		s.msg.Actor.SendBad(msg)
		return
	}

	itemName := s.words[0]
	itemNum := 1
	if len(s.words) == 2 {
		if val, err := strconv.Atoi(s.words[1]); err == nil {
			itemNum = val
		}
	}

	what := s.actor.Inventory.Search(itemName, itemNum)
	if what == nil {
		what = s.actor.Equipment.Search(itemName, itemNum)
	}
	if what == nil {
		s.msg.Actor.SendInfo("Drink what?")
		return
	}
	if what.ItemType != 17 {
		s.msg.Actor.SendBad("You can't drink that.")
		return
	}
	if what.MaxUses <= 0 {
		s.msg.Actor.SendBad("The " + what.Name + " is empty.")
		return
	}

	castMsg := ""
	if what.Spell != "" {
		if objects.Rooms[s.actor.ParentId].Flags["no_magic"] {
			s.msg.Actor.SendBad("An oppressive anti-magic aura prevents you from drinking that here.")
			return
		}
		spellInstance, ok := objects.Spells[strings.ToLower(what.Spell)]
		if !ok {
			s.msg.Actor.SendBad("Spell doesn't exist in this world. ")
			return
		}
		castMsg = objects.Cast(s.actor, s.actor, spellInstance.Effect, spellInstance.Magnitude)
	}

	s.actor.RunHook("use")
	s.actor.SetTimer("use", 8)
	s.msg.Actor.SendGood("You take a sip of the " + what.Name + ".")
	s.msg.Observers.SendGood(s.actor.Name + " takes a sip of their " + what.Name + ".")
	if strings.Contains(castMsg, "$CRIPT") {
		go Script(s.actor, strings.Replace(castMsg, "$CRIPT ", "", 1))
	} else if castMsg != "" {
		s.msg.Actor.SendGood(castMsg)
	}

	objects.Effects["drunk"](s.actor, s.actor, what.Adjustment)

	what.MaxUses -= 1
	if what.MaxUses <= 0 {
		s.msg.Actor.SendInfo("You drain the last drop from the " + what.Name + " and toss it aside.")
		if s.actor.Equipment.Off == what {
			s.actor.Equipment.UnequipSpecific("off")
		} else if err := s.actor.Inventory.Remove(what); err != nil {
			s.msg.Actor.SendBad("Game Error when attempting to remove item from inventory.")
			log.Println("Error removing item from inventory: ", err)
		}
	}
	s.ok = true
}
