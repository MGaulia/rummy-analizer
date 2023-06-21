package check

import "main/model"

// IsCombination checks if the given cards are a valid combination
func IsCombination(input []model.Card) bool {
	return IsValidSet(input) || IsValidSequence(input)
}
