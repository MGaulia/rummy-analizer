package mapper

import (
	"main/model"
	"testing"
)

func TestNumberToColor(t *testing.T) {
	tests := []struct {
		input   string
		want    model.Color
		wantErr error
	}{
		{"grey", model.COLOR_GREY, nil},
		{"black", model.COLOR_BLACK, nil},
		{"green", model.COLOR_GREEN, nil},
		{"purple", model.COLOR_PURPLE, nil},
		{"red", "", ErrInvalidColorInput},
	}
	for _, test := range tests {
		got, err := stringToColor(test.input)
		if got != test.want {
			t.Errorf("StringToColor(%s) = %s, want %s", test.input, got, test.want)
		}
		if err != nil && err.Error() != test.wantErr.Error() {
			t.Errorf("StringToColor(%s) = %s, want %s", test.input, err, test.wantErr)
		}
	}
}
