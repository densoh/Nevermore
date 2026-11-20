package objects

import (
	"testing"

	"github.com/ArcCS/Nevermore/config"
)

func poisonTarget(con int, resist bool) *Character {
	return &Character{
		Flags: map[string]bool{"resist-poison": resist},
		Con:   Meter{Current: con},
		Stam:  Meter{Max: 500, Current: 500},
		Vit:   Meter{Max: 500, Current: 500},
	}
}

// Without resist-poison the full hit lands, whatever the con.
func TestPoisonDamageUnresisted(t *testing.T) {
	c := poisonTarget(30, false)
	if stam, vit, resisted := c.ReceivePoisonDamage(100); stam != 100 || vit != 0 || resisted != 0 {
		t.Errorf("no resist: got %d/%d resisted %d, want 100/0 resisted 0", stam, vit, resisted)
	}
}

// resist-poison blocks a con-scaled share, and the magic-only layers don't
// apply (a dodge flag changes nothing).
func TestPoisonDamageResistScalesWithCon(t *testing.T) {
	c := poisonTarget(20, true)
	c.Flags["dodge"] = true
	want := config.PoisonResistPercent(20) // 25
	if stam, _, resisted := c.ReceivePoisonDamage(100); resisted != want || stam != 100-want {
		t.Errorf("con 20: dealt %d resisted %d, want %d resisted %d", stam, resisted, 100-want, want)
	}
}

// 15% plus 1% per 2 con, rounded down.
func TestPoisonResistPercent(t *testing.T) {
	for con, want := range map[int]int{4: 17, 5: 17, 20: 25, 45: 37} {
		if got := config.PoisonResistPercent(con); got != want {
			t.Errorf("con %d: %d%%, want %d%%", con, got, want)
		}
	}
}

// Mob poison resistance: a percent off, negative adds damage, over 100 blocks
// all without healing, and armor is ignored.
func TestMobPoisonDamage(t *testing.T) {
	cases := []struct{ resist, want int }{{0, 40}, {25, 30}, {-50, 60}, {150, 0}}
	for _, c := range cases {
		m := &Mob{Armor: 100, PoisonResistance: c.resist, Stam: Meter{Max: 100, Current: 100}, Flags: map[string]bool{}}
		if dealt, _, _ := m.ReceivePoisonDamage(40); dealt != c.want {
			t.Errorf("resist %d: dealt %d, want %d", c.resist, dealt, c.want)
		}
		if m.Stam.Current > 100 {
			t.Errorf("resist %d: healed to %d", c.resist, m.Stam.Current)
		}
	}
}
