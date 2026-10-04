package cmd

import (
	"github.com/ArcCS/Nevermore/config"
	"github.com/ArcCS/Nevermore/objects"
	"github.com/ArcCS/Nevermore/permissions"
	"github.com/ArcCS/Nevermore/utils"
	"log"
	"math"
	"strconv"
)

func init() {
	addHandler(scriptRepair{},
		"",
		permissions.Player,
		"$REPAIR")
	addHandler(confirmRepair{},
		"",
		permissions.Player,
		"$CONFIRMREPAIR")
}

type scriptRepair cmd

func (scriptRepair) process(s *state) {
	if len(s.words) < 1 {
		s.msg.Actor.SendBad("Repair what?")
		return
	}

	targetStr := s.words[0]
	targetNum := 1

	if len(s.words) > 1 {
		if val, err := strconv.Atoi(s.words[1]); err == nil {
			targetNum = val
		}
	}

	what := s.actor.Inventory.Search(targetStr, targetNum)
	if what == nil {
		s.msg.Actor.SendBad("You don't have anything like that in your inventory to repair.")
		return
	}
	cost, overhaul, problem := repairQuote(what)
	if problem != "" {
		s.msg.Actor.SendBad(problem)
		return
	}
	if overhaul {
		s.msg.Actor.SendInfo("This item has been patched up too many times and needs a full overhaul.  The overhaul will cost " + strconv.Itoa(cost) + ", and there is a chance it won't survive it, though you won't be charged if it doesn't.  Do you want to overhaul it? (Type yes to overhaul)")
	} else {
		s.msg.Actor.SendInfo("The cost to repair this item will be " + strconv.Itoa(cost) + "." + repairCountNote(what) + "  Do you want to repair it? (Type yes to repair)")
	}
	s.actor.AddCommands("yes", "$CONFIRMREPAIR "+targetStr+" "+strconv.Itoa(targetNum))
}

type confirmRepair cmd

func (confirmRepair) process(s *state) {
	if len(s.words) < 1 {
		s.msg.Actor.SendBad("Repair what?")
		return
	}

	targetStr := s.words[0]
	targetNum := 1

	if len(s.words) > 1 {
		if val, err := strconv.Atoi(s.words[1]); err == nil {
			targetNum = val
		}
	}

	what := s.actor.Inventory.Search(targetStr, targetNum)
	if what == nil {
		s.msg.Actor.SendBad("You don't have anything like that in your inventory to repair.")
		return
	}
	cost, overhaul, problem := repairQuote(what)
	if problem != "" {
		s.msg.Actor.SendBad(problem)
		return
	}
	if !s.actor.Gold.CanSubtract(cost) {
		s.msg.Actor.SendBad("You don't have enough gold to repair this item.")
		return
	}
	if overhaul && utils.Roll(100, 1, 0) <= config.OverhaulFailChance {
		if err := s.actor.Inventory.Remove(what); err != nil {
			log.Println("Error removing item destroyed in overhaul: ", err)
			return
		}
		s.msg.Actor.SendBad("Your " + what.DisplayName() + " falls apart during the overhaul!  You are not charged for the attempt.")
		return
	}
	s.actor.Gold.Subtract(cost)
	what.MaxUses = objects.Items[what.ItemId].MaxUses
	if overhaul {
		what.Repairs = 0
		s.msg.Actor.SendInfo("Your " + what.DisplayName() + " was overhauled and is as good as new.")
		return
	}
	if countsRepairs(what) {
		what.Repairs++
	}
	s.msg.Actor.SendInfo("Your item was repaired.")
}

// countsRepairs reports whether the item keeps a repair count toward an
// overhaul: every weapon and piece of armor but quest loot.
func countsRepairs(what *objects.Item) bool {
	return (utils.IntIn(what.ItemType, config.WeaponTypes) || utils.IntIn(what.ItemType, config.ArmorTypes)) && !what.Flags["quest_loot"]
}

// repairCountNote tells the player how many repairs an item has left before it
// needs an overhaul.
func repairCountNote(what *objects.Item) string {
	if !countsRepairs(what) {
		return ""
	}
	left := config.RepairsBeforeOverhaul - what.Repairs
	if left == 1 {
		return "  After this repair it will need an overhaul."
	}
	return "  It can be repaired " + strconv.Itoa(left-1) + " more times after this before it needs an overhaul."
}

// repairQuote prices a repair of the item and says whether it is an overhaul.
// problem is set, and the rest meaningless, when the item can't be repaired.
func repairQuote(what *objects.Item) (cost int, overhaul bool, problem string) {
	if what.Flags["always_crit"] {
		return 0, false, "This item cannot be repaired."
	}
	if !utils.IntIn(what.ItemType, config.ArmorTypes) && !utils.IntIn(what.ItemType, config.WeaponTypes) {
		return 0, false, "This is not a repairable item."
	}
	maxUses := objects.Items[what.ItemId].MaxUses
	if maxUses <= 0 || what.MaxUses >= maxUses {
		return 0, false, "That item doesn't need repairing."
	}
	price := config.RepairCostFraction * float64(what.Value) * float64(maxUses-what.MaxUses) / float64(maxUses)
	overhaul = countsRepairs(what) && what.Repairs >= config.RepairsBeforeOverhaul
	if overhaul {
		price *= config.OverhaulCostMultiplier
	}
	if what.Flags["quest_loot"] {
		price *= config.QuestLootRepairMultiplier
	}
	return int(math.Round(price)), overhaul, ""
}
