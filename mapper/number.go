package mapper

import (
	"errors"
	"strconv"

	"github.com/MGaulia/rummy-analizer/model"
)

var ErrInvalidNumberInput = errors.New("invalid number, available values: 1-13")

func stringToNumber(input string) (model.Number, error) {
	number, err := strconv.Atoi(input)
	// Keep the original canonical decimal syntax: no signs or leading zeros.
	if err != nil || number < int(model.MinNumber) || number > int(model.MaxNumber) || strconv.Itoa(number) != input {
		return 0, ErrInvalidNumberInput
	}
	return model.Number(number), nil
}
