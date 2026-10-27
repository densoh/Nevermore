package objects

import "testing"

// Armor of 200 or more, or a resistance over 100%, stops a hit outright; it
// never heals the mob.
func TestMobDamageNeverHeals(t *testing.T) {
	armored := &Mob{Armor: 2000, Stam: Meter{Max: 100, Current: 50}, Flags: map[string]bool{}}
	if dealt, _, resisted := armored.ReceiveDamage(40); dealt != 0 || resisted != 40 {
		t.Errorf("armor 2000: dealt %d resisted %d, want 0 and 40", dealt, resisted)
	}
	if armored.Stam.Current != 50 {
		t.Errorf("armor 2000: stamina %d, want 50", armored.Stam.Current)
	}

	resistant := &Mob{FireResistance: 150, Stam: Meter{Max: 100, Current: 50}, Flags: map[string]bool{}}
	if dealt, _, _ := resistant.ReceiveMagicDamage(40, "fire"); dealt != 0 {
		t.Errorf("150%% fire resist: dealt %d, want 0", dealt)
	}
	if resistant.Stam.Current != 50 {
		t.Errorf("150%% fire resist: stamina %d, want 50", resistant.Stam.Current)
	}

	plain := &Mob{Armor: 100, Stam: Meter{Max: 100, Current: 50}, Flags: map[string]bool{}}
	if dealt, _, _ := plain.ReceiveDamage(40); dealt != 20 || plain.Stam.Current != 30 {
		t.Errorf("armor 100: dealt %d stamina %d, want 20 and 30", dealt, plain.Stam.Current)
	}
}
