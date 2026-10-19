package check

import "main/model"

// IsCombination reports whether cards form one complete set or sequence.
// It does not search a larger hand or validate a table containing several groups.
func IsCombination(input []model.Card) bool {
	return IsValidSet(input) || IsValidSequence(input)
}
