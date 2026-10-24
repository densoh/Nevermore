package cmd

import (
	"testing"

	"github.com/ArcCS/Nevermore/config"
	"github.com/ArcCS/Nevermore/objects"
)

// attackerState builds the minimum actor DetermineMissChance needs: a class,
// a skill table, stats and a (possibly empty) equipment set.
func attackerState(class int, main *objects.Item) *state {
	skills := map[int]*objects.Accumulator{}
	for slot := 0; slot <= 11; slot++ {
		skills[slot] = &objects.Accumulator{}
	}
	return &state{actor: &objects.Character{
		Class:     class,
		Skills:    skills,
		Flags:     map[string]bool{},
		Modifiers: map[string]int{},
		Equipment: &objects.Equipment{Main: main},
	}}
}

// An unarmed monk must not dereference the empty main hand.  This crashed
// "kill" for every monk before the hand-to-hand skill was read first.
func TestDetermineMissChanceUnarmedMonk(t *testing.T) {
	s := attackerState(config.MONK, nil)
	s.actor.Skills[config.HandSkill].Value = config.SkillExpLevels[3]

	got := DetermineMissChance(s, 0, 1)
	want := config.WeaponMissChance(config.SkillExpLevels[3])
	if got != want {
		t.Errorf("unarmed monk miss chance = %d, want %d (hand skill)", got, want)
	}
}

// A monk holding a weapon still fights with the hand-to-hand skill.
func TestDetermineMissChanceArmedMonkUsesHandSkill(t *testing.T) {
	s := attackerState(config.MONK, &objects.Item{ItemType: 0})
	s.actor.Skills[config.HandSkill].Value = config.SkillExpLevels[5]
	s.actor.Skills[0].Value = 0

	got := DetermineMissChance(s, 0, 1)
	want := config.WeaponMissChance(config.SkillExpLevels[5])
	if got != want {
		t.Errorf("armed monk miss chance = %d, want %d (hand skill)", got, want)
	}
}

// Everyone else reads the wielded weapon's skill.
func TestDetermineMissChanceWeaponSkill(t *testing.T) {
	s := attackerState(config.FIGHTER, &objects.Item{ItemType: 1})
	s.actor.Skills[1].Value = config.SkillExpLevels[4]
	s.actor.Skills[config.HandSkill].Value = config.SkillExpLevels[9]

	got := DetermineMissChance(s, 0, 1)
	want := config.WeaponMissChance(config.SkillExpLevels[4])
	if got != want {
		t.Errorf("fighter miss chance = %d, want %d (weapon skill)", got, want)
	}
}

// Follow-up swings add the penalty before the clamp, so dex works it off down
// to the skill level's floor. Monk flurries follow the same rule on the
// hand-to-hand skill.
func TestFollowUpMissChance(t *testing.T) {
	cases := []struct {
		name              string
		class, level, dex int
		lvlDiff, engaged  int
		want              int
	}{
		{"master 20 dex hits the master floor", config.FIGHTER, 9, 20, 0, 1, 15},
		{"ace 10 dex", config.FIGHTER, 5, 10, 0, 1, 20 + 35 - 10},
		{"specialist 40 dex hits the floor", config.FIGHTER, 8, 40, 0, 1, 20},
		{"expert 15 dex", config.FIGHTER, 7, 15, 0, 1, 10 + 35 - 15},
		{"monk hand level 4 13 dex takes the capped 25", config.MONK, 4, 13, 0, 1, 22 + 25 - 13},
		{"adept monk 13 dex takes the capped 25", config.MONK, 6, 13, 0, 1, 15 + 25 - 13},
		{"specialist monk 30 dex hits the floor", config.MONK, 8, 30, 0, 1, 20},
		{"grandmaster monk keeps the lower 20", config.MONK, 10, 0, 0, 1, 0 + 20},
	}
	for _, c := range cases {
		var main *objects.Item
		skill := config.HandSkill
		if c.class != config.MONK {
			main = &objects.Item{ItemType: 1}
			skill = 1
		}
		s := attackerState(c.class, main)
		s.actor.Skills[skill].Value = config.SkillExpLevels[c.level]
		s.actor.Dex.Current = c.dex
		if got := FollowUpMissChance(s, c.lvlDiff, c.engaged, c.level); got != c.want {
			t.Errorf("%s: follow-up miss = %d, want %d", c.name, got, c.want)
		}
	}
}

