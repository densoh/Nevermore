package cmd

import "strconv"

// Shared precondition checks for commands. Each returns true when the actor
// may go ahead, and otherwise tells them why not and returns false, so a
// command can open with `if !s.requireSight() { return }`.

// requireSight fails while the actor is blind.
func (s *state) requireSight() bool {
	if s.actor.CheckFlag("blind") {
		s.msg.Actor.SendBad("You can't see anything!")
		return false
	}
	return true
}

// requireStamina fails once the actor's stamina is spent.
func (s *state) requireStamina() bool {
	if s.actor.Stam.Current <= 0 {
		s.msg.Actor.SendBad("You are far too tired to do that.")
		return false
	}
	return true
}

// requireTier fails below the given tier.
func (s *state) requireTier(tier int) bool {
	if s.actor.Tier < tier {
		s.msg.Actor.SendBad("You must be at least tier " + strconv.Itoa(tier) + " to use this skill.")
		return false
	}
	return true
}
