package check

import "main/model"

// IsSet checks if the given cards are a valid set
func IsSet(input []model.Card) bool {
	return IsValidSet(input) || IsValidSequence(input)
}
