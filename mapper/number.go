package mapper

import (
	"errors"
	"main/model"
)

var ErrInvalidNumberInput = errors.New("invalid input, available values: 1-13")

func stringToNumber(input string) (num model.Number, err error) {
	switch input {
	case "1":
		num = 1
	case "2":
		num = 2
	case "3":
		num = 3
	case "4":
		num = 4
	case "5":
		num = 5
	case "6":
		num = 6
	case "7":
		num = 7
	case "8":
		num = 8
	case "9":
		num = 9
	case "10":
		num = 10
	case "11":
		num = 11
	case "12":
		num = 12
	case "13":
		num = 13
	default:
		err = ErrInvalidNumberInput
	}
	return
}
