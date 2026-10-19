package cli

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/MGaulia/rummy-analizer/internal/game"
)

func TestRun(t *testing.T) {
	tests := []struct {
		name, input, want string
		count             int
	}{
		{"empty EOF", "", "Available colors:", 0},
		{"blank lines", "\n \t\n", "Available colors:", 0},
		{"piped lines", "blue 1\nblue 2\nblue 3\n", "[blue 1, blue 2, blue 3]\nCan form group/run: true", 3},
		{"unterminated last line", "blue 1\nblue 2\nblue 3", "[blue 1, blue 2, blue 3]\nCan form group/run: true", 3},
		{"CRLF", "blue 1\r\nblue 2\r\nblue 3\r\n", "[blue 1, blue 2, blue 3]\nCan form group/run: true", 3},
		{"recover after invalid input", "blue1\nblue 1\nunknown 7\nblue 2\nblue 3\n", "[blue 1, blue 2, blue 3]\nCan form group/run: true", 3},
		{"long run", "blue 1\nblue 2\nblue 3\nblue 4\nblue 5\n", "[blue 1, blue 2, blue 3, blue 4, blue 5]\nCan form group/run: true", 5},
		{"Rummikub no wrap", "blue 12\nblue 13\nblue 1\n", "[blue 12, blue 13, blue 1]\nCan form group/run: false", 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var output bytes.Buffer
			if err := Run(nil, strings.NewReader(tt.input), &output); err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(output.String(), tt.want) {
				t.Fatalf("output %q does not contain %q", output.String(), tt.want)
			}
			if got := strings.Count(output.String(), "Can form group/run:"); got != tt.count {
				t.Fatalf("processed %d tiles, want %d; output %q", got, tt.count, output.String())
			}
		})
	}
}

var errIO = errors.New("test I/O failure")

type failingReader struct{}

func (failingReader) Read([]byte) (int, error) { return 0, errIO }

type failingWriter struct{ writesLeft int }

func (w *failingWriter) Write(p []byte) (int, error) {
	if w.writesLeft == 0 {
		return 0, errIO
	}
	w.writesLeft--
	return len(p), nil
}

func TestRunIOErrors(t *testing.T) {
	input := io.MultiReader(strings.NewReader("blue 1\n"), failingReader{})
	if err := Run(nil, input, io.Discard); !errors.Is(err, errIO) {
		t.Fatalf("read error = %v, want %v", err, errIO)
	}
	for _, input := range []string{"blue 1\n", "invalid\n"} {
		for writesLeft := 0; writesLeft <= 2; writesLeft++ {
			err := Run(nil, strings.NewReader(input), &failingWriter{writesLeft: writesLeft})
			if !errors.Is(err, errIO) {
				t.Fatalf("write error after %d writes = %v, want %v", writesLeft, err, errIO)
			}
		}
	}
	if err := Run(nil, strings.NewReader(strings.Repeat("x", 70_000)), io.Discard); err == nil {
		t.Fatal("expected scanner error for an oversized line")
	}
}

func TestFormatTiles(t *testing.T) {
	tests := []struct {
		tiles []game.Tile
		want  string
	}{
		{nil, "[]"},
		{[]game.Tile{{Color: game.Blue, Number: 1}}, "[blue 1]"},
		{[]game.Tile{{Color: game.Blue, Number: 1}, {Color: game.Orange, Number: 13}}, "[blue 1, orange 13]"},
	}
	for _, tt := range tests {
		if got := formatTiles(tt.tiles); got != tt.want {
			t.Errorf("formatTiles() = %q, want %q", got, tt.want)
		}
	}
}
