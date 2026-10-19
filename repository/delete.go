package repository

import "github.com/MGaulia/rummy-analizer/model"

// Delete returns a copy with every matching card removed, preserving order.
// It does not mutate input. For a future move solver, removing one physical tile
// will need a separate operation because Rummikub includes duplicate copies.
func Delete(input []model.Card, target model.Card) []model.Card {
	var result []model.Card
	for _, card := range input {
		if card != target {
			if result == nil {
				result = make([]model.Card, 0, len(input))
			}
			result = append(result, card)
		}
	}
	return result
}
