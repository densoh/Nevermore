package cmd

import (
	"slices"
	"testing"

	"github.com/ArcCS/Nevermore/config"
	"github.com/ArcCS/Nevermore/message"
	"github.com/ArcCS/Nevermore/objects"
)

func putItem(name string, itemType int, flags map[string]bool) *objects.Item {
	i := &objects.Item{ItemType: itemType, Flags: flags, Storage: &objects.ItemInventory{}, MaxUses: 10}
	i.Name = name
	return i
}

// boundPutState gives a fighter a ring and a bag (soulbound or not) and a
// permanent chest on the floor, then runs "put ring <where>".
func boundPutState(ring *objects.Item, bagFlags map[string]bool, where string) (*state, *objects.Item, *objects.Item) {
	bag := putItem("bag", 9, bagFlags)
	chest := putItem("chest", 9, map[string]bool{"permanent": true})
	s := attackerState(config.FIGHTER, nil)
	s.actor.Stam = objects.Meter{Max: 100, Current: 100}
	s.actor.Inventory = &objects.ItemInventory{Contents: []*objects.Item{ring, bag}}
	s.where = &objects.Room{Items: &objects.ItemInventory{Contents: []*objects.Item{chest}}}
	s.words = []string{"ring", where}
	s.input = s.words
	s.msg.Actor = &message.Buffer{}
	put{}.process(s)
	return s, bag, chest
}

func stored(container, item *objects.Item) bool {
	return slices.Contains(container.Storage.Contents, item)
}

// Soulbound items and bound rings go into a soulbound bag the owner carries.
func TestPutBoundIntoSoulboundBag(t *testing.T) {
	soulbound := putItem("ring", 24, map[string]bool{"permanent": true})
	if _, bag, _ := boundPutState(soulbound, map[string]bool{"permanent": true}, "bag"); !stored(bag, soulbound) {
		t.Error("soulbound item refused by a soulbound bag")
	}
	ring := putItem("ring", 24, map[string]bool{})
	ring.BoundTo = "someone"
	if _, bag, _ := boundPutState(ring, map[string]bool{"permanent": true}, "bag"); !stored(bag, ring) {
		t.Error("bound ring refused by a soulbound bag")
	}
}

// An ordinary bag or a permanent room fixture still refuses bound items.
func TestPutBoundElsewhereRefused(t *testing.T) {
	ring := putItem("ring", 24, map[string]bool{"permanent": true})
	if _, bag, _ := boundPutState(ring, map[string]bool{}, "bag"); stored(bag, ring) {
		t.Error("bound item went into an ordinary bag")
	}
	if _, _, chest := boundPutState(ring, map[string]bool{}, "chest"); stored(chest, ring) {
		t.Error("bound item went into a permanent room chest")
	}
}
