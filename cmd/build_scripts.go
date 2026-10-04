package cmd

import (
	"sort"
	"strings"

	"github.com/ArcCS/Nevermore/objects"
	"github.com/ArcCS/Nevermore/permissions"
)

func init() {
	addHandler(scripts{},
		"Usage:  scripts  \n Print all of the command scripts that can be attached to things. \n"+
			"Chain steps with ;  Attach under @PUT to run when an item is put into a container, \n"+
			"or under @DEATH on a mob to run as its killer when it dies.",
		permissions.Builder,
		"scripts")
}

type scripts cmd

var ScriptList = map[string]string{
	"$TELEPORTTO": "Usage: $TELEPORTO room_id, message; send player to a different room",
	"$TELEPORT":   "Usage: $TELEPORT message; invokes a randomized teleport with a message",
	"$REPAIR":     "Usage: $REPAIR, allows a player to repair weapons and armor",
	"$POOF":       "Usage: $POOF, announces an arrival of a player",
	"$BALANCE":    "Usage: $BALANCE, provides the player their bank information",
	"$DEPOSIT":    "Usage: $DEPOSIT, Deposits a value from primary gold to bank gold",
	"$WITHDRAW":   "Usage: $WITHDRAW, Withdraws gold from the bank and places it in the users gold pouch",
	"$BUY":        "Usage: $BUY item_name, will exchange a store list item for users gold",
	"$LIST":       "Usage: $LIST, lists all of the items for sale in the store",
	"$SELL":       "Usage: $SELL,  ALlows the user to pawn items.",
	"$ECHO":       "Usage: $ECHO, sends a message to the actor using the command",
	"$ECHOALL":    "Usage: $ECHOALL, sends a messae to the actor and everyone in the room",
	"$TEACH":      "Usage: $TEACH, teaches a spell to the actor",
	"$MELD":       "Usage: $MELD, melds 2 like items together",
	"$SOULBIND":   "Usage: $SOULBIND, binds an item to a player",
	"$RENAME":     "Usage: $RENAME, renames an item",
	"$SELLCHEST":  "Usage: $SELLCHEST, sells the contents of a chest/container",
	"$SLOT":       "Usage: $SLOT, adds slot machine",
	"$BAGLIST":    "Usage: $BAGLIST, list bag purchasing options",
	"$BUYBAG":     "Usage: $BUYBAG, buys a bag",
	"$MODBAG":     "Usage: $MODBAG, modifies a bag",
	"$SPLIT":      "Usage: $SPLIT, generates a split order for items",
	"$REQUIRE":    "Usage: $REQUIRE ITEM item_id [item_id...], event scripts: refuse unless the item is one of these",
	"$CONSUME":    "Usage: $CONSUME, event scripts: destroy the item; put it after any step that can refuse",
	"$EXITTO":     "Usage: $EXITTO room_id exit to_room_id seconds, the exit leads to another room for that long, then back (0 = put it back now)",
	"$IFSTAGE":    "Usage: $IFSTAGE event stage, stops the script unless the event is running and has reached stage",
	"$EVENTSTAGE": "Usage: $EVENTSTAGE event stage, advances a running event to stage (never back)",
	"$GIVEGOLD":   "Usage: $GIVEGOLD amount, gives the actor gold",
	"$GIVEITEM":   "Usage: $GIVEITEM item_id, gives the actor a copy of an item (dropped at their feet if too heavy)",
	"$WEAKEN":     "Usage: $WEAKEN room_id mob_id percent max_steps, permanently take percent% of a mob's armor/resists/hp/damage, up to max_steps times",
}

func (scripts) process(s *state) {
	for key, value := range ScriptList {
		s.msg.Actor.SendInfo(key + "| " + value + "\n")
	}
	s.msg.Actor.SendInfo("\nMob scripts, under " + strings.Join(objects.MobTriggers, " ") + " on a mob:")
	verbs := make([]string, 0, len(objects.MobVerbs))
	for verb := range objects.MobVerbs {
		verbs = append(verbs, verb)
	}
	sort.Strings(verbs)
	for _, verb := range verbs {
		s.msg.Actor.SendInfo(verb + "| " + objects.MobVerbs[verb].Usage + "\n")
	}
	s.msg.Actor.SendInfo("Targets: target attacker random all. Damage types: physical fire air earth water true.")

	s.ok = true
	return
}
