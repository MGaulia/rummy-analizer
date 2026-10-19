package game

import (
	"reflect"
	"testing"
)

func TestValidationBoundaries(t *testing.T) {
	blue := Blue
	black := Black
	tests := []struct {
		name     string
		cards    []Tile
		set, run bool
	}{
		{name: "nil"},
		{name: "empty", cards: []Tile{}},
		{name: "one tile", cards: []Tile{{Color: blue, Number: 1}}},
		{name: "two tile group", cards: []Tile{{Color: blue, Number: 1}, {Color: black, Number: 1}}},
		{name: "two tile run", cards: []Tile{{Color: blue, Number: 1}, {Color: blue, Number: 2}}},
		{name: "unordered group", cards: []Tile{{Color: blue, Number: 13}, {Color: Orange, Number: 13}, {Color: black, Number: 13}}, set: true},
		{name: "unordered run", cards: []Tile{{Color: blue, Number: 3}, {Color: blue, Number: 1}, {Color: blue, Number: 2}}, run: true},
		{name: "upper run", cards: []Tile{{Color: blue, Number: 13}, {Color: blue, Number: 11}, {Color: blue, Number: 12}}, run: true},
		{name: "long run", cards: []Tile{{Color: blue, Number: 5}, {Color: blue, Number: 4}, {Color: blue, Number: 3}, {Color: blue, Number: 2}, {Color: blue, Number: 1}}, run: true},
		{name: "no ace high", cards: []Tile{{Color: blue, Number: 12}, {Color: blue, Number: 13}, {Color: blue, Number: 1}}},
		{name: "no wrapping", cards: []Tile{{Color: blue, Number: 13}, {Color: blue, Number: 1}, {Color: blue, Number: 2}}},
		{name: "gap", cards: []Tile{{Color: blue, Number: 1}, {Color: blue, Number: 2}, {Color: blue, Number: 4}}},
		{name: "duplicate run tile", cards: []Tile{{Color: blue, Number: 1}, {Color: blue, Number: 2}, {Color: blue, Number: 2}, {Color: blue, Number: 4}}},
		{name: "duplicate group color", cards: []Tile{{Color: blue, Number: 7}, {Color: black, Number: 7}, {Color: blue, Number: 7}}},
		{name: "mixed run colors", cards: []Tile{{Color: blue, Number: 1}, {Color: black, Number: 2}, {Color: blue, Number: 3}}},
		{name: "unknown group color", cards: []Tile{{Color: blue, Number: 7}, {Color: black, Number: 7}, {Color: "unknown", Number: 7}}},
		{name: "unknown run color", cards: []Tile{{Color: "unknown", Number: 1}, {Color: "unknown", Number: 2}, {Color: "unknown", Number: 3}}},
		{name: "zero group number", cards: []Tile{{Color: blue, Number: 0}, {Color: black, Number: 0}, {Color: Red, Number: 0}}},
		{name: "high group number", cards: []Tile{{Color: blue, Number: 14}, {Color: black, Number: 14}, {Color: Red, Number: 14}}},
		{name: "negative run number", cards: []Tile{{Color: blue, Number: -1}, {Color: blue, Number: 0}, {Color: blue, Number: 1}}},
		{name: "high run number", cards: []Tile{{Color: blue, Number: 12}, {Color: blue, Number: 13}, {Color: blue, Number: 14}}},
		{name: "five tile group", cards: []Tile{{Color: blue, Number: 1}, {Color: black, Number: 1}, {Color: Red, Number: 1}, {Color: Orange, Number: 1}, {Color: blue, Number: 1}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			before := append([]Tile{}, tt.cards...)
			if got := CanFormGroup(tt.cards); got != tt.set {
				t.Errorf("CanFormGroup() = %v, want %v", got, tt.set)
			}
			if got := CanFormRun(tt.cards); got != tt.run {
				t.Errorf("CanFormRun() = %v, want %v", got, tt.run)
			}
			if got := CanFormMeld(tt.cards); got != (tt.set || tt.run) {
				t.Errorf("CanFormMeld() = %v, want %v", got, tt.set || tt.run)
			}
			if len(tt.cards) > 0 && !reflect.DeepEqual(tt.cards, before) {
				t.Fatal("validation mutated the input")
			}
		})
	}
}

// Test every subset of 1–13 against independently generated legal intervals.
// Reverse order ensures validation does not rely on sorting the input.
func TestEveryNumberSubset(t *testing.T) {
	legal := make(map[int]bool)
	for start := 1; start <= 11; start++ {
		mask := 0
		for end := start; end <= 13; end++ {
			mask |= 1 << (end - 1)
			if end-start+1 >= 3 {
				legal[mask] = true
			}
		}
	}
	for _, color := range []Color{Blue, Black, Red, Orange} {
		for mask := 0; mask < 1<<13; mask++ {
			var cards []Tile
			for number := 13; number >= 1; number-- {
				if mask&(1<<(number-1)) != 0 {
					cards = append(cards, Tile{Color: color, Number: int(number)})
				}
			}
			if got := CanFormRun(cards); got != legal[mask] {
				t.Fatalf("color %s, subset %013b: got %v, want %v", color, mask, got, legal[mask])
			}
		}
	}
}
