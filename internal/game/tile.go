// Package game contains the Rummikub domain, independent of text or image I/O.
package game

import "errors"

// Color identifies a canonical numbered-tile color, not a camera color reading.
type Color string

const (
	Black  Color = "black"
	Blue   Color = "blue"
	Orange Color = "orange"
	Red    Color = "red"
)

const (
	MinNumber = 1
	MaxNumber = 13
)

// TileKind keeps a joker distinct from a malformed numbered tile.
type TileKind uint8

const (
	Numbered TileKind = iota
	Joker
)

// Tile is a tile value, not the identity of a physical copy. Equal values can
// occur twice in a standard set. Jokers must have an empty Color and Number 0.
// The zero Tile is invalid. Literal construction is allowed; validate at entry
// points rather than assuming that every caller used a constructor.
type Tile struct {
	Kind   TileKind
	Color  Color
	Number int
}

var ErrInvalidTile = errors.New("invalid tile")

// Validate checks representation only, not whether a tile can be played.
func (t Tile) Validate() error {
	switch t.Kind {
	case Numbered:
		if colorIndex(t.Color) < 0 || t.Number < MinNumber || t.Number > MaxNumber {
			return ErrInvalidTile
		}
	case Joker:
		if t.Color != "" || t.Number != 0 {
			return ErrInvalidTile
		}
	default:
		return ErrInvalidTile
	}
	return nil
}

func colorIndex(color Color) int {
	switch color {
	case Black:
		return 0
	case Blue:
		return 1
	case Orange:
		return 2
	case Red:
		return 3
	default:
		return -1
	}
}

// tileIndex is only called after Validate succeeds.
func tileIndex(tile Tile) int {
	if tile.Kind == Joker {
		return 4 * MaxNumber
	}
	return colorIndex(tile.Color)*MaxNumber + tile.Number - MinNumber
}
