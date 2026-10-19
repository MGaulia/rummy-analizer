package mapper

import (
	"errors"
	"strings"

	"github.com/MGaulia/rummy-analizer/model"
)

var ErrInputFormat = errors.New("invalid input, expected <color> <number>")

// ErrInputLength is retained for callers using the old error name.
// Deprecated: use ErrInputFormat. Input is validated by fields, not byte length.
var ErrInputLength = ErrInputFormat

// StringToCard parses a card from a string
func StringToCard(input string) (model.Card, error) {
	fields := strings.Fields(input)
	if len(fields) != 2 {
		return model.Card{}, ErrInputFormat
	}
	color, err := stringToColor(strings.ToLower(fields[0]))
	if err != nil {
		return model.Card{}, err
	}
	number, err := stringToNumber(fields[1])
	if err != nil {
		return model.Card{}, err
	}
	return model.Card{Color: color, Number: number}, nil
}
