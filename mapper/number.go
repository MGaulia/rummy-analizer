package mapper

import (
	"errors"
	"main/model"
)

var ErrInvalidNumberInput = errors.New("invalid number, available values: 1-13")

func stringToNumber(input string) (num model.Number, err error) {
	switch input {
	case "1":
		num = model.NUMBER_ACE
	case "2":
		num = model.NUMBER_TWO
	case "3":
		num = model.NUMBER_THREE
	case "4":
		num = model.NUMBER_FOUR
	case "5":
		num = model.NUMBER_FIVE
	case "6":
		num = model.NUMBER_SIX
	case "7":
		num = model.NUMBER_SEVEN
	case "8":
		num = model.NUMBER_EIGHT
	case "9":
		num = model.NUMBER_NINE
	case "10":
		num = model.NUMBER_TEN
	case "11":
		num = model.NUMBER_JACK
	case "12":
		num = model.NUMBER_QUEEN
	case "13":
		num = model.NUMBER_KING
	default:
		err = ErrInvalidNumberInput
	}
	return
}
