//go:build js && wasm

package game

import "time"

// Let the host JavaScript event loop deliver cancellation during a search.
// Desktop/native builds do not incur this scheduling delay.
func searchYield() { time.Sleep(time.Millisecond) }
