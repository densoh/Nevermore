package objects

import "strings"

// resolveKey mirrors command resolution: an exact key wins, otherwise an
// input of at least two characters that is a prefix of exactly one key
// resolves to that key.
func resolveKey(input string, keys []string) (string, bool) {
	input = strings.ToLower(input)
	for _, key := range keys {
		if key == input {
			return key, true
		}
	}
	if len(input) < 2 {
		return "", false
	}
	match := ""
	for _, key := range keys {
		if strings.HasPrefix(key, input) {
			if match != "" {
				return "", false
			}
			match = key
		}
	}
	return match, match != ""
}

// ResolveSpell finds a spell by exact name or unique prefix.
func ResolveSpell(input string) (Spell, bool) {
	keys := make([]string, 0, len(Spells))
	for key := range Spells {
		keys = append(keys, key)
	}
	if key, ok := resolveKey(input, keys); ok {
		return Spells[key], true
	}
	return Spell{}, false
}

// ResolveSong finds a song by exact name or unique prefix, returning its key.
func ResolveSong(input string) (string, map[string]string, bool) {
	keys := make([]string, 0, len(Songs))
	for key := range Songs {
		keys = append(keys, key)
	}
	if key, ok := resolveKey(input, keys); ok {
		return key, Songs[key], true
	}
	return "", nil, false
}
