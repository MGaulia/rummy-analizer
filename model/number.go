package model

type Number int

// MinNumber and MaxNumber bound the numbered tiles in Rummikub.
const (
	MinNumber Number = 1
	MaxNumber Number = 13
)

// Legacy card-style names remain for source compatibility. In Rummikub these
// are simply the numbers 1–13; 1 never follows 13 in a run.
const (
	NUMBER_ACE   = 1
	NUMBER_TWO   = 2
	NUMBER_THREE = 3
	NUMBER_FOUR  = 4
	NUMBER_FIVE  = 5
	NUMBER_SIX   = 6
	NUMBER_SEVEN = 7
	NUMBER_EIGHT = 8
	NUMBER_NINE  = 9
	NUMBER_TEN   = 10
	NUMBER_JACK  = 11
	NUMBER_QUEEN = 12
	NUMBER_KING  = 13
)
