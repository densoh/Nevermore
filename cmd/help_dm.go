package cmd

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/ArcCS/Nevermore/config"
	"github.com/ArcCS/Nevermore/objects"
	"github.com/ArcCS/Nevermore/utils"
)

// dmGuide is a staff help page under HELP DM.  page returns the text for the
// words after the topic, or an error line.
type dmGuide struct {
	summary string
	page    func(args []string) string
}

var dmGuides = map[string]dmGuide{
	"WEAPROLLS": {"weaprolls <tier> [2h]  standard drop weapon rolls for a tier", weapRollsGuide},
}

// weapRollsWindow is how far either side of the target average (in points)
// the guide lists rolls.
const weapRollsWindow = 4

// weapRollsLineWidth is the widest a guide line may run before it wraps.
const weapRollsLineWidth = 100

func dmHelp(s *state, args []string) {
	if len(args) == 0 {
		topics := make([]string, 0, len(dmGuides))
		for name := range dmGuides {
			topics = append(topics, name)
		}
		sort.Strings(topics)
		lines := []string{"DM guides (help dm <topic>):"}
		for _, name := range topics {
			lines = append(lines, "  "+dmGuides[name].summary)
		}
		s.msg.Actor.SendGood(strings.Join(lines, "\n"))
		return
	}
	guide, ok := dmGuides[args[0]]
	if !ok {
		s.msg.Actor.SendBad("No DM guide on that. Type help dm for the list.")
		return
	}
	s.msg.Actor.SendGood(guide.page(args[1:]))
}

