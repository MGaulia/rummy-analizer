package mapper

import (
	"errors"
	"main/model"
	"testing"
)

func TestStringToNumber(t *testing.T) {
	tests := []struct {
		input   string
		want    int
		wantErr error
	}{
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
		{"14", 0, errors.New("invalid input")},
	}
	for _, test := range tests {
		got, err := StringToNumber(test.input)
		if got != model.Number(test.want) {
			t.Errorf("StringToNumber(%s) = %d, want %d", test.input, got, test.want)
		}
		if err != nil && err.Error() != test.wantErr.Error() {
			t.Errorf("StringToNumber(%s) = %s, want %s", test.input, err, test.wantErr)
		}
	}
}
