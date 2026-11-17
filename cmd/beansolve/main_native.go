//go:build !js

package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"time"

	"github.com/vinodhalaharvi/dyna"
)

func shares(t *Tree, w, h float64, out map[int]float64) {
	if t.P >= 0 {
		out[t.P] = w * h
		return
	}
	if t.C == "H" {
		shares(t.L, w*t.F, h, out)
		shares(t.R, w*(1-t.F), h, out)
	} else {
		shares(t.L, w, h*t.F, out)
		shares(t.R, w, h*(1-t.F), out)
	}
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "json" {
		// the table for ten panes used equally, to draw before the first solve finishes
		json.NewEncoder(os.Stdout).Encode(solve([]float64{1.6, 1.6, 0.9, 2.4, 1.6, 0.8, 1.2, 1.6, 1.0, 1.3}, []float64{1, 1, 1, 1, 1, 1, 1, 1, 1, 1}))
		return
	}
	// the specification: a small case against every whole bean, one by one
	as := []float64{1.6, 1.6, 0.9, 2.4, 0.8}
	ws := []float64{1, 3, 1, 1, 2}
	s := newSolver(as, ws)
	s.targets = []int{s.g.Of(1), s.g.Of(1.7), s.g.Of(2.4), s.g.Of(3.4)} // brute force is slow: check four shapes
	dp, _ := dyna.Dyna(s.algebra(), coalgebra(s.targets), s.root())
	brute, trees := dyna.BruteForce(s.algebra(), coalgebra(s.targets), s.root())
	same := true
	for _, t := range s.targets {
		a, b := dp.At[t], brute.At[t]
		if len(a) != len(b) || (len(a) > 0 && math.Abs(a[0].Score-b[0].Score) > 1e-12) {
			same = false
		}
	}
	fmt.Printf("5 panes: brute force over %d whole beans agrees at all %d shapes: %v\n", trees, len(s.targets), same)

	ten := []float64{1.6, 1.6, 0.9, 2.4, 1.6, 0.8, 1.2, 1.6, 1.0, 1.3}
	for _, w := range [][]float64{{1, 1, 1, 1, 1, 1, 1, 1, 1, 1}, {1, 1, 1, 1, 6, 1, 1, 3, 1, 1}} {
		t0 := time.Now()
		r := solve(ten, w)
		fmt.Printf("10 panes, weights %v: %d states, %d alternatives, %d ms, %d shapes\n", w, r.States, r.Alternatives, time.Since(t0).Milliseconds(), len(r.Table))
		for _, e := range r.Table {
			if e.Aspect > 2.2 && e.Aspect < 2.6 {
				f, b := map[int]float64{}, map[int]float64{}
				shares(e.Front, 1, 1, f)
				shares(e.Back, 1, 1, b)
				fmt.Printf("  shape %.2f: front %v\n               back  %v\n", e.Aspect, pct(f), pct(b))
			}
		}
	}
}

func pct(m map[int]float64) string {
	s := ""
	for p := 0; p < 12; p++ {
		if v, ok := m[p]; ok {
			s += fmt.Sprintf("%d:%.0f%% ", p, v*100)
		}
	}
	return s
}
