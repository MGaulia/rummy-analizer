package check

import "main/model"

// IsValidSequence checks if the given cards are a valid sequence
func IsValidSequence(input []model.Card) (result bool) {
	tocheck := input[0]
	for i := 1; i < len(input); i++ {
		val := input[i]
		if val.Colour != tocheck.Colour {
			return
		}
		if val.Number-tocheck.Number != 1 {
			return
		}
		tocheck = val
	}
	return true
}