func weapRollsGuide(args []string) string {
	usage := "Usage: help dm weaprolls <tier 1-" + strconv.Itoa(len(config.MaxWeaponDamage)) + "> [2h]"
	if len(args) < 1 {
		return usage
	}
	tier, err := strconv.Atoi(args[0])
	if _, ok := config.MaxWeaponDamage[tier]; err != nil || !ok {
		return usage
	}
	twoHanded := len(args) > 1 && args[1] == "2H"

	realCap := config.RealWeaponCap(tier, twoHanded)
	pct := config.WeaponRollTargetPercent(tier)
	target2 := realCap * pct * 2 / 100
	kind := "one-handed"
	if twoHanded {
		kind = "two-handed"
	}

	below := []string{"A new weapon only needs to beat most of the tier below."}
	if best := bestDropWeapon(tier-1, twoHanded); best != nil {
		r := config.WeaponRoll{Dice: best.NumDice, Sides: best.SidesDice, Plus: best.PlusDice}
		below = append(below, fmt.Sprintf("  Tier %d best drop: %s (%d) %dd%d+%d, avg %.1f.",
			tier-1, best.Name, best.ItemId, r.Dice, r.Sides, r.Plus, float64(r.Avg2())/2))
	}

	header := []string{
		fmt.Sprintf("Tier %d %s drop rolls.  Real cap %d (highest wieldable max).", tier, kind, realCap),
		"Format: rolls read ndice d sdice + pdice, so 2d11+35 is ndice 2, sdice 11, pdice 35.",
		"Rows: avg (min-max): rolls with that average.",
		"  A 47.0/47.5 row lists the .0 rolls, then the .5 rolls after the /.",
		fmt.Sprintf("Max %d-%d, min at least %d (%d%% of cap).",
			realCap-config.WeaponRollMaxBelowCapMax, realCap-config.WeaponRollMaxBelowCapMin,
			(realCap*config.WeaponRollMinPercent+99)/100, config.WeaponRollMinPercent),
		fmt.Sprintf("Dice: a lone die only up to d%d; at most 4 + 2 x sides dice (8d2, 10d3, 12d4...).",
			config.WeaponRollSingleDieMaxSides),
		fmt.Sprintf("Tier median and mean should sit at %d%% of cap: avg %.1f.  Showing avg %.1f-%.1f.",
			pct, float64(realCap*pct)/100, float64(target2)/2-weapRollsWindow, float64(target2)/2+weapRollsWindow),
		"* indicates target median roll",
	}
	header = append(header, below...)
	special := pct + config.WeaponRollSpecialBonusPercent
	header = append(header,
		fmt.Sprintf("Fewer dice + bigger plus at a lower max = narrower roll.  Special items: max %d, avg ~%.1f (%d%%).",
			realCap, float64(realCap*special)/100, special),
		"Always-crit weapons are exempt.  Fighters wield one tier up; at tier "+strconv.Itoa(config.LevelCap)+" caps no longer apply.",
	)
	var lines []string
	for _, h := range header {
		lines = append(lines, wrapWords(h, weapRollsLineWidth)...)
	}

	// Rows are padded to the page's widest label so the roll lists line up;
	// a row with no label is a section heading.
	type tableRow struct{ label, rolls string }
	var table []tableRow
	width := 0

	rolls := config.StandardWeaponRolls(realCap)
	for max := realCap - config.WeaponRollMaxBelowCapMin; max >= realCap-config.WeaponRollMaxBelowCapMax; max-- {
		table = append(table, tableRow{"", fmt.Sprintf("-- max %d --", max)})
		// One row per average; rolls arrive grouped by it.
		var avgs []int
		row := map[int][]string{}
		for _, r := range rolls {
			if r.Max() != max || r.Avg2() < target2-weapRollsWindow*2 || r.Avg2() > target2+weapRollsWindow*2 {
				continue
			}
			if _, ok := row[r.Avg2()]; !ok {
				avgs = append(avgs, r.Avg2())
			}
			row[r.Avg2()] = append(row[r.Avg2()], fmt.Sprintf("%dd%d+%d", r.Dice, r.Sides, r.Plus))
		}
		// Mark the row nearest the target, compared in hundredths.
		nearest, best := -1, -1
		for _, avg2 := range avgs {
			off := avg2*50 - realCap*pct
			if off < 0 {
				off = -off
			}
			if best < 0 || off < best {
				nearest, best = avg2, off
			}
		}
		// A .5 average shares a row with the .0 below it, its rolls after the slash.
		for k := 0; k < len(avgs); k++ {
			avg2 := avgs[k]
			label := fmt.Sprintf("%.1f (%d-%d)", float64(avg2)/2, avg2-max, max)
			rolls := strings.Join(row[avg2], "  ")
			mark := avg2 == nearest
			if avg2%2 == 1 && k+1 < len(avgs) && avgs[k+1] == avg2-1 {
				low := avg2 - 1
				label = fmt.Sprintf("%.1f/%.1f (%d/%d-%d)", float64(low)/2, float64(avg2)/2, low-max, avg2-max, max)
				rolls = strings.Join(row[low], "  ") + "  /  " + rolls
				mark = mark || low == nearest
				k++
			}
			star := " "
			if mark {
				star = "*"
			}
			label = star + label + ":"
			if len(label) > width {
				width = len(label)
			}
			table = append(table, tableRow{label, rolls})
		}
	}
	// Long roll lists wrap onto lines indented to the roll column.
	for _, t := range table {
		if t.label == "" {
			lines = append(lines, t.rolls)
			continue
		}
		line := fmt.Sprintf("%-*s", width, t.label)
		started := false
		for _, roll := range strings.Split(t.rolls, "  ") {
			sep := "  "
			if !started {
				sep = " "
			}
			if started && len(line)+len(sep)+len(roll) > weapRollsLineWidth {
				lines = append(lines, line)
				line, sep = strings.Repeat(" ", width), " "
			}
			line += sep + roll
			started = true
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

// wrapWords breaks a line at spaces so no piece runs past width, indenting
// the pieces after the first by two.
func wrapWords(text string, width int) []string {
	if len(text) <= width {
		return []string{text}
	}
	var out []string
	line := ""
	for _, word := range strings.Split(text, " ") {
		if line != "" && strings.TrimSpace(line) != "" && len(line)+1+len(word) > width {
			out = append(out, line)
			line = " "
		}
		if line == "" {
			line = word
		} else {
			line += " " + word
		}
	}
	return append(out, line)
}

// bestDropWeapon is the highest-average weapon of a tier and handedness that a
// spawning hostile mob drops, leaving out quest loot and always-crit weapons.
// Ties go to the lower id.
func bestDropWeapon(tier int, twoHanded bool) *objects.Item {
	if tier < 1 {
		return nil
	}
	// Spawn tables are only written by builder commands; a read here is safe enough.
	spawns := map[int]bool{}
	for _, r := range objects.Rooms {
		if r == nil {
			continue
		}
		for mobId, rate := range r.EncounterTable {
			if rate > 0 {
				spawns[mobId] = true
			}
		}
	}
	var best *objects.Item
	bestAvg2 := -1
	for mobId := range spawns {
		m, ok := objects.Mobs[mobId]
		if !ok || m == nil || !m.Flags["hostile"] {
			continue
		}
		for itemId, chance := range m.ItemList {
			i, ok := objects.Items[itemId]
			if !ok || i == nil || chance <= 0 || !utils.IntIn(i.ItemType, []int{0, 1, 2, 3, 4}) ||
				i.Flags["always_crit"] || i.Flags["quest_loot"] || i.Flags["two_handed"] != twoHanded {
				continue
			}
			r := config.WeaponRoll{Dice: i.NumDice, Sides: i.SidesDice, Plus: i.PlusDice}
			if config.WeaponTier(r.Max(), twoHanded) != tier {
				continue
			}
			if r.Avg2() > bestAvg2 || r.Avg2() == bestAvg2 && i.ItemId < best.ItemId {
				best, bestAvg2 = i, r.Avg2()
			}
		}
	}
	return best
}
