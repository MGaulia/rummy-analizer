package check

import "main/model"

// IsValidSet checks if the given cards are a valid set
func IsValidSet(input []model.Card) (result bool) {
	number := input[0].Number
	usedColours := make(map[model.Color]bool)
	for i := 0; i < len(input); i++ {
		val := input[i]
		if val.Number != number {
			return
		}
		if usedColours[val.Color] {
			return
		}
		usedColours[val.Color] = true
	}
	return true
}
