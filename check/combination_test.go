package check

import (
	"main/model"
	"testing"
)

func TestIsCombination(t *testing.T) {
	tests := []struct {
		name  string
		input []model.Card
		want  bool
	}{
		{
			name: "colour mismatch",
			input: []model.Card{
				{Color: model.COLOR_GREY, Number: 1},
				{Color: model.COLOR_BLACK, Number: 2},
			},
		},
		{
			name: "number same",
			input: []model.Card{
				{Color: model.COLOR_GREY, Number: 1},
				{Color: model.COLOR_GREY, Number: 1},
			},
		},
		{
			name: "number too big",
			input: []model.Card{
				{Color: model.COLOR_GREY, Number: 1},
				{Color: model.COLOR_GREY, Number: 3},
			},
		},
		{
			name: "colour same",
			input: []model.Card{
				{Color: model.COLOR_GREY, Number: 1},
				{Color: model.COLOR_GREY, Number: 1},
			},
		},
		{
			name: "not equal numbers",
			input: []model.Card{
				{Color: model.COLOR_GREY, Number: 1},
				{Color: model.COLOR_BLACK, Number: 2},
			},
		},
		{
			name: "valid",
			want: true,
			input: []model.Card{
				{Color: model.COLOR_GREY, Number: 1},
				{Color: model.COLOR_BLACK, Number: 1},
				{Color: model.COLOR_GREEN, Number: 1},
				{Color: model.COLOR_PURPLE, Number: 1},
			},
		},
		{
			name: "true",
			want: true,
			input: []model.Card{
				{Color: model.COLOR_GREY, Number: 1},
				{Color: model.COLOR_GREY, Number: 2},
				{Color: model.COLOR_GREY, Number: 3},
				{Color: model.COLOR_GREY, Number: 4},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsCombination(tt.input)
			if result != tt.want {
				t.Errorf("IsCombination() = %v, want %v", result, tt.want)
			}
		})
	}
}
