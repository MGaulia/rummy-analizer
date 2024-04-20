package combinations

import (
	"main/model"
	"reflect"
)

var fourSetCombinations = [][]model.Card{
	{{Color: model.COLOR_BLACK, Number: model.NUMBER_ACE}, {Color: model.COLOR_GREEN, Number: model.NUMBER_ACE}, {Color: model.COLOR_GREY, Number: model.NUMBER_ACE}, {Color: model.COLOR_PURPLE, Number: model.NUMBER_ACE}},
	{{Color: model.COLOR_BLACK, Number: model.NUMBER_TWO}, {Color: model.COLOR_GREEN, Number: model.NUMBER_TWO}, {Color: model.COLOR_GREY, Number: model.NUMBER_TWO}, {Color: model.COLOR_PURPLE, Number: model.NUMBER_TWO}},
	{{Color: model.COLOR_BLACK, Number: model.NUMBER_THREE}, {Color: model.COLOR_GREEN, Number: model.NUMBER_THREE}, {Color: model.COLOR_GREY, Number: model.NUMBER_THREE}, {Color: model.COLOR_PURPLE, Number: model.NUMBER_THREE}},
	{{Color: model.COLOR_BLACK, Number: model.NUMBER_FOUR}, {Color: model.COLOR_GREEN, Number: model.NUMBER_FOUR}, {Color: model.COLOR_GREY, Number: model.NUMBER_FOUR}, {Color: model.COLOR_PURPLE, Number: model.NUMBER_FOUR}},
	{{Color: model.COLOR_BLACK, Number: model.NUMBER_FIVE}, {Color: model.COLOR_GREEN, Number: model.NUMBER_FIVE}, {Color: model.COLOR_GREY, Number: model.NUMBER_FIVE}, {Color: model.COLOR_PURPLE, Number: model.NUMBER_FIVE}},
	{{Color: model.COLOR_BLACK, Number: model.NUMBER_SIX}, {Color: model.COLOR_GREEN, Number: model.NUMBER_SIX}, {Color: model.COLOR_GREY, Number: model.NUMBER_SIX}, {Color: model.COLOR_PURPLE, Number: model.NUMBER_SIX}},
	{{Color: model.COLOR_BLACK, Number: model.NUMBER_SEVEN}, {Color: model.COLOR_GREEN, Number: model.NUMBER_SEVEN}, {Color: model.COLOR_GREY, Number: model.NUMBER_SEVEN}, {Color: model.COLOR_PURPLE, Number: model.NUMBER_SEVEN}},
	{{Color: model.COLOR_BLACK, Number: model.NUMBER_EIGHT}, {Color: model.COLOR_GREEN, Number: model.NUMBER_EIGHT}, {Color: model.COLOR_GREY, Number: model.NUMBER_EIGHT}, {Color: model.COLOR_PURPLE, Number: model.NUMBER_EIGHT}},
	{{Color: model.COLOR_BLACK, Number: model.NUMBER_NINE}, {Color: model.COLOR_GREEN, Number: model.NUMBER_NINE}, {Color: model.COLOR_GREY, Number: model.NUMBER_NINE}, {Color: model.COLOR_PURPLE, Number: model.NUMBER_NINE}},
	{{Color: model.COLOR_BLACK, Number: model.NUMBER_TEN}, {Color: model.COLOR_GREEN, Number: model.NUMBER_TEN}, {Color: model.COLOR_GREY, Number: model.NUMBER_TEN}, {Color: model.COLOR_PURPLE, Number: model.NUMBER_TEN}},
	{{Color: model.COLOR_BLACK, Number: model.NUMBER_JACK}, {Color: model.COLOR_GREEN, Number: model.NUMBER_JACK}, {Color: model.COLOR_GREY, Number: model.NUMBER_JACK}, {Color: model.COLOR_PURPLE, Number: model.NUMBER_JACK}},
	{{Color: model.COLOR_BLACK, Number: model.NUMBER_QUEEN}, {Color: model.COLOR_GREEN, Number: model.NUMBER_QUEEN}, {Color: model.COLOR_GREY, Number: model.NUMBER_QUEEN}, {Color: model.COLOR_PURPLE, Number: model.NUMBER_QUEEN}},
	{{Color: model.COLOR_BLACK, Number: model.NUMBER_KING}, {Color: model.COLOR_GREEN, Number: model.NUMBER_KING}, {Color: model.COLOR_GREY, Number: model.NUMBER_KING}, {Color: model.COLOR_PURPLE, Number: model.NUMBER_KING}},
}

func IsFourSetCombination(input []model.Card) bool {
	if len(input) != 4 {
		return false
	}
	for _, combination := range fourSetCombinations {
		// fmt.Println("input: ", input, "combination: ", combination, "result: ", reflect.DeepEqual(input, combination))
		if reflect.DeepEqual(input, combination) {
			return true
		}
	}
	return false
}
