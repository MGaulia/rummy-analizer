package game

import (
	"errors"
	"testing"
)

func TestEditionInventories(t *testing.T) {
	for _, tc := range []struct {
		edition               Edition
		copies, jokers, total int
	}{{FourPlayer, 2, 2, 106}, {SixPlayer, 3, 4, 160}} {
		var tiles []Tile
		for _, color := range []Color{Black, Blue, Orange, Red} {
			for number := 1; number <= 13; number++ {
				for copy := 0; copy < tc.copies; copy++ {
					tiles = append(tiles, numbered(color, number))
				}
			}
		}
		for j := 0; j < tc.jokers; j++ {
			tiles = append(tiles, Tile{Kind: Joker})
		}
		inventory, err := NewInventoryForEdition(tc.edition, tiles)
		if err != nil || inventory.Len() != tc.total {
			t.Fatalf("edition %d: length=%d err=%v", tc.edition, inventory.Len(), err)
		}
		for _, value := range append([]Tile{numbered(Blue, 7)}, Tile{Kind: Joker}) {
			tooMany := append(append([]Tile(nil), tiles...), value)
			if got, err := NewInventoryForEdition(tc.edition, tooMany); !errors.Is(err, ErrTooManyCopies) || got != (Inventory{}) {
				t.Fatalf("edition %d accepts excess %+v: %v", tc.edition, value, err)
			}
		}
		empty, err := inventory.ConsumeAll(tiles)
		if err != nil || empty != (Inventory{}) || inventory.Len() != tc.total {
			t.Fatal("consumption must be atomic and copy based")
		}
	}
	if _, err := NewInventoryForEdition(5, nil); err == nil {
		t.Fatal("invalid edition accepted")
	}
	if DefaultRules(SixPlayer).NeedsOpening(Turn{Opened: true}) || !DefaultRules(FourPlayer).NeedsOpening(Turn{}) || (Rules{Edition: SixPlayer}).NeedsOpening(Turn{}) {
		t.Fatal("opening state and rules are not independent")
	}
}

func FuzzEditionInventory(f *testing.F) {
	f.Add(false, []byte{52, 52, 52})
	f.Add(true, []byte{52, 52, 52, 52})
	f.Add(true, []byte{0, 0, 0, 0})
	f.Fuzz(func(t *testing.T, expanded bool, data []byte) {
		if len(data) > 161 {
			data = data[:161]
		}
		edition, copies, jokers := FourPlayer, 2, 2
		if expanded {
			edition, copies, jokers = SixPlayer, 3, 4
		}
		colors := []Color{Black, Blue, Orange, Red}
		tiles := make([]Tile, len(data))
		counts := map[Tile]int{}
		over := false
		for i, b := range data {
			v := int(b) % 53
			tile := Tile{Kind: Joker}
			limit := jokers
			if v < 52 {
				tile = numbered(colors[v/13], v%13+1)
				limit = copies
			}
			tiles[i] = tile
			counts[tile]++
			over = over || counts[tile] > limit
		}
		got, err := NewInventoryForEdition(edition, tiles)
		if over {
			if !errors.Is(err, ErrTooManyCopies) || got != (Inventory{}) {
				t.Fatal("incorrect excess acceptance")
			}
			return
		}
		if err != nil || got.Len() != len(tiles) {
			t.Fatalf("valid rejection %v", err)
		}
		for tile, want := range counts {
			if checkedCount(t, got, tile) != want {
				t.Fatal("wrong count")
			}
		}
	})
}
