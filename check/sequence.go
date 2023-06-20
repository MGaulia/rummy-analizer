package check

import "main/model"

// IsValidSequence checks if the given cards are a valid sequence
func IsValidSequence(input []model.Card) (result bool) {
	tocheck := input[0]
	var numbers []model.Number
	for i := 0; i < len(input); i++ {
		val := input[i]
		if val.Color != tocheck.Color {
			return
		}
		numbers = append(numbers, val.Number)
		tocheck = val
	}
	for i := 0; i < len(numbers)-1; i++ {
		if numbers[i+1]-numbers[i] != 1 {
			return
		}
	}
	return true
}
