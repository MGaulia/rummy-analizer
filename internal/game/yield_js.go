//go:build js && wasm

package game

import "time"

// Let the worker event loop receive a cancellation while Go search is running.
// Desktop/native builds do not incur this scheduling delay.
func searchYield() { time.Sleep(time.Millisecond) }
