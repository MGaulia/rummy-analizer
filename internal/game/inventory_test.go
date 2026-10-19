package game

import (
	"errors"
	"reflect"
	"testing"
)

func standardTiles() []Tile {
	var tiles []Tile
	for _, color := range []Color{Black, Blue, Orange, Red} {
		for number := MinNumber; number <= MaxNumber; number++ {
			tile := numbered(color, number)
			tiles = append(tiles, tile, tile)
		}
	}
	return append(tiles, Tile{Kind: Joker}, Tile{Kind: Joker})
}

func TestInventoryWholeSet(t *testing.T) {
	tiles := standardTiles()
	inventory, err := NewInventory(tiles)
	if err != nil || inventory.Len() != 106 {
		t.Fatalf("full set: len=%d, err=%v", inventory.Len(), err)
	}
	for _, tile := range tiles {
		if count, err := inventory.Count(tile); err != nil || count != 2 {
			t.Fatalf("%+v: count=%d, err=%v", tile, count, err)
		}
	}
	for _, tile := range tiles {
		tooMany := append(append([]Tile(nil), tiles...), tile)
		if got, err := NewInventory(tooMany); !errors.Is(err, ErrTooManyCopies) || got != (Inventory{}) {
			t.Fatalf("third copy: %+v, %v", got, err)
		}
	}
	empty, err := inventory.ConsumeAll(tiles)
	if err != nil || empty != (Inventory{}) || inventory.Len() != 106 {
		t.Fatalf("consume set: %+v, %v", empty, err)
	}
}

func TestInventoryConsumesOnlyOneCopy(t *testing.T) {
	tile := numbered(Black, 7)
	tiles := []Tile{tile, tile, numbered(Blue, 3)}
	before := append([]Tile(nil), tiles...)
	inventory, err := NewInventory(tiles)
	if err != nil {
		t.Fatal(err)
	}
	result, err := inventory.Consume(tile)
	if err != nil {
		t.Fatal(err)
	}
	count, _ := result.Count(tile)
	original, _ := inventory.Count(tile)
	if count != 1 || original != 2 || result.Len() != 2 || !reflect.DeepEqual(tiles, before) {
		t.Fatalf("wrong removal or mutation: count=%d original=%d len=%d", count, original, result.Len())
	}
	tiles[0] = numbered(Red, 13)
	if count, _ := inventory.Count(tile); count != 2 {
		t.Fatal("inventory aliases input")
	}
}

func TestInventoryFailuresAreAtomic(t *testing.T) {
	tile := numbered(Blue, 1)
	inventory, err := NewInventory([]Tile{tile})
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		tiles []Tile
		want  error
	}{
		{[]Tile{tile, tile}, ErrTileNotFound},
		{[]Tile{tile, {}}, ErrInvalidTile},
		{[]Tile{numbered(Red, 1)}, ErrTileNotFound},
		{[]Tile{{Kind: Joker}}, ErrTileNotFound},
		{[]Tile{{Kind: Joker, Number: 1}}, ErrInvalidTile},
	} {
		got, err := inventory.ConsumeAll(tc.tiles)
		if !errors.Is(err, tc.want) || got != inventory {
			t.Errorf("ConsumeAll(%+v) = %+v, %v", tc.tiles, got, err)
		}
	}
	if got, err := inventory.ConsumeAll(nil); err != nil || got != inventory {
		t.Fatal("nil request changed state")
	}
	if got, err := NewInventory([]Tile{tile, {}}); !errors.Is(err, ErrInvalidTile) || got != (Inventory{}) {
		t.Fatalf("partial construction: %+v, %v", got, err)
	}
	if count, err := inventory.Count(Tile{}); count != 0 || !errors.Is(err, ErrInvalidTile) {
		t.Fatal("invalid Count accepted")
	}
	if count, err := (Inventory{}).Count(tile); count != 0 || err != nil {
		t.Fatal("empty Count")
	}
}

func FuzzInventoryConsumption(f *testing.F) {
	f.Add([]byte{0, 0, 1, 52}, []byte{0, 52})
	f.Add([]byte{0}, []byte{0, 0})
	f.Fuzz(func(t *testing.T, have, remove []byte) {
		pool := standardTiles()
		decode := func(data []byte) []Tile {
			tiles := make([]Tile, len(data))
			for i, b := range data {
				tiles[i] = pool[2*(int(b)%53)]
			}
			return tiles
		}
		tiles := decode(have)
		inventory, err := NewInventory(tiles)
		if err != nil {
			return
		}
		want := make(map[Tile]int)
		for _, tile := range tiles {
			want[tile]++
		}
		request := decode(remove)
		possible := true
		for _, tile := range request {
			if want[tile] == 0 {
				possible = false
				break
			}
			want[tile]--
		}
		got, err := inventory.ConsumeAll(request)
		if !possible {
			if !errors.Is(err, ErrTileNotFound) || got != inventory {
				t.Fatal("failed removal mutated state")
			}
			return
		}
		if err != nil || got.Len() != len(tiles)-len(request) {
			t.Fatalf("valid removal: %v", err)
		}
		for tile, expected := range want {
			if count, _ := got.Count(tile); count != expected {
				t.Fatal("wrong multiplicity")
			}
		}
	})
}

// ConsumeAll means every requested occurrence, not every matching value in the
// inventory. This is the regression that distinguishes it from old RemoveAll.
func TestConsumeAllUsesRequestedMultiplicities(t *testing.T) {
	tile, other := numbered(Black, 7), numbered(Blue, 3)
	inventory, err := NewInventory([]Tile{tile, tile, other})
	if err != nil {
		t.Fatal(err)
	}
	for copies := 0; copies <= 3; copies++ {
		request := make([]Tile, copies)
		for i := range request {
			request[i] = tile
		}
		result, err := inventory.ConsumeAll(request)
		if copies > 2 {
			if !errors.Is(err, ErrTileNotFound) || result != inventory {
				t.Fatal("excess request did not preserve original state")
			}
			continue
		}
		if err != nil {
			t.Fatal(err)
		}
		count, _ := result.Count(tile)
		otherCount, _ := result.Count(other)
		originalCount, _ := inventory.Count(tile)
		if count != 2-copies || otherCount != 1 || originalCount != 2 {
			t.Fatalf("requested %d: count=%d other=%d original=%d", copies, count, otherCount, originalCount)
		}
	}
}
