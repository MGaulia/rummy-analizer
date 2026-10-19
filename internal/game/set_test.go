package game

import (
	"testing"
)

func TestCanFormGroup(t *testing.T) {
	tests := []struct {
		name  string
		input []Tile
		want  bool
	}{
		{
			name: "colour same",
			input: []Tile{
				{Color: Blue, Number: 1},
				{Color: Blue, Number: 1},
			},
		},
		{
			name: "not equal numbers",
			input: []Tile{
				{Color: Blue, Number: 1},
				{Color: Black, Number: 2},
			},
		},
		{
			name: "valid",
			want: true,
			input: []Tile{
				{Color: Blue, Number: 1},
				{Color: Black, Number: 1},
				{Color: Red, Number: 1},
				{Color: Orange, Number: 1},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CanFormGroup(tt.input)
			if result != tt.want {
				t.Errorf("isValidSet() = %v, want %v", result, tt.want)
			}
		})
	}
}
