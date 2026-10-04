package cmd

import (
	"log"
	"strconv"

	"github.com/ArcCS/Nevermore/objects"
	"github.com/ArcCS/Nevermore/permissions"
	"github.com/jinzhu/copier"
)

// Syntax: $GIVEITEM item_id
//
// Gives the actor a fresh copy of an item. If they can't carry it, it lands
// at their feet instead, so a reward is never lost.
func init() {
	addHandler(scriptGiveItem{},
		"",
		permissions.Anyone,
		"$GIVEITEM")
}

type scriptGiveItem cmd

func (scriptGiveItem) process(s *state) {
	if len(s.words) < 1 {
		log.Println("$GIVEITEM needs an item_id")
		return
	}
	itemId, err := strconv.Atoi(s.words[0])
	if err != nil {
		log.Println("$GIVEITEM item_id " + s.words[0] + " is not a number")
		return
	}
	template, ok := objects.Items[itemId]
	if !ok {
		log.Println("$GIVEITEM item " + s.words[0] + " does not exist")
		return
	}

	newItem := objects.Item{}
	if err := copier.CopyWithOption(&newItem, template, copier.Option{DeepCopy: true}); err != nil {
		log.Println("Error copying item: ", err)
		return
	}

	if s.actor.GetCurrentWeight()+newItem.GetWeight() > s.actor.MaxWeight() {
		newItem.Placement = s.actor.Placement
		s.where.Items.Add(&newItem)
		s.msg.Actor.SendInfo("You can't carry ", newItem.Name, ", so it is left at your feet.")
	} else {
		s.actor.Inventory.Add(&newItem)
		s.msg.Actor.SendGood("You receive ", newItem.Name, ".")
	}
	s.ok = true
}
