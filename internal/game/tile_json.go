package game

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// Tile JSON uses a joker boolean instead of exposing internal enum ordinals.
// Decoding preserves malformed representations for position validation to
// report with their physical ID and location; unknown fields are rejected.
func (t Tile) MarshalJSON() ([]byte, error) {
	if err := t.Validate(); err != nil {
		return nil, err
	}
	return json.Marshal(struct {
		Joker  bool  `json:"joker,omitempty"`
		Color  Color `json:"color,omitempty"`
		Number int   `json:"number,omitempty"`
	}{t.Kind == Joker, t.Color, t.Number})
}

func (t *Tile) UnmarshalJSON(data []byte) error {
	if bytes.Equal(bytes.TrimSpace(data), []byte("null")) {
		return fmt.Errorf("tile must be an object, not null")
	}
	var wire struct {
		Joker  bool  `json:"joker"`
		Color  Color `json:"color"`
		Number int   `json:"number"`
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&wire); err != nil {
		return err
	}
	kind := Numbered
	if wire.Joker {
		kind = Joker
	}
	*t = Tile{Kind: kind, Color: wire.Color, Number: wire.Number}
	return nil
}
