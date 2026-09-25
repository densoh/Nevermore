package config

import (
	"math"
	"testing"
)

func TestCalcManaMonkScalesOnPiety(t *testing.T) {
	if got := CalcMana(1, 30, 10, MONK); got != 24 {
		t.Errorf("tier 1 pie 10 = %d, want 24", got)
	}
	if got := CalcMana(25, 5, 25, MONK); got != 150 {
		t.Errorf("tier 25 pie 25 = %d, want 150", got)
	}
	// Intelligence must not matter for monks.
	if CalcMana(10, 5, 15, MONK) != CalcMana(10, 40, 15, MONK) {
		t.Error("monk chi changed with intelligence")
	}
	// Other classes keep the old formula and ignore piety.
	want := (10 * Classes["mage"].Mana) + int(float64(10)*(float64(20)/float64(IntManaPoolDiv))*float64(IntManaPool))
	if got := CalcMana(10, 20, 99, MAGE); got != want {
		t.Errorf("mage mana = %d, want %d", got, want)
	}
}

func TestTodKillChance(t *testing.T) {
	cases := []struct {
		tier       int
		c, h, want float64
	}{
		{10, 1, 1, 0},
		{10, 0.5, 0.9, 0},
		{10, 1, 0.9, 0.015},
		{10, 1, 0.7, 0.135},
		{10, 1, 0.5, 0.375},
		{10, 1, 0.4, 0.54},
		{10, 1, 0.24, TodMaxChance},
		{10, 1, 0, TodMaxChance},
		{10, 0.5, 0.2, 0.135},
		// The coefficient grows with tier: 2.0 at 20, 2.25 at 25.
		{20, 1, 0.5, 0.5},
		{25, 1, 0.5, 0.5625},
		{25, 1, 0.38, TodMaxChance},
	}
	for _, tc := range cases {
		if got := TodKillChance(tc.tier, tc.c, tc.h); math.Abs(got-tc.want) > 1e-9 {
			t.Errorf("TodKillChance(%d, %v, %v) = %v, want %v", tc.tier, tc.c, tc.h, got, tc.want)
		}
	}
	if TodChanceScaleFor(5) != TodChanceScale || TodChanceScaleFor(10) != TodChanceScale {
		t.Error("scale grew below the touch tier")
	}
	if math.Abs(TodChanceScaleFor(25)-2.25) > 1e-9 {
		t.Errorf("scale at 25 = %v, want 2.25", TodChanceScaleFor(25))
	}
}

func TestTodReferenceGrows(t *testing.T) {
	prev := TodReference(1)
	for tier := 2; tier <= 25; tier++ {
		if r := TodReference(tier); r <= prev {
			t.Fatalf("TodReference(%d) = %d not above %d", tier, r, prev)
		} else {
			prev = r
		}
	}
	// R equals a low-piety monk's full bar, so a full-chi touch is at c = 1
	// with nothing to spare; a typical monk keeps a small surplus for flurry.
	for tier := MonkTodTier; tier <= 25; tier++ {
		if MonkMaxChi(tier, 10) != TodReference(tier) {
			t.Errorf("tier %d: low-piety max chi %d should equal R %d", tier, MonkMaxChi(tier, 10), TodReference(tier))
		}
		if MonkMaxChi(tier, 12+tier/4) <= TodReference(tier) {
			t.Errorf("tier %d: typical monk has no chi surplus over R", tier)
		}
	}
}

func TestTodCooldown(t *testing.T) {
	cases := map[int]int{10: 600, 15: 600, 16: 540, 19: 360, 20: 300, 25: 300}
	for tier, want := range cases {
		if got := TodCooldown(tier); got != want {
			t.Errorf("TodCooldown(%d) = %d, want %d", tier, got, want)
		}
	}
}

func TestMonkFlurryFollowsFighterTable(t *testing.T) {
	ace := MultiAttackMultipliers[FlurryMinSkill]
	for skill := 0; skill < FlurryMinSkill; skill++ {
		got := MonkFlurryFor(skill)
		if len(got) != len(ace) || got[0] != ace[0] || got[1] != ace[1] {
			t.Errorf("skill %d flurry = %v, want the Ace row %v", skill, got, ace)
		}
	}
	for skill := FlurryMinSkill; skill <= 10; skill++ {
		want := MultiAttackMultipliers[skill]
		got := MonkFlurryFor(skill)
		if len(got) != len(want) {
			t.Errorf("skill %d flurry = %v, want fighter row %v", skill, got, want)
		}
	}
	if len(MonkFlurryFor(10)) <= len(MonkFlurryFor(5)) {
		t.Error("Grandmaster flurry should out-swing Ace")
	}
}

