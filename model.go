package main

type Number int

type Colour int

const (
	GREY Colour = iota
	BLACK
	GREEN
	PURPLE
)

type Card struct {
	Colour Colour
	Number Number
}
