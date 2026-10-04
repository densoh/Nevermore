// Copyright 2017 Andrew 'Diddymus' Rolfe. All rights reserved.
//
// Use of this source code is governed by the license in the LICENSE file
// included with the source code.

package cmd

import (
	"log"
	"strings"
	"time"

	"github.com/ArcCS/Nevermore/config"
	"github.com/ArcCS/Nevermore/objects"
	"github.com/ArcCS/Nevermore/permissions"
	menu "github.com/ArcCS/Nevermore/prompt"
	"github.com/ArcCS/Nevermore/utils"
)

// handler is the interface for command processing handlers.
type handler interface {
	process(*state)
}

type helpTextStruct struct {
	helpText string
	aliases  string
}

// handlers is a list of commands and their handlers. addHandler should be used
// to add new handlers. dispatchHandler then uses this list to look up the
// correct handler to invoke for a given command.
var handlers = map[string]handler{}
var handlerPermission = map[string]permissions.Permissions{}
var helpText = map[string]helpTextStruct{}
var oocCommands = []string{"SAY", "QUIT", "HELP", "WHO", "LOOK", "IC", "$POOF", "$DEATH", "AFK", "GO", "ACT"}
var excludeFromLogs = []string{"SAYTO", "SAY", "TELL", "OSAY", "SEND", "R", "REPLY", "REP", "PARTYTELL", "PTELL", "K", "KILL"}
var reverseLookup = map[string]string{}

// emotes is filled by the act handler from its emote table.
var emotes []string

// addHandler adds the given commands for the specified handler.
// It requires the command handler,  a help string to add to the help data, a bitmask permission, and the relative
// commands that will be each added to dispatch
func addHandler(h handler, helpString string, permission permissions.Permissions, cmds ...string) {
	primeString := ""
	if len(cmds) != 0 {
		primeString = strings.ToUpper(cmds[0])
	}
	if helpString != "" {
		helpText[strings.ToUpper(cmds[0])] = helpTextStruct{helpString, strings.Join(cmds[0:], ", ")}
	}
	for _, cmd := range cmds {
		if cmd != primeString {
			reverseLookup[strings.ToUpper(cmd)] = primeString
		}
		handlers[strings.ToUpper(cmd)] = h
		handlerPermission[strings.ToUpper(cmd)] = permission
	}
}

// runTyped runs the script in cmds that matches what the player typed and
// reports whether there was one. A trigger for the whole line wins over one
// for just the command word, which gets the rest of the line as arguments.
func (s *state) runTyped(cmds map[string]menu.MenuItem, completeCommand string, actorOnly bool) bool {
	if val, ok := cmds[completeCommand]; ok {
		s.runTrigger(actorOnly, val.Command)
		return true
	}
	if val, ok := cmds[s.cmd]; ok {
		s.runTrigger(actorOnly, val.Command, s.original)
		return true
	}
	return false
}

// dispatch handler takes the command sent and attempts to find it in a stack of command locations for execution
func dispatchHandler(s *state) {

	//log.Println("Last Activity was: " + strconv.Itoa(int(time.Now().Sub(objects.GetLastActivity(s.actor.Name)).Seconds())))

	if len(s.cmd) > 0 {

		if !utils.StringIn(strings.ToUpper(s.cmd), emotes) && !utils.StringIn(strings.ToUpper(s.cmd), excludeFromLogs) {
			log.Println(s.actor.Name + " sent " + s.cmd + " " + strings.Join(s.input, " "))
		}

		if !s.scripting {
			objects.LastActivity[s.actor.Name] = time.Now()
		}

		// $DEATH runs on its own goroutine, so anything the player already had
		// queued up would otherwise execute alongside it and fight over their
		// location - a queued movement completing after the death moves them
		// back out of the healing hand. Drop their input until the death script
		// has finished relocating them.
		if s.actor.DeathInProgress && !s.scripting && s.cmd != "QUIT" {
			s.msg.Actor.SendBad("You are in no condition to do that.")
			return
		}

		if s.where.RoomId == config.OocRoom &&
			!s.actor.Permission.HasAnyFlags(permissions.Dungeonmaster, permissions.Gamemaster) &&
			!utils.StringIn(strings.ToUpper(s.cmd), oocCommands) {
			s.msg.Actor.SendBad("You must be IC to do that.")
			return
		}

		if s.cmd[0] == '$' && !s.scripting {
			s.msg.Actor.SendBad("Unknown command, type HELP to get a list of commands (2)")
			return
		}

		// Scripts attached to things go before the built in commands, most
		// personal first: the player's own temporary commands (confirm
		// prompts, paging), what they carry, what they wear, the room, then
		// permanent items and mobs at their placement.
		completeCommand := s.cmd + " " + strings.Join(s.input, " ")
		if s.runTyped(s.actor.Commands, completeCommand, true) {
			return
		}
		s.actor.EmptyCommands()

		for _, i := range s.actor.Inventory.Contents {
			if s.runTyped(i.Commands, completeCommand, false) {
				return
			}
		}

		for _, i := range s.actor.Equipment.List() {
			if s.runTyped(i.Commands, completeCommand, false) {
				return
			}
		}

		if s.runTyped(s.where.Commands, completeCommand, false) {
			return
		}

		for _, i := range s.where.Items.Contents {
			if i.Flags["permanent"] && i.Placement == s.actor.Placement {
				if s.runTyped(i.Commands, completeCommand, false) {
					return
				}
			}
		}

		for _, i := range s.where.Mobs.Contents {
			if i.Flags["permanent"] && i.Placement == s.actor.Placement {
				if s.runTyped(i.Commands, completeCommand, false) {
					return
				}
			}
		}

		if len(s.cmd) > 1 {
			filtered_values := []string{}
			h_keys := []string{}
			for key, _ := range handlers {
				h_keys = append(h_keys, key)
			}
			for _, value := range h_keys {

				if strings.HasPrefix(value, s.cmd) {
					if len(filtered_values) == 0 || handlers[filtered_values[0]] != handlers[value] {
						filtered_values = append(filtered_values, value)
					}
				}
			}
			if len(filtered_values) == 1 {
				s.cmd = filtered_values[0]
			}
		}
		switch handler, valid := handlers[s.cmd]; {
		case valid:
			if s.actor.Permission.HasFlag(handlerPermission[s.cmd]) || s.actor.Permission.HasAnyFlags(permissions.Dungeonmaster, permissions.Gamemaster) {
				handler.process(s)
			} else {
				s.msg.Actor.SendInfo("Unknown command, type HELP to get a list of commands")
			}
		default:
			s.msg.Actor.SendBad("Unknown command, type HELP to get a list of commands (3)")
		}

	} else {
		s.msg.Actor.SendBad("Unknown command, type HELP to get a list of commands (4)")
	}
}
