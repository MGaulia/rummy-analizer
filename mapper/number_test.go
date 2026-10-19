package mapper

import (
	"errors"
	"testing"

	"github.com/MGaulia/rummy-analizer/model"
)

func TestStringToNumber(t *testing.T) {
	tests := []struct {
		input   string
		want    int
		wantErr error
	}{
		{"", 0, ErrInvalidNumberInput},
		{"0", 0, ErrInvalidNumberInput},
		{"-1", 0, ErrInvalidNumberInput},
		{"+1", 0, ErrInvalidNumberInput},
		{"01", 0, ErrInvalidNumberInput},
		{"1.0", 0, ErrInvalidNumberInput},
		{"one", 0, ErrInvalidNumberInput},
		{"99999999999999999999999999", 0, ErrInvalidNumberInput},
		{"1", 1, nil},
		{"2", 2, nil},
		{"3", 3, nil},
		{"4", 4, nil},
		{"5", 5, nil},
		{"6", 6, nil},
		{"7", 7, nil},
		{"8", 8, nil},
		{"9", 9, nil},
		{"10", 10, nil},
		{"11", 11, nil},
		{"12", 12, nil},
		{"13", 13, nil},
		{"14", 0, ErrInvalidNumberInput},
	}
	for _, test := range tests {
		got, err := stringToNumber(test.input)
		if got != model.Number(test.want) {
			t.Errorf("StringToNumber(%s) = %d, want %d", test.input, got, test.want)
		}
		if !errors.Is(err, test.wantErr) {
			t.Errorf("StringToNumber(%q) error = %v, want %v", test.input, err, test.wantErr)
		}
	}
}
