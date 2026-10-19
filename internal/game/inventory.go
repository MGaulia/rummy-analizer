package game

import "errors"

// CopiesPerTile is the standard edition limit, retained for standard callers.
const CopiesPerTile = 2

var (
	ErrTooManyCopies = errors.New("inventory exceeds the selected edition copy limit")
	ErrTileNotFound  = errors.New("tile is not present in inventory")
)

// Inventory is a bounded multiset of 52 numbered values and one joker value.
// Construction enforces the selected edition. The zero value is empty. Arrays
// give it value semantics: copying a search state never shares mutable storage.
// It tracks counts, not displayed order, physical IDs, melds, or legal moves.
type Inventory struct {
	counts [4*MaxNumber + 1]uint8
}

// NewInventory copies and validates tiles for the four-player edition. On failure it returns an empty
// inventory, never a partially built one. Count combined table+rack tiles here
// when checking the known inventory; checking each separately is insufficient.
func NewInventory(tiles []Tile) (Inventory, error) {
	return NewInventoryForEdition(FourPlayer, tiles)
}

// NewInventoryForEdition uses the selected tile pool. The stored counts retain
// value semantics; consumption cannot increase counts or change edition limits.
func NewInventoryForEdition(edition Edition, tiles []Tile) (Inventory, error) {
	numbered, jokers, _, err := edition.Limits()
	if err != nil {
		return Inventory{}, err
	}
	var inventory Inventory
	for _, tile := range tiles {
		if err := tile.Validate(); err != nil {
			return Inventory{}, err
		}
		index := tileIndex(tile)
		limit := numbered
		if tile.Kind == Joker {
			limit = jokers
		}
		if int(inventory.counts[index]) == limit {
			return Inventory{}, ErrTooManyCopies
		}
		inventory.counts[index]++
	}
	return inventory, nil
}

// Count returns the multiplicity of a valid tile value. Invalid input is an
// error, rather than being confused with a valid tile that is absent.
func (inventory Inventory) Count(tile Tile) (int, error) {
	if err := tile.Validate(); err != nil {
		return 0, err
	}
	return int(inventory.counts[tileIndex(tile)]), nil
}

func (inventory Inventory) Len() int {
	var total int
	for _, count := range inventory.counts {
		total += int(count)
	}
	return total
}

// Consume removes exactly one copy, returning a new value. On failure it
// returns the original inventory. It does not establish that a move is legal.
func (inventory Inventory) Consume(tile Tile) (Inventory, error) {
	if err := tile.Validate(); err != nil {
		return inventory, err
	}
	index := tileIndex(tile)
	if inventory.counts[index] == 0 {
		return inventory, ErrTileNotFound
	}
	inventory.counts[index]--
	return inventory, nil
}

// ConsumeAll is atomic: an invalid or missing tile returns the original state,
// even when earlier requested tiles could have been consumed successfully.
func (inventory Inventory) ConsumeAll(tiles []Tile) (Inventory, error) {
	result := inventory
	for _, tile := range tiles {
		next, err := result.Consume(tile)
		if err != nil {
			return inventory, err
		}
		result = next
	}
	return result, nil
}
