//go:build js

package main

import (
	"encoding/json"
	"syscall/js"
)

func floats(v js.Value) []float64 {
	out := make([]float64, v.Length())
	for i := range out {
		out[i] = v.Index(i).Float()
	}
	return out
}

func main() {
	// beanSolve(aspects, weights) returns the table as JSON
	js.Global().Set("beanSolve", js.FuncOf(func(_ js.Value, args []js.Value) any {
		b, _ := json.Marshal(solve(floats(args[0]), floats(args[1])))
		return string(b)
	}))
	select {}
}
