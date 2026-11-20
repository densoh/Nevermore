package cmd

import (
	"strings"
	"time"

	"github.com/ArcCS/Nevermore/config"
	"github.com/ArcCS/Nevermore/permissions"
	"github.com/ArcCS/Nevermore/utils"
)

func init() {
	addHandler(devious{},
		"Usage:  devious <strike> \n\n"+
			"Ready a devious strike. It fires in place of the sneak attack on your\n"+
			"next hit that would sneak attack, on whichever enemy that is. Only one\n"+
			"strike can be readied at a time; readying another replaces it. It is\n"+
			"let go once the fight is over.\n\n"+
			"  devious             show your readied strike and the ones you know\n"+
			"  devious <strike>    ready a strike\n"+
			"  devious clear       let your readied strike go\n\n"+
			"Devious strikes are learned from scrolls.\n\n"+
			"Strikes:\n"+
			"  poison           poison seeps into the wound and keeps hurting it\n"+
			"  expose-weakness  open a hole in the enemy's armor for everyone to\n"+
			"                   strike at, until it dies; it doesn't stack",
		permissions.Thief,
		"devious")
}

type devious cmd

func (devious) process(s *state) {
	if !s.requireTier(config.SneakAttackTier) {
		return
	}

	if len(s.words) < 1 {
		readied := "nothing"
		if s.actor.DeviousStrike != "" {
			readied = s.actor.DeviousStrike
		}
		lines := []string{"Readied devious strike: " + readied, "Devious strikes you know:"}
		known := 0
		for _, name := range config.DeviousStrikeOrder {
			if knowsDevious(s, name) {
				lines = append(lines, "  "+name)
				known++
			}
		}
		if known == 0 {
			lines = append(lines, "  none yet; they are learned from scrolls")
		}
		s.msg.Actor.SendInfo(strings.Join(lines, "\n"))
		return
	}

	// Readying isn't a combat action: only its own short anti-spam cooldown
	// applies, not stun, the global timer or haste.
	if left := s.actor.TimerRemaining("devious"); left > 0 {
		s.msg.Actor.SendBad("You're still settling on your last devious strike.")
		return
	}

	word := strings.ToLower(s.words[0])
	if word == "clear" || word == "none" || word == "cancel" {
		if s.actor.DeviousStrike == "" {
			s.msg.Actor.SendInfo("You have no devious strike readied.")
			return
		}
		s.actor.DeviousStrike = ""
		startDeviousCooldown(s)
		s.msg.Actor.SendInfo("You let your devious strike go.")
		return
	}

	strike := matchDevious(word)
	if strike == "" {
		s.msg.Actor.SendBad("You don't know a devious strike called " + word + ".")
		return
	}
	if !knowsDevious(s, strike) {
		s.msg.Actor.SendBad("You haven't learned " + strike + ". Devious strikes are learned from scrolls.")
		return
	}
	if !s.requireTier(config.DeviousStrikeTiers[strike]) {
		return
	}
	if s.actor.DeviousStrike == strike {
		s.msg.Actor.SendInfo("You already have " + strike + " readied.")
		return
	}
	previous := s.actor.DeviousStrike
	s.actor.DeviousStrike = strike
	startDeviousCooldown(s)
	if previous != "" {
		s.msg.Actor.SendGood("You ready " + strike + " in place of " + previous + ".")
	} else {
		s.msg.Actor.SendGood("You ready " + strike + ".")
	}
}

// matchDevious finds the strike a word names: the full name, or a prefix
// only one strike starts with ("expose", "poi").
func matchDevious(word string) string {
	if _, ok := config.DeviousStrikeTiers[word]; ok {
		return word
	}
	found := ""
	for _, name := range config.DeviousStrikeOrder {
		if strings.HasPrefix(name, word) {
			if found != "" {
				return ""
			}
			found = name
		}
	}
	return found
}

// knowsDevious reports whether the actor has learned the strike from a scroll.
func knowsDevious(s *state, strike string) bool {
	return utils.StringIn(config.DeviousSpell(strike), s.actor.Spells)
}

// startDeviousCooldown starts the fixed cooldown between changes of readied
// strike.
func startDeviousCooldown(s *state) {
	if s.actor.Timers == nil {
		s.actor.Timers = map[string]time.Time{}
	}
	s.actor.Timers["devious"] = time.Now().Add(config.DeviousCooldown * time.Second)
}
