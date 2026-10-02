package cmd

import (
	"regexp"
	"sort"
	"strings"

	"github.com/ArcCS/Nevermore/data"
	"github.com/ArcCS/Nevermore/permissions"
)

func init() {
	cmds := []string{"act", "emote", "me"}
	names := make([]string, 0, len(emoteDict))
	for name := range emoteDict {
		names = append(names, name)
	}
	sort.Strings(names)
	cmds = append(cmds, names...)

	addHandler(act{},
		"Usage:  act performs for all to see \n \n Perform actions.  Some emotes can also be aimed at someone: hug <name>",
		permissions.Player,
		cmds...)

	for _, c := range cmds {
		emotes = append(emotes, strings.ToUpper(c))
	}
}

// emote holds what an emote shows the room. solo follows the actor's name
// when there is no target; with goes between the actor's name and the
// target's. An emote missing solo needs a target, one missing with ignores it.
type emote struct {
	solo string
	with string
}

var emoteDict = map[string]emote{
	"beam":         {solo: "beams happily."},
	"blink":        {solo: "blinks slowly."},
	"blush":        {solo: "blushes."},
	"bounce":       {solo: "bounces around."},
	"bow":          {solo: "bows respectfully.", with: "bows before"},
	"burp":         {solo: "burps."},
	"cackle":       {solo: "cackles with insane glee!"},
	"cheer":        {solo: "cheers!"},
	"chuckle":      {solo: "chuckles politely."},
	"clap":         {solo: "claps enthusiastically."},
	"confused":     {solo: "looks very confused."},
	"cough":        {solo: "coughs."},
	"cringe":       {solo: "cringes."},
	"crossarms":    {solo: "crosses their arms."},
	"crossfingers": {solo: "crosses their fingers."},
	"cry":          {solo: "cries."},
	"curtsy":       {solo: "curtsies.", with: "curtsies to"},
	"dance":        {solo: "dances around."},
	"eyeroll":      {solo: "rolls their eyes."},
	"facepalm":     {solo: "buries their face in their palm."},
	"flex":         {solo: "flexes their muscles."},
	"flinch":       {solo: "flinches."},
	"frown":        {solo: "frowns."},
	"gasp":         {solo: "gasps."},
	"giggle":       {solo: "giggles."},
	"glare":        {solo: "glares.", with: "glares at"},
	"grin":         {solo: "grins.", with: "grins at"},
	"groan":        {solo: "groans."},
	"growl":        {solo: "growls."},
	"grumble":      {solo: "grumbles."},
	"grunt":        {solo: "grunts."},
	"hiccup":       {solo: "hiccups."},
	"hug":          {with: "hugs"},
	"hum":          {solo: "hums a tune."},
	"jump":         {solo: "jumps up and down."},
	"kick":         {with: "kicks"},
	"kneel":        {solo: "kneels down."},
	"laugh":        {solo: "laughs.", with: "laughs at"},
	"mutter":       {solo: "mutters under their breath."},
	"nod":          {solo: "nods.", with: "nods at"},
	"nudge":        {with: "nudges"},
	"pat":          {with: "pats"},
	"point":        {solo: "points.", with: "points at"},
	"poke":         {with: "pokes"},
	"ponder":       {solo: "ponders the situation."},
	"pout":         {solo: "pouts."},
	"salute":       {solo: "salutes."},
	"scowl":        {solo: "scowls."},
	"scream":       {solo: "screams!"},
	"shake":        {solo: "shakes their head.", with: "shakes hands with"},
	"shiver":       {solo: "shivers."},
	"shrug":        {solo: "shrugs."},
	"sigh":         {solo: "sighs."},
	"slap":         {with: "slaps"},
	"smile":        {solo: "smiles.", with: "smiles at"},
	"smirk":        {solo: "smirks."},
	"snap":         {solo: "snaps their fingers."},
	"sneeze":       {solo: "sneezes, ACHOOO!"},
	"snicker":      {solo: "snickers."},
	"sniff":        {solo: "sniffs."},
	"snort":        {solo: "snorts."},
	"spit":         {solo: "spits."},
	"stare":        {solo: "stares off into space.", with: "stares at"},
	"stretch":      {solo: "stretches."},
	"sulk":         {solo: "sulks."},
	"tap":          {solo: "taps their foot impatiently."},
	"thumbsdown":   {solo: "gives a thumbs down."},
	"thumbsup":     {solo: "gives a thumbs up."},
	"tickle":       {with: "tickles"},
	"wave":         {solo: "waves.", with: "waves at"},
	"whistle":      {solo: "whistles."},
	"wink":         {solo: "winks.", with: "winks at"},
	"yawn":         {solo: "yawns."},
}

type act cmd

func (act) process(s *state) {
	cmdStr := strings.ToLower(s.cmd)
	s.actor.RunHook("act")
	if cmdStr == "act" || cmdStr == "emote" || cmdStr == "me" {
		// Did they send an action?
		if len(s.words) == 0 {
			s.msg.Actor.SendBad("... what were you trying to do???")
			return
		}
		action := strings.Join(s.input, " ")
		match, _ := regexp.MatchString("([?.,\"'()!;:])", action[len(action)-1:])
		if !match {
			action = action + "."
		}
		data.StoreChatLog(3, s.actor.CharId, 0, action)
		s.msg.Actor.SendInfo(s.actor.Name, " ", action)
		s.msg.Observers.SendInfo(s.actor.Name, " ", action)
		s.ok = true
		return
	}

	e, ok := emoteDict[cmdStr]
	if !ok {
		s.msg.Actor.SendBad("Action not available")
		s.ok = true
		return
	}

	if e.with == "" || (len(s.words) == 0 && e.solo != "") {
		data.StoreChatLog(3, s.actor.CharId, 0, e.solo)
		s.msg.Actor.SendInfo(s.actor.Name, " ", e.solo)
		s.msg.Observers.SendInfo(s.actor.Name, " ", e.solo)
		s.ok = true
		return
	}

	whoWith := ""
	whoId := 0
	if len(s.words) > 0 {
		if targetPlayer := s.where.Chars.Search(s.words[0], s.actor); targetPlayer != nil {
			s.participant = targetPlayer
			whoWith = targetPlayer.Name
			whoId = targetPlayer.CharId
		} else if targetMob := s.where.Mobs.Search(s.words[0], 1, s.actor); targetMob != nil {
			whoWith = targetMob.Name
		}
	}
	if whoWith == "" {
		s.msg.Actor.SendBad("Who did you want to do that with?")
		s.ok = true
		return
	}

	data.StoreChatLog(3, s.actor.CharId, whoId, e.with)
	s.msg.Actor.SendInfo(s.actor.Name, " ", e.with, " ", whoWith, ".")
	if s.participant != nil {
		s.msg.Participant.SendInfo(s.actor.Name, " ", e.with, " you.")
	}
	s.msg.Observers.SendInfo(s.actor.Name, " ", e.with, " ", whoWith, ".")

	s.ok = true
}
