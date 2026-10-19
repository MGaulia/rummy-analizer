//go:build js && wasm

// The phone web view runs this Go engine locally in a Web Worker.
package main

import (
	"context"
	"encoding/json"
	"strings"
	"sync"
	"syscall/js"

	"github.com/MGaulia/rummy-analizer/internal/app"
)

func main() {
	var mu sync.Mutex
	active := map[string]context.CancelFunc{}
	execute := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		id, raw := args[0].String(), args[1].String()
		executor := js.FuncOf(func(this js.Value, promiseArgs []js.Value) interface{} {
			resolve := promiseArgs[0]
			ctx, cancel := context.WithCancel(context.Background())
			mu.Lock()
			if previous := active[id]; previous != nil {
				previous()
			}
			active[id] = cancel
			mu.Unlock()
			go func() {
				defer cancel()
				var request app.Request
				response := app.Response{}
				if err := app.Decode(strings.NewReader(raw), &request); err != nil {
					response.Error = err.Error()
				} else {
					response = app.Execute(ctx, request)
				}
				encoded, err := json.Marshal(response)
				if err != nil {
					encoded = []byte(`{"error":"response encoding failed"}`)
				}
				mu.Lock()
				delete(active, id)
				mu.Unlock()
				resolve.Invoke(string(encoded))
			}()
			return nil
		})
		promise := js.Global().Get("Promise").New(executor)
		executor.Release()
		return promise
	})
	cancel := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		mu.Lock()
		if stop := active[args[0].String()]; stop != nil {
			stop()
		}
		mu.Unlock()
		return nil
	})
	js.Global().Set("rummyExecute", execute)
	js.Global().Set("rummyCancel", cancel)
	select {}
}
