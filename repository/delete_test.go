package repository

import (
	"reflect"
	"testing"

	"github.com/MGaulia/rummy-analizer/model"
)

func TestDelete(t *testing.T) {
	tests := []struct {
		name     string
		input    []model.Card
		todelete model.Card
		want     []model.Card
	}{
		{name: "nil input"},
		{
			name:     "remove all matching copies",
			input:    []model.Card{{Color: model.COLOR_GREY, Number: 1}, {Color: model.COLOR_BLACK, Number: 2}, {Color: model.COLOR_GREY, Number: 1}},
			todelete: model.Card{Color: model.COLOR_GREY, Number: 1},
			want:     []model.Card{{Color: model.COLOR_BLACK, Number: 2}},
		},
		{
			name:     "all removed returns nil",
			input:    []model.Card{{Color: model.COLOR_GREY, Number: 1}},
			todelete: model.Card{Color: model.COLOR_GREY, Number: 1},
		},
		{
			name: "delete success",
			input: []model.Card{
				{Color: model.COLOR_GREY, Number: 1},
				{Color: model.COLOR_BLACK, Number: 2},
			},
			todelete: model.Card{Color: model.COLOR_GREY, Number: 1},
			want: []model.Card{
				{Color: model.COLOR_BLACK, Number: 2},
			},
		},
		{
			name: "not found success",
			input: []model.Card{
				{Color: model.COLOR_GREY, Number: 1},
			},
			todelete: model.Card{Color: model.COLOR_GREEN, Number: 1},
			want: []model.Card{
				{Color: model.COLOR_GREY, Number: 1},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Delete(tt.input, tt.todelete); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Delete() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDeleteDoesNotAliasInput(t *testing.T) {
	input := []model.Card{{Color: model.COLOR_GREY, Number: 1}, {Color: model.COLOR_BLACK, Number: 2}}
	for _, target := range []model.Card{input[0], {Color: model.COLOR_GREEN, Number: 3}} {
		before := append([]model.Card(nil), input...)
		result := Delete(input, target)
		result[0].Number = 13
		if !reflect.DeepEqual(input, before) {
			t.Fatal("mutating the result changed the input")
		}
	}
}
