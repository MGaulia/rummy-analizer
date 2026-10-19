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
	count := checkedCount(t, result, tile)
	original := checkedCount(t, inventory, tile)
	if count != 1 || original != 2 || result.Len() != 2 || !reflect.DeepEqual(tiles, before) {
		t.Fatalf("wrong removal or mutation: count=%d original=%d len=%d", count, original, result.Len())
	}
	tiles[0] = numbered(Red, 13)
	if count := checkedCount(t, inventory, tile); count != 2 {
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

func checkedCount(t *testing.T, inventory Inventory, tile Tile) int {
	t.Helper()
	count, err := inventory.Count(tile)
	if err != nil {
		t.Fatalf("Count(%+v): %v", tile, err)
	}
	return count
}

func FuzzInventoryConsumption(f *testing.F) {
	f.Add([]byte{}, []byte{})
	f.Add([]byte{0, 0, 1, 52}, []byte{0, 52})
	f.Add([]byte{0}, []byte{0, 0})
	f.Add([]byte{0, 0, 0}, []byte{})
	var fullSet []byte
	for value := byte(0); value < 53; value++ {
		fullSet = append(fullSet, value, value)
	}
	f.Add(fullSet, fullSet)
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
		before := append([]Tile{}, tiles...)
		// Establish constructor acceptance independently, before calling it.
		want := make(map[Tile]int)
		overfull := false
		for _, tile := range tiles {
			want[tile]++
			overfull = overfull || want[tile] > 2
		}
		inventory, err := NewInventory(tiles)
		if !reflect.DeepEqual(tiles, before) {
			t.Fatal("constructor mutated input")
		}
		if overfull {
			if !errors.Is(err, ErrTooManyCopies) || inventory != (Inventory{}) {
				t.Fatalf("overfull inventory: got %+v, %v", inventory, err)
			}
			return
		}
		if err != nil || inventory.Len() != len(tiles) {
			t.Fatalf("valid constructor input: len=%d, err=%v", inventory.Len(), err)
		}
		// Check absent values too, not just the values present in the input.
		for i := 0; i < 53; i++ {
			tile := pool[2*i]
			if got := checkedCount(t, inventory, tile); got != want[tile] {
				t.Fatalf("constructor Count(%+v) = %d, want %d", tile, got, want[tile])
			}
		}
		original := inventory
		request := decode(remove)
		requestBefore := append([]Tile{}, request...)
		possible := true
		for _, tile := range request {
			if want[tile] == 0 {
				possible = false
				break
			}
			want[tile]--
		}
		got, err := inventory.ConsumeAll(request)
		if inventory != original || !reflect.DeepEqual(request, requestBefore) {
			t.Fatal("consumption mutated caller state")
		}
		if !possible {
			if !errors.Is(err, ErrTileNotFound) || got != inventory {
				t.Fatal("failed removal did not return the original state")
			}
			return
		}
		if err != nil || got.Len() != len(tiles)-len(request) {
			t.Fatalf("valid removal: len=%d, err=%v", got.Len(), err)
		}
		for i := 0; i < 53; i++ {
			tile := pool[2*i]
			if count := checkedCount(t, got, tile); count != want[tile] {
				t.Fatalf("remaining Count(%+v) = %d, want %d", tile, count, want[tile])
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
		count := checkedCount(t, result, tile)
		otherCount := checkedCount(t, result, other)
		originalCount := checkedCount(t, inventory, tile)
		if count != 2-copies || otherCount != 1 || originalCount != 2 {
			t.Fatalf("requested %d: count=%d other=%d original=%d", copies, count, otherCount, originalCount)
		}
	}
}
