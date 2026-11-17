// Package main lays terminal panes out on the two halves of a bean.
//
// One dynamic program decides three things together: which panes go on the
// front half and which on the back, how each half is cut up, and how big each
// pane is. A pane's weight says how big it should be next to the others; the
// program maximises the smallest share divided by its weight, so a pane used
// more gets more room. The answer is a table with one entry per half shape,
// so resizing the bean is a lookup and only a change of weights is a re-solve.
package main

import (
	"math"

	"github.com/vinodhalaharvi/collage/engine"
	"github.com/vinodhalaharvi/dyna"
)

const (
	buckets   = 21 // shapes from 1:8 to 8:1
	maxAspect = 8.0
	loTarget  = 0.7 // half shapes worth an entry in the table
	hiTarget  = 6.5
	fitPower  = 2.0 // how much a half minds being stretched to its shape
)

const (
	cut  = iota // a collage step: a pane, or two groups side by side or stacked
	half        // these panes are one half, stretched to shape T
	deal        // front half and back half
)

type label struct {
	engine.Step
	Kind uint8
	T    int
}

type state struct {
	M uint32
	F uint8 // 0: a group of panes; 1: a half; 2: the whole bean
}

func coalgebra(targets []int) dyna.Coalgebra[state, label] {
	splits := engine.Splits()
	return func(s state) []dyna.Node[label, state] {
		var out []dyna.Node[label, state]
		switch s.F {
		case 0:
			for _, n := range splits(s.M) {
				kids := make([]state, len(n.Kids))
				for i, k := range n.Kids {
					kids[i] = state{M: k}
				}
				out = append(out, dyna.Node[label, state]{Label: label{Step: n.Label}, Kids: kids})
			}
		case 1:
			for _, t := range targets {
				out = append(out, dyna.Node[label, state]{Label: label{Kind: half, T: t}, Kids: []state{{M: s.M}}})
			}
		default:
			// the half holding the lowest pane is the front, so the same deal
			// the other way round is not counted again
			low := s.M & -s.M
			rest := s.M ^ low
			for sub := rest; ; sub = (sub - 1) & rest {
				if b := rest ^ sub; b != 0 {
					out = append(out, dyna.Node[label, state]{Label: label{Kind: deal}, Kids: []state{{M: low | sub, F: 1}, {M: b, F: 1}}})
				}
				if sub == 0 {
					break
				}
			}
		}
		return out
	}
}

// Tree is a half's layout: a pane, or a cut with the share its first side takes.
type Tree struct {
	P int     `json:"p"`
	C string  `json:"c,omitempty"`
	F float64 `json:"f,omitempty"`
	L *Tree   `json:"l,omitempty"`
	R *Tree   `json:"r,omitempty"`
	a float64
}

// Entry is the best bean for one half shape.
type Entry struct {
	Aspect float64 `json:"aspect"`
	Score  float64 `json:"score"`
	Front  *Tree   `json:"front"`
	Back   *Tree   `json:"back"`
}

// Result is the whole table.
type Result struct {
	States       int     `json:"states"`
	Alternatives int     `json:"alternatives"`
	Table        []Entry `json:"table"`
}

type solver struct {
	aspects, weights []float64
	g                *engine.Grades
	targets          []int
}

func newSolver(aspects, weights []float64) *solver {
	s := &solver{aspects: aspects, weights: weights, g: engine.NewGrades(buckets, maxAspect)}
	for b := 0; b < buckets; b++ {
		if c := s.g.Center(b); c >= loTarget && c <= hiTarget {
			s.targets = append(s.targets, b)
		}
	}
	return s
}

func (s *solver) algebra() dyna.Algebra[label, dyna.Vec[dyna.Ranked[label]]] {
	base := s.g.Grading(s.aspects)
	grading := dyna.Grading[label]{N: s.g.B, Value: s.g.Center, Of: func(l label, kids []int) int {
		switch l.Kind {
		case half:
			return l.T
		case deal:
			if kids[0] != kids[1] {
				return -1
			}
			return kids[0]
		}
		return base.Of(l.Step, kids)
	}}
	score := func(l label, kids []float64, _ float64, grades []float64) float64 {
		switch {
		case l.Kind == half:
			t := s.g.Center(l.T)
			fit := math.Min(grades[0]/t, t/grades[0])
			return kids[0] * math.Pow(fit, fitPower)
		case l.Kind == deal:
			return math.Min(kids[0], kids[1]) // the bean is as good as its worse half
		case l.Leaf:
			return 1 / s.weights[l.Photo]
		}
		f := engine.FracL(l.Cut, grades[0], grades[1])
		return math.Min(kids[0]*f, kids[1]*(1-f))
	}
	return dyna.Lift(dyna.TopK(1, score), grading)
}

func (s *solver) root() state { return state{M: uint32(1)<<len(s.aspects) - 1, F: 2} }

func (s *solver) tree(e *dyna.Entry[label]) *Tree {
	if e.Label.Leaf {
		return &Tree{P: e.Label.Photo, a: s.aspects[e.Label.Photo]}
	}
	l, r := s.tree(e.Kids[0]), s.tree(e.Kids[1])
	return &Tree{P: -1, C: e.Label.Cut.String(), F: math.Round(engine.FracL(e.Label.Cut, l.a, r.a)*1e4) / 1e4,
		L: l, R: r, a: engine.Aspect(e.Label.Cut, l.a, r.a)}
}

func solve(aspects, weights []float64) Result {
	s := newSolver(aspects, weights)
	root, st := dyna.Dyna(s.algebra(), coalgebra(s.targets), s.root())
	res := Result{States: st.States, Alternatives: st.Alternatives}
	for _, t := range s.targets {
		if root.At == nil || len(root.At[t]) == 0 {
			continue
		}
		e := root.At[t][0]
		res.Table = append(res.Table, Entry{Aspect: s.g.Center(t), Score: e.Score,
			Front: s.tree(e.Kids[0].Kids[0]), Back: s.tree(e.Kids[1].Kids[0])})
	}
	return res
}
