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
