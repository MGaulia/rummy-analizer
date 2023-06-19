package main

import (
	"bufio"
	"fmt"
	"main/mapper"
	"main/model"
	"os"
	"strconv"
)

func main() {
	var table []model.Card
	var input string
	for input != "exit" {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Enter text: ")
		input, _ = reader.ReadString('\n')

		card, err := mapper.StringToCard(input)
		if err != nil {
			fmt.Println(err)
		} else {
			table = append(table, card)
			printTable(table)
		}
	}
}

func printTable(table []model.Card) {
	result := "["
	var num string
	var color string
	for _, card := range table {
		num = strconv.Itoa(int(card.Number))
		switch card.Color {
		case model.COLOR_GREY:
			color = "grey"
		case model.COLOR_BLACK:
			color = "black"
		case model.COLOR_GREEN:
			color = "green"
		default:
			color = "purple"
		}
		result += color + " " + num + ", "
	}
	result = result + "]"
	fmt.Println(result)
}