func TestFlurryChiCostScalesWithSkill(t *testing.T) {
	cases := map[int]int{0: 5, 6: 5, 7: 7, 8: 7, 9: 7, 10: 7}
	for skill, want := range cases {
		if got := FlurryChiCost(skill); got != want {
			t.Errorf("FlurryChiCost(%d) = %d, want %d", skill, got, want)
		}
	}
}

func TestMonkDodgeChance(t *testing.T) {
	base := MonkDodgeChance(0, 0, false)
	if base != MonkDodgeBase {
		t.Errorf("base dodge = %d, want %d", base, MonkDodgeBase)
	}
	if MonkDodgeChance(20, 5, false) <= base {
		t.Error("dex and skill did not raise dodge")
	}
	plain := MonkDodgeChance(20, 5, false)
	if MonkDodgeChance(20, 5, true) != plain+FeintDodgeBonus {
		t.Error("feint bonus not applied")
	}
	if MonkDodgeChance(45, 10, true) != MonkDodgeCap {
		t.Errorf("dodge not capped: %d", MonkDodgeChance(45, 10, true))
	}
}

func TestChiHelpers(t *testing.T) {
	if ChiDecayAmount(5) != 1 {
		t.Errorf("decay on a tiny pool = %d, want 1", ChiDecayAmount(5))
	}
	if ChiDecayAmount(100) != 100*ChiDecayPercent/100 {
		t.Errorf("decay on 100 = %d", ChiDecayAmount(100))
	}
	// The piety share is floored at 1, so low piety still earns 3.
	if ChiPerHitFor(0) != 3 || ChiPerHitFor(7) != 3 {
		t.Errorf("chi per hit at low piety = %d / %d, want 3", ChiPerHitFor(0), ChiPerHitFor(7))
	}
	if ChiPerHitFor(16) != 4 || ChiPerHitFor(24) != 5 {
		t.Errorf("chi per hit at pie 16 / 24 = %d / %d, want 4 / 5", ChiPerHitFor(16), ChiPerHitFor(24))
	}
}

func TestSweepStunChanceBounds(t *testing.T) {
	if SweepStunChance(10, 45, 0) != SweepChanceCap {
		t.Errorf("sweep not capped: %d", SweepStunChance(10, 45, 0))
	}
	if SweepStunChance(0, 0, 50) != 0 {
		t.Errorf("sweep went negative: %d", SweepStunChance(0, 0, 50))
	}
	if SweepStunChance(5, 20, 0) != SweepChance+5*SweepChancePerSkill+20/SweepChanceDexDiv {
		t.Errorf("sweep formula off: %d", SweepStunChance(5, 20, 0))
	}
}

func TestMeditateHelpers(t *testing.T) {
	if MeditateRestorePercent(1, 12) != MeditateBasePercent+12/MeditatePieDiv+1 {
		t.Errorf("restore percent at tier 1 pie 12 = %d", MeditateRestorePercent(1, 12))
	}
	if MeditateDuration(25) != MeditateBaseDuration+25*MeditateDurationPerTier {
		t.Errorf("duration at tier 25 = %d", MeditateDuration(25))
	}
}

func TestMonkReachesGrandmaster(t *testing.T) {
	if WeaponLevel(WeaponExpLevels[10], MONK, 5) != 10 {
		t.Error("monk capped below weapon level 10")
	}
	if WeaponExpTitle(WeaponExpLevels[10], MONK, 5) != WeaponTitles[10] {
		t.Error("monk denied the grandmaster title")
	}
	if WeaponExpNext(WeaponExpLevels[9], MONK, 5) != WeaponExpLevels[10] {
		t.Error("monk next-level lookup stops at 9")
	}
	if WeaponLevel(WeaponExpLevels[10], THIEF, 0) != 9 {
		t.Error("thief unexpectedly reaches level 10")
	}
}

