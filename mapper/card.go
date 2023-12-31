package mapper

import (
	"errors"
	"main/model"
	"strings"
)

var ErrInputLength = errors.New("invalid input length, must be 5-8 characters")

// StringToCard parses a card from a string
func StringToCard(input string) (card model.Card, err error) {
	input = strings.TrimSpace(input)
	if len(input) < 5 || len(input) > 8 {
		return model.Card{}, ErrInputLength
	}
	input = strings.ToLower(input)

	if tokens := strings.Split(input, " "); len(tokens) == 2 {
		if c, err := stringToColor(tokens[0]); err != nil {
			return model.Card{}, err
		} else {
			card.Color = c
		}
		if n, err := stringToNumber(tokens[1]); err != nil {
			return model.Card{}, err
		} else {
			card.Number = n
		}
	}

	return
}
