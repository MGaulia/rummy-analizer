package mapper

import (
	"errors"
	"main/model"
)

var ErrInvalidColorInput = errors.New("invalid input, available values: grey, black, green, purple")

func StringToColor(input string) (color model.Color, err error) {
	switch input {
	case "grey":
		color = model.COLOR_GREY
	case "black":
		color = model.COLOR_BLACK
	case "green":
		color = model.COLOR_GREEN
	case "purple":
		color = model.COLOR_PURPLE
	default:
		err = ErrInvalidColorInput
	}
	return

}
