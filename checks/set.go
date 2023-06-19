package checks

import "main/models"

// IsValidSet checks if the given cards are a valid set
func IsValidSet(input []models.Card) (result bool) {
	number := input[0].Number
	usedColours := make(map[models.Color]bool)
	for i := 0; i < len(input); i++ {
		val := input[i]
		if val.Number != number {
			return
		}
		if usedColours[val.Colour] {
			return
		}
		usedColours[val.Colour] = true
	}
	return true
}
