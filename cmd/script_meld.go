package cmd

import (
	"github.com/ArcCS/Nevermore/objects"
	"log"
	"math"
	"strconv"

	"github.com/ArcCS/Nevermore/config"
	"github.com/ArcCS/Nevermore/permissions"
	"github.com/ArcCS/Nevermore/utils"
)

func init() {
	addHandler(scriptMeld{},
		"",
		permissions.Player,
		"$MELD")
	addHandler(confirmMeld{},
		"",
		permissions.Player,
		"$CONFIRMMELD")
}

// meldItems finds the target and the item being melded into it from
// "target [n] meld [n]". problem is set when the words don't name two items.
func meldItems(s *state) (target, meld *objects.Item, args string, problem string) {
	if len(s.words) < 2 {
		return nil, nil, "", "Meld what into what?"
	}

	argParse := 1
	targetStr := s.words[0]
	targetNum := 1

	if val, err := strconv.Atoi(s.words[1]); err == nil {
		targetNum = val
		argParse = 2
	}

	if argParse == 2 && len(s.words) <= 2 {
		return nil, nil, "", "Meld it into what?"
	}

	meldStr := s.words[argParse]
	meldNum := 1

	if len(s.words) >= argParse+2 {
		if val, err := strconv.Atoi(s.words[argParse+1]); err == nil {
			meldNum = val
		}
	}

	target = s.actor.Inventory.Search(targetStr, targetNum)
	meld = s.actor.Inventory.Search(meldStr, meldNum)
	if target == nil || meld == nil {
		return nil, nil, "", "You have no " + targetStr + " to meld."
	}
	args = targetStr + " " + strconv.Itoa(targetNum) + " " + meldStr + " " + strconv.Itoa(meldNum)
	return target, meld, args, ""
}

// meldPerUse is the price of each use melded in: t² + MeldPriceBase for the
// lowest tier any class casts the spell at, or a third of the template value
// per use when the item holds no known spell.
func meldPerUse(what *objects.Item) int {
	if spell, ok := objects.Spells[what.Spell]; ok && len(spell.Classes) > 0 {
		minTier := math.MaxInt
		for _, tier := range spell.Classes {
			minTier = min(minTier, tier)
		}
		return minTier*minTier + config.MeldPriceBase
	}
	template, ok := objects.Items[what.ItemId]
	if !ok || template.MaxUses <= 0 {
		return 0
	}
	return template.Value / template.MaxUses / 3
}

// meldBaseUses is the uses the item's template comes with, so uses already
// melded onto it aren't paid for again.
func meldBaseUses(what *objects.Item) int {
	if template, ok := objects.Items[what.ItemId]; ok && template.MaxUses > 0 {
		return template.MaxUses
	}
	return what.MaxUses
}

// meldQuote prices melding meld into target, or says why it can't be done.
// It charges for the larger of the two items' base uses, whichever way round
// they're melded.
func meldQuote(target, meld *objects.Item) (cost int, problem string) {
	if target == meld {
		return 0, "You cannot meld the item into itself."
	}
	if !utils.IntIn(meld.ItemType, []int{6, 15}) {
		return 0, "These are not meldable items"
	}
	if meld.ItemType != target.ItemType {
		return 0, "These are not the same item type."
	}
	if meld.Spell != target.Spell {
		return 0, "These items do not contain the same spell."
	}
	price := float64(max(meldPerUse(target), meldPerUse(meld)) * max(meldBaseUses(target), meldBaseUses(meld)))
	if total := target.MaxUses + meld.MaxUses; total > config.MeldPenaltyStep {
		price *= 1 + config.MeldPenaltyPerStep*float64(total/config.MeldPenaltyStep)
	}
	return int(math.Round(price)), ""
}

type scriptMeld cmd

func (scriptMeld) process(s *state) {
	target, meld, args, problem := meldItems(s)
	if problem != "" {
		s.msg.Actor.SendBad(problem)
		return
	}
	cost, problem := meldQuote(target, meld)
	if problem != "" {
		s.msg.Actor.SendBad(problem)
		return
	}
	s.msg.Actor.SendInfo("The cost to meld this item will be " + strconv.Itoa(cost) + ".  Do you want to meld it? (Type yes to meld)")
	s.actor.AddCommands("yes", "$CONFIRMMELD "+args)
	s.actor.AddCommands("y", "$CONFIRMMELD "+args)
}

type confirmMeld cmd

func (confirmMeld) process(s *state) {
	target, meld, _, problem := meldItems(s)
	if problem != "" {
		s.msg.Actor.SendBad("Meld error")
		return
	}
	cost, problem := meldQuote(target, meld)
	if problem != "" {
		s.msg.Actor.SendBad(problem)
		return
	}
	if s.actor.Gold.Value < cost {
		s.msg.Actor.SendBad("You do not have enough money to meld this item.")
		return
	}
	s.actor.Gold.Subtract(cost)
	target.MaxUses += meld.MaxUses
	if err := s.actor.Inventory.Remove(meld); err != nil {
		s.msg.Actor.SendBad("Meld error")
		log.Println("Error removing item: ", err)
		return
	}
	s.msg.Actor.SendGood("Meld completed. You now have " + strconv.Itoa(target.MaxUses) + " uses on this item.")
}
