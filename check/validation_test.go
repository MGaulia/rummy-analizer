package check

import (
	"main/model"
	"reflect"
	"testing"
)

func TestValidationBoundaries(t *testing.T) {
	grey := model.COLOR_GREY
	black := model.COLOR_BLACK
	tests := []struct {
		name     string
		cards    []model.Card
		set, run bool
	}{
		{name: "nil"},
		{name: "empty", cards: []model.Card{}},
		{name: "one tile", cards: []model.Card{{Color: grey, Number: 1}}},
		{name: "two tile group", cards: []model.Card{{Color: grey, Number: 1}, {Color: black, Number: 1}}},
		{name: "two tile run", cards: []model.Card{{Color: grey, Number: 1}, {Color: grey, Number: 2}}},
		{name: "unordered group", cards: []model.Card{{Color: grey, Number: 13}, {Color: model.COLOR_PURPLE, Number: 13}, {Color: black, Number: 13}}, set: true},
		{name: "unordered run", cards: []model.Card{{Color: grey, Number: 3}, {Color: grey, Number: 1}, {Color: grey, Number: 2}}, run: true},
		{name: "upper run", cards: []model.Card{{Color: grey, Number: 13}, {Color: grey, Number: 11}, {Color: grey, Number: 12}}, run: true},
		{name: "long run", cards: []model.Card{{Color: grey, Number: 5}, {Color: grey, Number: 4}, {Color: grey, Number: 3}, {Color: grey, Number: 2}, {Color: grey, Number: 1}}, run: true},
		{name: "no ace high", cards: []model.Card{{Color: grey, Number: 12}, {Color: grey, Number: 13}, {Color: grey, Number: 1}}},
		{name: "no wrapping", cards: []model.Card{{Color: grey, Number: 13}, {Color: grey, Number: 1}, {Color: grey, Number: 2}}},
		{name: "gap", cards: []model.Card{{Color: grey, Number: 1}, {Color: grey, Number: 2}, {Color: grey, Number: 4}}},
		{name: "duplicate run tile", cards: []model.Card{{Color: grey, Number: 1}, {Color: grey, Number: 2}, {Color: grey, Number: 2}, {Color: grey, Number: 4}}},
		{name: "duplicate group color", cards: []model.Card{{Color: grey, Number: 7}, {Color: black, Number: 7}, {Color: grey, Number: 7}}},
		{name: "mixed run colors", cards: []model.Card{{Color: grey, Number: 1}, {Color: black, Number: 2}, {Color: grey, Number: 3}}},
		{name: "unknown group color", cards: []model.Card{{Color: grey, Number: 7}, {Color: black, Number: 7}, {Color: "unknown", Number: 7}}},
		{name: "unknown run color", cards: []model.Card{{Color: "unknown", Number: 1}, {Color: "unknown", Number: 2}, {Color: "unknown", Number: 3}}},
		{name: "zero group number", cards: []model.Card{{Color: grey, Number: 0}, {Color: black, Number: 0}, {Color: model.COLOR_GREEN, Number: 0}}},
		{name: "high group number", cards: []model.Card{{Color: grey, Number: 14}, {Color: black, Number: 14}, {Color: model.COLOR_GREEN, Number: 14}}},
		{name: "negative run number", cards: []model.Card{{Color: grey, Number: -1}, {Color: grey, Number: 0}, {Color: grey, Number: 1}}},
		{name: "high run number", cards: []model.Card{{Color: grey, Number: 12}, {Color: grey, Number: 13}, {Color: grey, Number: 14}}},
		{name: "five tile group", cards: []model.Card{{Color: grey, Number: 1}, {Color: black, Number: 1}, {Color: model.COLOR_GREEN, Number: 1}, {Color: model.COLOR_PURPLE, Number: 1}, {Color: grey, Number: 1}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			before := append([]model.Card{}, tt.cards...)
			if got := IsValidSet(tt.cards); got != tt.set {
				t.Errorf("IsValidSet() = %v, want %v", got, tt.set)
			}
			if got := IsValidSequence(tt.cards); got != tt.run {
				t.Errorf("IsValidSequence() = %v, want %v", got, tt.run)
			}
			if got := IsCombination(tt.cards); got != (tt.set || tt.run) {
				t.Errorf("IsCombination() = %v, want %v", got, tt.set || tt.run)
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
	for _, color := range []model.Color{model.COLOR_GREY, model.COLOR_BLACK, model.COLOR_GREEN, model.COLOR_PURPLE} {
		for mask := 0; mask < 1<<13; mask++ {
			var cards []model.Card
			for number := 13; number >= 1; number-- {
				if mask&(1<<(number-1)) != 0 {
					cards = append(cards, model.Card{Color: color, Number: model.Number(number)})
				}
			}
			if got := IsValidSequence(cards); got != legal[mask] {
				t.Fatalf("color %s, subset %013b: got %v, want %v", color, mask, got, legal[mask])
			}
		}
	}
}
