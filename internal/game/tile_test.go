package game

import (
	"errors"
	"testing"
)

func numbered(color Color, number int) Tile { return Tile{Color: color, Number: number} }

func TestTileRepresentation(t *testing.T) {
	for _, color := range []Color{Black, Blue, Orange, Red} {
		for number := MinNumber; number <= MaxNumber; number++ {
			if err := numbered(color, number).Validate(); err != nil {
				t.Fatalf("%s %d: %v", color, number, err)
			}
		}
	}
	if err := (Tile{Kind: Joker}).Validate(); err != nil {
		t.Fatal(err)
	}
	for _, tile := range []Tile{
		{}, numbered(Black, 0), numbered(Black, 14), numbered(Blue, -1),
		numbered("", 1), numbered("grey", 1), numbered("BLUE", 1),
		{Kind: Joker, Color: Red}, {Kind: Joker, Number: 1},
		{Kind: TileKind(255), Color: Black, Number: 1},
	} {
		if !errors.Is(tile.Validate(), ErrInvalidTile) {
			t.Errorf("accepted invalid representation: %+v", tile)
		}
	}
}
