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
	for {
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
	for _, card := range table {
		num = strconv.Itoa(int(card.Number))
		result += string(card.Color) + " " + num + ", "
	}
	result = result + "]"
	fmt.Println(result)
}
