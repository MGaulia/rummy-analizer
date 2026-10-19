// Package app is the shared offline boundary for CLI and phone adapters.
package app

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/MGaulia/rummy-analizer/internal/game"
	"github.com/MGaulia/rummy-analizer/internal/render"
)

const MaxRequestBytes = 1 << 20

type Settings struct {
	Edition         game.Edition `json:"edition"`
	OpeningRequired *bool        `json:"opening_required,omitempty"`
}

func (s Settings) Rules() game.Rules {
	r := game.DefaultRules(s.Edition)
	if s.OpeningRequired != nil {
		r.OpeningRequired = *s.OpeningRequired
	}
	return r
}

type Request struct {
	Operation string        `json:"operation"`
	Settings  Settings      `json:"settings"`
	Turn      game.Turn     `json:"turn"`
	Position  game.Position `json:"position"`
	Move      *game.Move    `json:"move,omitempty"`
	MaxNodes  int           `json:"max_nodes,omitempty"`
}

type Response struct {
	Position *game.PositionReport `json:"position,omitempty"`
	Search   *game.SearchResult   `json:"search,omitempty"`
	Move     *game.MoveReport     `json:"move,omitempty"`
	SVG      string               `json:"svg,omitempty"`
	Error    string               `json:"error,omitempty"`
}

// Decode rejects unknown fields, oversized requests and trailing JSON values.
func Decode(reader io.Reader, target interface{}) error {
	data, err := io.ReadAll(io.LimitReader(reader, MaxRequestBytes+1))
	if err != nil {
		return err
	}
	if len(data) > MaxRequestBytes {
		return fmt.Errorf("request exceeds %d bytes", MaxRequestBytes)
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	var extra interface{}
	if err := decoder.Decode(&extra); err != io.EOF {
		return fmt.Errorf("expected exactly one JSON value")
	}
	return nil
}

func Execute(ctx context.Context, request Request) Response {
	rules := request.Settings.Rules()
	if err := rules.Edition.Validate(); err != nil {
		return Response{Error: err.Error()}
	}
	position := game.ValidatePosition(request.Position, rules.Edition)
	if !position.Valid {
		return Response{Position: &position, Error: "invalid position"}
	}
	switch request.Operation {
	case "validate":
		svg, err := render.SVG(request.Position, rules.Edition, nil)
		if err != nil {
			return Response{Error: err.Error()}
		}
		return Response{Position: &position, SVG: svg}
	case "check-move":
		if request.Move == nil {
			return Response{Error: "move is required"}
		}
		checked := game.ValidateMove(request.Position, rules, request.Turn, *request.Move)
		response := Response{Move: &checked}
		if checked.Valid {
			svg, err := render.SVG(checked.After, rules.Edition, checked.Played)
			if err != nil {
				return Response{Error: err.Error()}
			}
			response.SVG = svg
		}
		return response
	case "solve":
		found := game.Search(ctx, request.Position, rules, request.Turn, game.SearchOptions{MaxNodes: request.MaxNodes})
		response := Response{Search: &found}
		if found.Move != nil {
			// Independently check at the application boundary before rendering as legal.
			verified := game.ValidateMove(request.Position, rules, request.Turn, *found.Move)
			if !verified.Valid {
				return Response{Error: "solver returned a move rejected by transition validation"}
			}
			svg, err := render.SVG(verified.After, rules.Edition, verified.Played)
			if err != nil {
				return Response{Error: err.Error()}
			}
			response.SVG = svg
		}
		return response
	default:
		return Response{Error: "operation must be validate, solve, or check-move"}
	}
}
