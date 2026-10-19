// Package combinations provides fixed-size checks for a single group of cards.
// For runs of arbitrary length, use check.IsCombination or check.IsValidSequence.
package combinations

import (
	"github.com/MGaulia/rummy-analizer/check"
	"github.com/MGaulia/rummy-analizer/model"
)

// IsThreeSetCombination reports whether cards form a three-card set.
func IsThreeSetCombination(cards []model.Card) bool {
	return len(cards) == 3 && check.IsValidSet(cards)
}

// IsFourSetCombination reports whether cards form a four-card set.
func IsFourSetCombination(cards []model.Card) bool {
	return len(cards) == 4 && check.IsValidSet(cards)
}

// IsThreeSequenceCombination reports whether cards form a three-card run.
func IsThreeSequenceCombination(cards []model.Card) bool {
	return len(cards) == 3 && check.IsValidSequence(cards)
}

// IsFourSequenceCombination reports whether cards form a four-card run.
func IsFourSequenceCombination(cards []model.Card) bool {
	return len(cards) == 4 && check.IsValidSequence(cards)
}
