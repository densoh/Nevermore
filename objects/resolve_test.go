package objects

import "testing"

func TestResolveKey(t *testing.T) {
	keys := []string{"vigor", "vitality", "mend", "heal", "heal-all"}
	cases := []struct {
		in   string
		want string
		ok   bool
	}{
		{"vigor", "vigor", true},
		{"VIG", "vigor", true},
		{"vi", "", false},    // ambiguous: vigor, vitality
		{"m", "", false},     // too short for a prefix match
		{"me", "mend", true},
		{"heal", "heal", true}, // exact beats prefix of heal-all
		{"heal-", "heal-all", true},
		{"zzz", "", false},
	}
	for _, c := range cases {
		got, ok := resolveKey(c.in, keys)
		if got != c.want || ok != c.ok {
			t.Errorf("resolveKey(%q) = %q,%v; want %q,%v", c.in, got, ok, c.want, c.ok)
		}
	}
}

func TestResolveSong(t *testing.T) {
	key, song, ok := ResolveSong("dra")
	if !ok || key != "draens-tale" || song["effect"] != "draens-tale" {
		t.Errorf("ResolveSong(dra) = %q,%v", key, ok)
	}
}
