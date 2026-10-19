package game

import "fmt"

// Edition selects a tile pool, not the number of people currently playing.
type Edition int

const (
	FourPlayer Edition = 4
	SixPlayer  Edition = 6
)

func (e Edition) Validate() error {
	if e != FourPlayer && e != SixPlayer {
		return fmt.Errorf("edition must be 4 or 6, got %d", e)
	}
	return nil
}

// Limits returns numbered copies, joker copies, and total tiles. Invalid
// editions return an error instead of silently selecting a standard set.
func (e Edition) Limits() (numbered, jokers, total int, err error) {
	if err = e.Validate(); err != nil {
		return
	}
	if e == SixPlayer {
		return 3, 4, 160, nil
	}
	return 2, 2, 106, nil
}

// Rules are game settings. Use DefaultRules to enable the official opening.
// A false OpeningRequired explicitly selects the no-opening house rule.
type Rules struct {
	Edition         Edition `json:"edition"`
	OpeningRequired bool    `json:"opening_required"`
}

func DefaultRules(edition Edition) Rules {
	return Rules{Edition: edition, OpeningRequired: true}
}

// Turn is player state and cannot be inferred from a table photograph. Drawn
// means a tile was already drawn on this turn, which ends the turn.
type Turn struct {
	Opened bool `json:"opened"`
	Drawn  bool `json:"drawn"`
}

func (r Rules) NeedsOpening(turn Turn) bool { return r.OpeningRequired && !turn.Opened }
