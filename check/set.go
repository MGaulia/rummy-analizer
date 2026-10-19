package check

import "github.com/MGaulia/rummy-analizer/model"

// IsValidSet reports whether cards contain three or four equal numbers in
// distinct colors. Input order does not matter and cards are never modified.
func IsValidSet(cards []model.Card) bool {
	if len(cards) < 3 || len(cards) > 4 {
		return false
	}
	number := cards[0].Number
	if number < model.MinNumber || number > model.MaxNumber {
		return false
	}
	var colors uint8
	for _, card := range cards {
		bit := colorBit(card.Color)
		if card.Number != number || bit == 0 || colors&bit != 0 {
			return false
		}
		colors |= bit
	}
	return true
}
