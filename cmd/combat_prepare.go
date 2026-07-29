package cmd

import (
	"log"
	"strconv"

	"github.com/ArcCS/Nevermore/config"
	"github.com/ArcCS/Nevermore/objects"
	"github.com/ArcCS/Nevermore/permissions"
	"github.com/ArcCS/Nevermore/utils"
)

func init() {
	addHandler(prepare{},
		"Usage:  prepare [item #] \n\n Ready a weapon from your pack at your side, so it can be quickdrawn.  Used with no item it puts the weapon you have ready back in your pack.",
		permissions.Ranger,
		"prepare", "prep")
}

type prepare cmd

func (prepare) process(s *state) {
	if s.actor.Tier < config.SpecialAbilityTier {
		s.msg.Actor.SendBad("You must be at least tier " + strconv.Itoa(config.SpecialAbilityTier) + " to use this skill.")
		return
	}

	// Nothing named, so this is a request to put away whatever is ready.
	if len(s.words) == 0 {
		if s.actor.Equipment.Prepared == (*objects.Item)(nil) {
			s.msg.Actor.SendBad("You have nothing prepared.")
			return
		}
		stowed := s.actor.Equipment.Unprepare()
		s.actor.Inventory.Add(stowed)
		s.msg.Actor.SendGood("You put a " + stowed.DisplayName() + " back in your pack.")
		s.msg.Observers.SendInfo(s.actor.Name + " puts a " + stowed.DisplayName() + " away.")
		s.ok = true
		return
	}

	name := s.input[0]
	nameNum := 1

	if len(s.words) > 1 {
		// Try to snag a number off the list
		if val, err := strconv.Atoi(s.words[1]); err == nil {
			nameNum = val
		}
	}

	what := s.actor.Inventory.Search(name, nameNum)
	if what == nil {
		s.msg.Actor.SendBad("You have no '" + name + "' to prepare.")
		return
	}

	if !utils.IntIn(what.ItemType, []int{0, 1, 2, 3, 4}) {
		s.msg.Actor.SendBad("You can only prepare a weapon.")
		return
	}

	if what.MaxUses <= 0 {
		s.msg.Actor.SendBad("The " + what.DisplayName() + " is broken.")
		return
	}

	if ok, msg := s.actor.CanEquip(what); !ok {
		s.msg.Actor.SendBad(msg)
		return
	}

	if err := s.actor.Inventory.Remove(what); err != nil {
		s.msg.Actor.SendBad("Failure to prepare item.")
		log.Println("Error removing item from inventory: ", err)
		return
	}

	// Preparing a second weapon replaces the first, which goes back in the pack.
	if released := s.actor.Equipment.Prepare(what); released != (*objects.Item)(nil) {
		s.actor.Inventory.Add(released)
		s.msg.Actor.SendInfo("You put a " + released.DisplayName() + " back in your pack.")
	}

	s.msg.Actor.SendGood("You have a " + what.DisplayName() + " ready at your side.")
	s.msg.Observers.SendInfo(s.actor.Name + " readies a " + what.DisplayName() + " at their side.")
	s.ok = true
}
