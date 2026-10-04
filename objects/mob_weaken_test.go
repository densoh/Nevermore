package objects

import "testing"

func weakenBoss() *Mob {
	return &Mob{
		Armor:           1000,
		WaterResistance: 90,
		AirResistance:   -20,
		FireResistance:  75,
		EarthResistance: 0,
		Stam:            Meter{Max: 30000, Current: 30000},
		NumDice:         4,
		SidesDice:       50,
		PlusDice:        200,
	}
}

// Steps come off the original values in equal amounts, so after n steps a
// stat sits at (100 - n*pct)% however many steps came before.
func TestWeakenIsLinear(t *testing.T) {
	m := weakenBoss()
	for step := 1; step <= 8; step++ {
		if !m.Weaken(10, 8) {
			t.Fatalf("step %d refused", step)
		}
		left := 100 - 10*step
		if want := 1000 * left / 100; m.Armor < want-2 || m.Armor > want+2 {
			t.Errorf("step %d: armor %d, want about %d", step, m.Armor, want)
		}
		if want := 30000 * left / 100; m.Stam.Max < want-5 || m.Stam.Max > want+5 {
			t.Errorf("step %d: max stamina %d, want about %d", step, m.Stam.Max, want)
		}
	}
	if m.WeakenSteps != 8 {
		t.Errorf("WeakenSteps = %d, want 8", m.WeakenSteps)
	}
	if m.PlusDice != 40 || m.SidesDice != 10 || m.NumDice != 4 {
		t.Errorf("damage dice = %dd%d+%d, want 4d10+40", m.NumDice, m.SidesDice, m.PlusDice)
	}
}

func TestWeakenStopsAtCap(t *testing.T) {
	m := weakenBoss()
	m.WeakenSteps = 8
	if m.Weaken(10, 8) {
		t.Error("weaken past the cap was accepted")
	}
	if m.Armor != 1000 || m.WeakenSteps != 8 {
		t.Error("a refused weaken still changed the mob")
	}
}

func TestWeakenLeavesWeaknessesAlone(t *testing.T) {
	m := weakenBoss()
	m.Weaken(25, 3)
	if m.AirResistance != -20 || m.EarthResistance != 0 {
		t.Errorf("non-positive resists changed: air %d earth %d", m.AirResistance, m.EarthResistance)
	}
	if m.WaterResistance != 68 || m.FireResistance != 56 {
		t.Errorf("resists = water %d fire %d, want 68 56", m.WaterResistance, m.FireResistance)
	}
}

func TestWeakenRejectsBadArgs(t *testing.T) {
	for _, c := range [][2]int{{0, 5}, {10, 0}, {10, 10}, {50, 2}, {-5, 3}} {
		if ValidWeaken(c[0], c[1]) {
			t.Errorf("ValidWeaken(%d, %d) accepted", c[0], c[1])
		}
	}
}

// A wounded boss stays just as wounded when its max shrinks.
func TestSyncWeakenedKeepsWoundFraction(t *testing.T) {
	template := weakenBoss()
	live := weakenBoss()
	live.Stam.Current = 15000
	template.Weaken(20, 4)

	live.SyncWeakened(template)

	if live.Stam.Max != 24000 || live.Stam.Current != 12000 {
		t.Errorf("stamina %d/%d, want 12000/24000", live.Stam.Current, live.Stam.Max)
	}
	if live.Armor != template.Armor || live.WeakenSteps != 1 {
		t.Errorf("live copy not synced: armor %d steps %d", live.Armor, live.WeakenSteps)
	}
}
