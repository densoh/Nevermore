package cmd

import (
	"github.com/ArcCS/Nevermore/data"
	"github.com/ArcCS/Nevermore/objects"
	"github.com/ArcCS/Nevermore/permissions"
	"github.com/ArcCS/Nevermore/utils"
	"strings"
)

func init() {
	addHandler(addCommand{},
		"Usage: addCommand room|mob|item (name) command_name command_string \n  Adds a command to the list of commands available \n"+
			"example:  addCommand mob dragon talk $TELEPORTTO room_id "+
			"$TELEPORTTO is from the list of script commands"+
			"The word talk is what will be processed as a command when the user, and the room_id is where the teleport will send them. "+
			"Use of the room option, will add the command to the current room",
		permissions.Builder,
		"addCommand")
}

type addCommand cmd

func (addCommand) process(s *state) {
	if len(s.words) < 3 {
		s.msg.Actor.SendBad("Not enough arguments to process the command.")
		return
	}

	// The script keeps the case it was typed in, so $ECHO and friends can
	// say things in mixed case. Verbs and triggers are matched uppercased.
	scriptFrom := func(at int) (string, bool) {
		script := strings.Join(s.input[at:], " ")
		if bad, ok := validScript(script); !ok {
			s.msg.Actor.SendBad("The inputted script was not recognized: " + bad)
			return "", false
		}
		return script, true
	}

	switch strings.ToLower(s.words[0]) {
	// Handle Rooms
	case "room":
		if script, ok := scriptFrom(2); ok {
			s.where.AddCommands(s.words[1], script)
			s.msg.Actor.SendGood("Script set on room")
			s.where.Save()
		}
		return
	// Handle Items
	case "item":
		if len(s.words) < 4 {
			s.msg.Actor.SendBad("There aren't enough arguments to complete this action.")
			return
		}
		item := s.actor.Inventory.Search(s.words[1], 1)
		if item == nil {
			s.msg.Actor.SendBad("Item not found.")
			return
		}
		if script, ok := scriptFrom(3); ok {
			item.AddCommands(s.words[2], script)
			s.msg.Actor.SendGood("Script set on item")
			saveScriptedItem(item)
		}
		return
	// Handle Mobs
	case "mob":
		if len(s.words) < 4 {
			s.msg.Actor.SendBad("There aren't enough arguments to complete this action.")
			return
		}
		mob := s.where.Mobs.Search(s.input[1], 1, s.actor)
		if mob == nil {
			s.msg.Actor.SendBad("Mob not found.")
			return
		}
		// Mob triggers run as the mob and take mob verbs; anything else
		// (typed triggers, @DEATH) runs as a player.
		if utils.StringIn(s.words[2], objects.MobTriggers) {
			script := strings.Join(s.input[3:], " ")
			if bad, ok := objects.ValidMobScript(script); !ok {
				s.msg.Actor.SendBad("The inputted mob script was not recognized: " + bad)
				return
			}
			mob.AddCommands(s.words[2], script)
			s.msg.Actor.SendGood("Script set on mob")
			saveScriptedMob(mob)
			return
		}
		if script, ok := scriptFrom(3); ok {
			mob.AddCommands(s.words[2], script)
			s.msg.Actor.SendGood("Script set on mob")
			saveScriptedMob(mob)
		}
		return
	default:
		s.msg.Actor.SendBad("Not an object that can be edited.")
	}

	s.ok = true
	return
}

// saveScriptedItem saves an item a builder just scripted and reloads its
// template, as build_edit does, so new copies carry the script too rather
// than only after a reboot.
func saveScriptedItem(item *objects.Item) {
	item.Save()
	objects.Items[item.ItemId], _ = objects.LoadItem(data.LoadItem(item.ItemId))
}

// saveScriptedMob is saveScriptedItem for mobs.
func saveScriptedMob(mob *objects.Mob) {
	mob.Save()
	objects.Mobs[mob.MobId], _ = objects.LoadMob(data.LoadMob(mob.MobId))
}
