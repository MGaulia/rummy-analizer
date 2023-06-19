package checks

import (
	"main/models"
	"testing"
)

func TestIsValidSet(t *testing.T) {
	tests := []struct {
		name  string
		input []models.Card
		want  bool
	}{
		{
			name: "colour same",
			input: []models.Card{
				{Colour: models.COLOR_GREY, Number: 1},
				{Colour: models.COLOR_GREY, Number: 1},
			},
		},
		{
			name: "not equal numbers",
			input: []models.Card{
				{Colour: models.COLOR_GREY, Number: 1},
				{Colour: models.COLOR_BLACK, Number: 2},
			},
		},
		{
			name: "valid",
			want: true,
			input: []models.Card{
				{Colour: models.COLOR_GREY, Number: 1},
				{Colour: models.COLOR_BLACK, Number: 1},
				{Colour: models.COLOR_GREEN, Number: 1},
				{Colour: models.COLOR_PURPLE, Number: 1},
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
