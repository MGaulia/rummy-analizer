package check

import "main/model"

// colorBit returns a distinct bit for each supported color, or zero for an
// invalid color. This keeps validation allocation-free.
func colorBit(color model.Color) uint8 {
	switch color {
	case model.COLOR_GREY:
		return 1
	case model.COLOR_BLACK:
		return 2
	case model.COLOR_GREEN:
		return 4
	case model.COLOR_PURPLE:
		return 8
	default:
		return 0
	}
}
