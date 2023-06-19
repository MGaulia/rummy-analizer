package check

import (
	"main/model"
	"testing"
)

func TestIsValidSet(t *testing.T) {
	tests := []struct {
		name  string
		input []model.Card
		want  bool
	}{
		{
			name: "colour same",
			input: []model.Card{
				{Colour: model.COLOR_GREY, Number: 1},
				{Colour: model.COLOR_GREY, Number: 1},
			},
		},
		{
			name: "not equal numbers",
			input: []model.Card{
				{Colour: model.COLOR_GREY, Number: 1},
				{Colour: model.COLOR_BLACK, Number: 2},
			},
		},
		{
			name: "valid",
			want: true,
			input: []model.Card{
				{Colour: model.COLOR_GREY, Number: 1},
				{Colour: model.COLOR_BLACK, Number: 1},
				{Colour: model.COLOR_GREEN, Number: 1},
				{Colour: model.COLOR_PURPLE, Number: 1},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidSet(tt.input)
			if result != tt.want {
				t.Errorf("isValidSet() = %v, want %v", result, tt.want)
			}
		})
	}
}