// Dex soaks the out-level penalty on follow-ups too, so a mob past the free
// range adds its full penalty to a follow-up that was above the floor.
func TestFollowUpMissChanceOutlevel(t *testing.T) {
	s := attackerState(config.FIGHTER, &objects.Item{ItemType: 1})
	s.actor.Skills[1].Value = config.SkillExpLevels[7]
	s.actor.Dex.Current = 15
	even := FollowUpMissChance(s, 0, 1, 7)
	for diff := 1; diff <= 6; diff++ {
		penalty := config.OutlevelMissPenalty(diff, 1)
		if got := FollowUpMissChance(s, diff, 1, 7); got != even+penalty {
			t.Errorf("mob +%d: follow-up miss = %d, want %d", diff, got, even+penalty)
		}
	}
}

// The miss chance never falls below 5, including when the raw value lands
// between 1 and 4.
func TestMissChanceFloor(t *testing.T) {
	for dex := 15; dex <= 40; dex++ {
		s := attackerState(config.MONK, nil)
		s.actor.Skills[config.HandSkill].Value = config.SkillExpLevels[3] // base 24
		s.actor.Dex.Current = dex
		want := 24 - dex*config.HitPerDex
		if want < 5 {
			want = 5
		}
		if got := DetermineMissChance(s, 0, 1); got != want {
			t.Errorf("dex %d miss chance = %d, want %d", dex, got, want)
		}
	}
}

func TestCanCrush(t *testing.T) {
	item := func(itemType int, twoHanded bool) *objects.Item {
		return &objects.Item{ItemType: itemType, Flags: map[string]bool{"two_handed": twoHanded}}
	}
	cases := []struct {
		name  string
		class int
		main  *objects.Item
		want  bool
	}{
		{"barbarian blunt", config.BARBARIAN, item(2, false), true},
		{"barbarian two-handed sword", config.BARBARIAN, item(0, true), true},
		{"barbarian one-handed sword", config.BARBARIAN, item(0, false), false},
		{"barbarian two-handed bow", config.BARBARIAN, item(4, true), false},
		{"barbarian unarmed", config.BARBARIAN, nil, false},
		{"fighter blunt", config.FIGHTER, item(2, false), false},
	}
	for _, c := range cases {
		if got := canCrush(attackerState(c.class, c.main)); got != c.want {
			t.Errorf("%s: canCrush = %v, want %v", c.name, got, c.want)
		}
	}
}

func TestChiHitsTriplesCritical(t *testing.T) {
	if got := (hitResult{hits: 1}).chiHits(); got != 1 {
		t.Errorf("plain hit: got %d chi hits, want 1", got)
	}
	if got := (hitResult{hits: 1, critical: true}).chiHits(); got != 3 {
		t.Errorf("critical hit: got %d chi hits, want 3", got)
	}
	// Only the first landed hit can crit; the rest count once each.
	if got := (hitResult{hits: 3, critical: true}).chiHits(); got != 5 {
		t.Errorf("critical plus two hits: got %d chi hits, want 5", got)
	}
}

// The level penalty ramps from one level past the engaged count, so a second
// character engaged with the mob takes a level off it.
func TestMissChanceOutlevelGroup(t *testing.T) {
	s := attackerState(config.FIGHTER, &objects.Item{ItemType: 0})
	base := DetermineMissChance(s, 0, 1)
	if got, want := DetermineMissChance(s, 4, 1), base+3*config.MissPerLevel; got != want {
		t.Errorf("solo 4 levels up = %d, want %d", got, want)
	}
	if got, want := DetermineMissChance(s, 4, 2), base+2*config.MissPerLevel; got != want {
		t.Errorf("duo 4 levels up = %d, want %d", got, want)
	}
	if got := DetermineMissChance(s, 1, 1); got != base {
		t.Errorf("solo 1 level up = %d, want %d", got, base)
	}
}
