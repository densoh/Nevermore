package config

import "testing"

func TestGearArmor(t *testing.T) {
	cases := []struct {
		name                                     string
		class, tier, con, dex, gear, light, want int
	}{
		// con only: 520 + 520*20*.005
		{"paladin", PALADIN, 19, 20, 15, 520, 0, 572},
		{"cleric ignores dex on light", CLERIC, 19, 20, 15, 520, 375, 572},
		// heavy set + shield + neck/rings: no light pieces, so no dex
		{"fighter heavy", FIGHTER, 19, 20, 15, 590, 0, 649},
		// light set: 520 + 52 + 375*15*.005 (28.1)
		{"fighter light", FIGHTER, 19, 20, 15, 520, 375, 600},
		// 440 + 33 + 375*22*.005 (41.25)
		{"thief", THIEF, 19, 15, 22, 440, 375, 514},
		{"ranger", RANGER, 19, 15, 22, 440, 375, 514},
		// 520 + 520*(25*.005 + 25*.005 + 19*.015)
		{"barbarian t19", BARBARIAN, 19, 25, 15, 520, 375, 798},
		// below ToughnessTier: con counted twice only
		{"barbarian t9", BARBARIAN, 9, 25, 15, 520, 375, 650},
		// dex replaces the second con when higher: 520*(.1 + .15 + .15)
		{"barbarian high dex", BARBARIAN, 10, 20, 30, 520, 0, 728},
	}
	for _, c := range cases {
		if got := GearArmor(c.class, c.tier, c.con, c.dex, c.gear, c.light); got != c.want {
			t.Errorf("%s: GearArmor = %d, want %d", c.name, got, c.want)
		}
	}
}

func TestInertialBarrierIgnore(t *testing.T) {
	cases := []struct {
		tier, intel int
		want        float64
	}{
		{9, 20, .19},    // tier term starts above 9
		{10, 20, .1925}, // 15 + 0.25 + 4
		{15, 25, .215},  // 15 + 1.5 + 5
		{20, 30, .2375}, // 15 + 2.75 + 6
		{25, 45, .28},   // theoretical top
		{5, 20, .19},    // never negative below the start tier
	}
	for _, c := range cases {
		got := InertialBarrierIgnore(c.tier, c.intel)
		if got < c.want-1e-9 || got > c.want+1e-9 {
			t.Errorf("InertialBarrierIgnore(%d, %d) = %.4f, want %.4f", c.tier, c.intel, got, c.want)
		}
	}
}

func TestCheckArmorHeavyUsesNextBand(t *testing.T) {
	const light, heavy = 1, 3
	cases := []struct {
		name                    string
		aType, tier, val, class int
		want                    bool
	}{
		{"t1 light body at cap", 5, 1, 50, light, true},
		{"t10 light body at cap", 5, 10, 90, light, true},
		{"t10 light body over cap", 5, 10, 91, light, false},
		{"t10 heavy body reads t15-19", 5, 10, 110, heavy, true},
		{"t10 heavy body over its cap", 5, 10, 111, heavy, false},
		{"t15 heavy legs reads t20-25", 20, 15, 80, heavy, true},
		{"t15 heavy legs over its cap", 20, 15, 81, heavy, false},
		{"t20 light body", 5, 20, 130, light, true},
		{"t20 heavy body reads the heavy row", 5, 20, 150, heavy, true},
		{"t25 heavy hands at cap", 26, 25, 65, heavy, true},
		{"t25 heavy head over cap", 25, 25, 81, heavy, false},
		{"unclassed neck uses its own band", 22, 15, 45, 0, true},
		{"past the last band", 5, 26, 10, light, false},
	}
	for _, c := range cases {
		if got := CheckArmor(c.aType, c.tier, c.val, c.class); got != c.want {
			t.Errorf("%s: CheckArmor = %v, want %v", c.name, got, c.want)
		}
	}
}

func TestHeavyArmorFits(t *testing.T) {
	cases := []struct {
		itemType, armor int
		want            bool
	}{
		{5, 70, false}, // t5-9 body cap
		{5, 71, true},
		{20, 50, false}, // t5-9 legs cap
		{20, 55, true},
		{26, 25, false}, // t5-9 hands cap
		{26, 30, true},
	}
	for _, c := range cases {
		if got := HeavyArmorFits(c.itemType, c.armor); got != c.want {
			t.Errorf("HeavyArmorFits(%d, %d) = %v, want %v", c.itemType, c.armor, got, c.want)
		}
	}
}

func TestArmorDescription(t *testing.T) {
	cases := map[[2]int]string{
		{5, 3}:  "heavy body armor",
		{20, 1}: "light legs armor",
		{5, 4}:  "robe body armor",
		{24, 0}: "finger armor",
	}
	for in, want := range cases {
		if got := ArmorDescription(in[0], in[1]); got != want {
			t.Errorf("ArmorDescription(%d, %d) = %q, want %q", in[0], in[1], got, want)
		}
	}
}

func TestArmorClassValid(t *testing.T) {
	cases := []struct {
		name                   string
		itemType, armor, class int
		want                   bool
	}{
		{"none on anything", 0, 0, 0, true},
		{"light body", 5, 60, 1, true},
		{"heavy body above floor", 5, 80, 3, true},
		{"heavy body at t5-9 level", 5, 70, 3, false},
		{"light on a ring", 24, 7, 1, false},
		{"robe on a shield", 23, 50, 4, false},
		{"light on a weapon", 0, 0, 1, false},
	}
	for _, c := range cases {
		if got := ArmorClassValid(c.itemType, c.armor, c.class); got != c.want {
			t.Errorf("%s: ArmorClassValid = %v, want %v", c.name, got, c.want)
		}
	}
}

func TestArmorClassFitsType(t *testing.T) {
	for _, itemType := range []int{22, 23, 24} {
		if !ArmorClassFitsType(itemType, 0) {
			t.Errorf("type %d should accept none", itemType)
		}
		for _, class := range []int{1, 3, 4} {
			if ArmorClassFitsType(itemType, class) {
				t.Errorf("type %d should reject class %d", itemType, class)
			}
		}
	}
	for _, itemType := range []int{5, 19, 20, 21, 25, 26} {
		for _, class := range []int{0, 1, 3, 4} {
			if !ArmorClassFitsType(itemType, class) {
				t.Errorf("type %d should accept class %d", itemType, class)
			}
		}
	}
}

func TestParseArmorClass(t *testing.T) {
	valid := map[string]int{
		"none": 0, "0": 0,
		"light": 1, "Light": 1, "1": 1,
		"heavy": 3, "HEAVY": 3, "3": 3,
		"robe": 4, "4": 4,
	}
	for input, want := range valid {
		got, ok := ParseArmorClass(input)
		if !ok || got != want {
			t.Errorf("ParseArmorClass(%q) = %d, %v; want %d, true", input, got, ok, want)
		}
	}
	for _, input := range []string{"medium", "2", "5", "-1", "", "plate"} {
		if _, ok := ParseArmorClass(input); ok {
			t.Errorf("ParseArmorClass(%q) should be rejected", input)
		}
	}
}
