package logic

import (
	"main/model"
	"testing"

	"reflect"
)

func TestDelete(t *testing.T) {
	tests := []struct {
		name     string
		input    []model.Card
		todelete model.Card
		want     []model.Card
	}{
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
