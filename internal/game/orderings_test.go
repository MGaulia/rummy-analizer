package game

import (
	"testing"
)

func TestEveryGroupOrdering(t *testing.T) {
	colors := []Color{Blue, Black, Red, Orange}
	for number := 1; number <= 13; number++ {
		for _, a := range colors {
			for _, b := range colors {
				for _, c := range colors {
					cards := []Tile{{Color: a, Number: int(number)}, {Color: b, Number: int(number)}, {Color: c, Number: int(number)}}
					want := a != b && a != c && b != c
					if got := CanFormGroup(cards); got != want {
						t.Fatalf("three-tile group %v: got %v, want %v", cards, got, want)
					}
					for _, d := range colors {
						four := append(cards[:len(cards):len(cards)], Tile{Color: d, Number: int(number)})
						wantFour := want && d != a && d != b && d != c
						if got := CanFormGroup(four); got != wantFour {
							t.Fatalf("four-tile group %v: got %v, want %v", four, got, wantFour)
						}
					}
				}
			}
		}
	}
}

func TestSequenceLengthsAndBoundaries(t *testing.T) {
	for _, color := range []Color{Blue, Black, Red, Orange} {
		for length := 0; length <= 14; length++ {
			for start := 0; start <= 14; start++ {
				var cards []Tile
				for number := start + length - 1; number >= start; number-- {
					cards = append(cards, Tile{Color: color, Number: int(number)})
				}
				valid := start >= 1 && start+length-1 <= 13
				if got := CanFormRun(cards); got != (valid && length >= 3 && length <= 13) {
					t.Fatalf("three-tile check: %v", cards)
				}
			}
		}
		for _, numbers := range [][]int{{12, 13, 1}, {11, 12, 13, 1}, {13, 1, 2}} {
			var cards []Tile
			for _, n := range numbers {
				cards = append(cards, Tile{Color: color, Number: int(n)})
			}
			if CanFormRun(cards) {
				t.Fatalf("accepted wrapping run: %v", cards)
			}
		}
	}
}
