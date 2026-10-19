package mapper

import (
	"errors"

	"github.com/MGaulia/rummy-analizer/model"
)

var ErrInvalidColorInput = errors.New("invalid color, available values: grey, black, green, purple")

func stringToColor(input string) (model.Color, error) {
	color := model.Color(input)
	switch color {
	case model.COLOR_GREY, model.COLOR_BLACK, model.COLOR_GREEN, model.COLOR_PURPLE:
		return color, nil
	default:
		return "", ErrInvalidColorInput
	}
}
