package logic

import "main/model"

func Delete(input []model.Card, todelete model.Card) (result []model.Card) {
	for _, card := range input {
		if card != todelete {
			result = append(result, card)
		}
	}
	return
}
