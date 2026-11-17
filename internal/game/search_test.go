package game

import (
	"context"
	"fmt"
	"reflect"
	"testing"
)

func TestSearchLargeTableFindsRackExtensions(t *testing.T) {
	// Two copies of each color, divided into short runs, make a legal crowded
	// table with many equivalent partitions. Preserve all physical pieces while
	// adding the known next values; this is independent of any photographed game.
	for _, rack := range [][]Piece{
		{mp("rack-blue", Blue, 13)},
		{mp("rack-blue", Blue, 13), mp("rack-orange", Orange, 13)},
	} {
		t.Run(fmt.Sprintf("%d-rack-tiles", len(rack)), func(t *testing.T) {
			before := Position{Rack: rack}
			for copyIndex := 0; copyIndex < 2; copyIndex++ {
				for _, color := range []Color{Black, Blue, Orange, Red} {
					for start := 1; start <= 10; start += 3 {
						meld := []Piece{}
						for number := start; number < start+3; number++ {
							meld = append(meld, mp(fmt.Sprintf("%d-%s-%d", copyIndex, color, number), color, number))
						}
						before.Table = append(before.Table, meld)
					}
				}
			}
			snapshot := ClonePosition(before)
			known := Move{Table: ClonePosition(before).Table}
			for _, piece := range rack {
				for m, meld := range known.Table {
					if meld[0].Tile.Color == piece.Tile.Color && meld[len(meld)-1].Tile.Number == 12 {
						known.Table[m] = append(meld, piece)
						break
					}
				}
			}
			if checked := ValidateMove(before, Rules{Edition: FourPlayer}, Turn{}, known); !checked.Valid {
				t.Fatalf("invalid known extension: %+v", checked)
			}
			got := Search(context.Background(), before, Rules{Edition: FourPlayer}, Turn{}, SearchOptions{MaxNodes: 200000})
			if got.Status != Optimal || got.Move == nil || got.Verified == nil || len(got.Verified.Played) != len(rack) {
				t.Fatalf("got %+v; want verified optimum playing every rack tile", got)
			}
			if checked := ValidateMove(before, Rules{Edition: FourPlayer}, Turn{}, *got.Move); !checked.Valid {
				t.Fatalf("returned move rejected: %+v", checked)
			}
			if !reflect.DeepEqual(before, snapshot) {
				t.Fatal("search mutated the input position")
			}
		})
	}
}
