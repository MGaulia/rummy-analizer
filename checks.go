package main

func isValidSequence(input []Card) (result bool) {
	tocheck := input[0]
	for i := 1; i < len(input); i++ {
		val := input[i]
		if val.Colour != tocheck.Colour {
			return
		}
		if val.Number-tocheck.Number != 1 {
			return
		}
		tocheck = val
	}
	return true
}

func isValidSet(input []Card) (result bool) {
	number := input[0].Number
	usedColours := make(map[Colour]bool)
	for i := 0; i < len(input); i++ {
		val := input[i]
		if val.Number != number {
			return
		}
		if usedColours[val.Colour] {
			return
		}
		usedColours[val.Colour] = true
	}
	return true
}