func TestMonkUnarmedRange(t *testing.T) {
	// Tier 10, str 20: base 30, str share 14, roll 2d15 -> 46..74.
	if lo, hi := MonkUnarmedRange(10, 20); lo != 46 || hi != 74 {
		t.Errorf("tier 10 str 20 range = %d..%d, want 46..74", lo, hi)
	}
	// Tier 21, str 20: base 57, str share 26, roll 2d28 -> 85..139.
	if lo, hi := MonkUnarmedRange(21, 20); lo != 85 || hi != 139 {
		t.Errorf("tier 21 str 20 range = %d..%d, want 85..139", lo, hi)
	}
	prevHi := 0
	for tier := 1; tier <= 25; tier++ {
		_, hi := MonkUnarmedRange(tier, 20)
		if hi <= prevHi {
			t.Fatalf("tier %d max %d not above tier %d", tier, hi, tier-1)
		}
		prevHi = hi
	}
}

func TestMonkNaturalArmor(t *testing.T) {
	// Tier 20, con 20: base 300 + 40 = 340, plus 10% from con = 374.
	if got := MonkNaturalArmor(20, 20); got != 374 {
		t.Errorf("tier 20 con 20 armor = %d, want 374", got)
	}
	// Con 0 leaves only the tier term.
	if got := MonkNaturalArmor(10, 0); got != 10*MonkArmorPerLevel {
		t.Errorf("tier 10 con 0 armor = %d, want %d", got, 10*MonkArmorPerLevel)
	}
	if MonkNaturalArmor(10, 30) <= MonkNaturalArmor(10, 10) {
		t.Error("more con should mean more armor")
	}
}

func TestMonkIronBody(t *testing.T) {
	if MonkVitalReduction(14) != 0 || MonkCriticalReduction(14) != 0 {
		t.Error("iron body applied below its tier")
	}
	cases := []struct {
		tier            int
		vital, critical float64
	}{
		{15, 0.15, 0.30},
		{25, 0.25, 0.50},
	}
	for _, tc := range cases {
		if got := MonkVitalReduction(tc.tier); math.Abs(got-tc.vital) > 1e-9 {
			t.Errorf("vital reduction at %d = %v, want %v", tc.tier, got, tc.vital)
		}
		if got := MonkCriticalReduction(tc.tier); math.Abs(got-tc.critical) > 1e-9 {
			t.Errorf("critical reduction at %d = %v, want %v", tc.tier, got, tc.critical)
		}
	}
}

func TestRangerWeaponRules(t *testing.T) {
	if WeaponLevel(WeaponExpLevels[10], RANGER, MissileSkill) != 10 {
		t.Error("ranger denied grandmaster with a bow")
	}
	if WeaponLevel(WeaponExpLevels[10], RANGER, 0) != 9 {
		t.Error("ranger reached grandmaster with a sword")
	}
	if WeaponAdvancementFor(RANGER, MissileSkill) != Classes["ranger"].WeaponAdvancement {
		t.Error("ranger missile advancement should be the class value")
	}
	if WeaponAdvancementFor(RANGER, 0) != RangerMeleeAdvancement {
		t.Error("ranger melee advancement should be the reduced rate")
	}
	if WeaponAdvancementFor(FIGHTER, 0) != Classes["fighter"].WeaponAdvancement {
		t.Error("non-ranger advancement changed")
	}
}

func TestMeditateSaveChance(t *testing.T) {
	// Tier 10 vs a level 10 mob with 20 con: base plus half con.
	if got := MeditateSaveChance(10, 10, 20); got != 60 {
		t.Errorf("t10 vs l10 con20 = %d, want 60", got)
	}
	// Five tiers over the mob and five past the gate: +10 and +5.
	if got := MeditateSaveChance(15, 10, 20); got != 75 {
		t.Errorf("t15 vs l10 con20 = %d, want 75", got)
	}
	// Outleveled: the diff cuts both ways.
	if got := MeditateSaveChance(10, 15, 20); got != 50 {
		t.Errorf("t10 vs l15 con20 = %d, want 50", got)
	}
	if got := MeditateSaveChance(25, 10, 30); got != MeditateSaveCap {
		t.Errorf("chance not capped: %d", got)
	}
	if got := MeditateSaveChance(10, 40, 0); got != 0 {
		t.Errorf("chance not floored: %d", got)
	}
}
