package combinations

import (
	"testing"

	"github.com/MGaulia/rummy-analizer/model"
)

func TestEveryGroupOrdering(t *testing.T) {
	colors := []model.Color{model.COLOR_GREY, model.COLOR_BLACK, model.COLOR_GREEN, model.COLOR_PURPLE}
	for number := 1; number <= 13; number++ {
		for _, a := range colors {
			for _, b := range colors {
				for _, c := range colors {
					cards := []model.Card{{Color: a, Number: model.Number(number)}, {Color: b, Number: model.Number(number)}, {Color: c, Number: model.Number(number)}}
					want := a != b && a != c && b != c
					if got := IsThreeSetCombination(cards); got != want {
						t.Fatalf("three-tile group %v: got %v, want %v", cards, got, want)
					}
					if IsFourSetCombination(cards) {
						t.Fatal("four-tile check accepted three tiles")
					}
					for _, d := range colors {
						four := append(cards[:len(cards):len(cards)], model.Card{Color: d, Number: model.Number(number)})
						wantFour := want && d != a && d != b && d != c
						if got := IsFourSetCombination(four); got != wantFour {
							t.Fatalf("four-tile group %v: got %v, want %v", four, got, wantFour)
						}
						if IsThreeSetCombination(four) {
							t.Fatal("three-tile check accepted four tiles")
						}
					}
				}
			}
		}
	}
}

func TestSequenceLengthsAndBoundaries(t *testing.T) {
	for _, color := range []model.Color{model.COLOR_GREY, model.COLOR_BLACK, model.COLOR_GREEN, model.COLOR_PURPLE} {
		for length := 0; length <= 14; length++ {
			for start := 0; start <= 14; start++ {
				var cards []model.Card
				for number := start + length - 1; number >= start; number-- {
					cards = append(cards, model.Card{Color: color, Number: model.Number(number)})
				}
				valid := start >= 1 && start+length-1 <= 13
				if got := IsThreeSequenceCombination(cards); got != (valid && length == 3) {
					t.Fatalf("three-tile check: %v", cards)
				}
				if got := IsFourSequenceCombination(cards); got != (valid && length == 4) {
					t.Fatalf("four-tile check: %v", cards)
				}
			}
		}
		for _, numbers := range [][]int{{12, 13, 1}, {11, 12, 13, 1}, {13, 1, 2}} {
			var cards []model.Card
			for _, n := range numbers {
				cards = append(cards, model.Card{Color: color, Number: model.Number(n)})
			}
			if IsThreeSequenceCombination(cards) || IsFourSequenceCombination(cards) {
				t.Fatalf("accepted wrapping run: %v", cards)
			}
		}
	}
}
