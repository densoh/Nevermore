package objects

import "strings"

// nameMatchQuality ranks how well alias matches name, ignoring case: 2 when
// alias starts a word in name, 1 when it only appears mid-word, and 0 when it
// doesn't appear. Searches keep only the best-ranked matches, so "male" picks
// the male wolverine over the female one while "wolf" still finds a werewolf.
func nameMatchQuality(name string, alias string) int {
	name = strings.ToLower(name)
	alias = strings.ToLower(alias)
	if !strings.Contains(name, alias) {
		return 0
	}
	if strings.HasPrefix(name, alias) || strings.Contains(name, " "+alias) || strings.Contains(name, "-"+alias) {
		return 2
	}
	return 1
}
