package check

import "github.com/MGaulia/rummy-analizer/model"

// IsValidSequence reports whether cards form a same-color run of at least three
// distinct numbers from 1 to 13. Runs cannot wrap from 13 to 1.
// Input order does not matter and cards are never modified.
func IsValidSequence(cards []model.Card) bool {
	if len(cards) < 3 || len(cards) > int(model.MaxNumber) {
		return false
	}
	color := cards[0].Color
	if colorBit(color) == 0 {
		return false
	}
	var seen [model.MaxNumber + 1]bool
	minimum, maximum := model.MaxNumber, model.Number(0)
	for _, card := range cards {
		number := card.Number
		if card.Color != color || number < model.MinNumber || number > model.MaxNumber || seen[number] {
			return false
		}
		seen[number] = true
		if number < minimum {
			minimum = number
		}
		if number > maximum {
			maximum = number
		}
	}
	return int(maximum-minimum)+1 == len(cards)
}
