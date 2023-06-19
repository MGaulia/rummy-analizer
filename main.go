package main

import (
	"bufio"
	"fmt"
	"main/mapper"
	"main/model"
	"os"
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
			fmt.Println(table)
		}
	}
}
