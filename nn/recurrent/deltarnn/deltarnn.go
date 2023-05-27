// Copyright 2019 spaGO Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package deltarnn

import (
	"encoding/gob"

	"github.com/nlpodyssey/spago/ag"
	"github.com/nlpodyssey/spago/mat"
	"github.com/nlpodyssey/spago/mat/float"
	"github.com/nlpodyssey/spago/nn"
)

var _ nn.Model = &Model{}

// Model contains the serializable parameters.
type Model struct {
	nn.Module
	W     *nn.Param
	WRec  *nn.Param
	B     *nn.Param
	BPart *nn.Param
	Alpha *nn.Param
	Beta1 *nn.Param
	Beta2 *nn.Param
}

func init() {
	gob.Register(&Model{})
}

// State represent a state of the DeltaRNN recurrent network.
type State struct {
	D1 ag.DualValue
	D2 ag.DualValue
	C  ag.DualValue
	P  ag.DualValue
	Y  ag.DualValue
}

// New returns a new model with parameters initialized to zeros.
func New[T float.DType](in, out int) *Model {
	return &Model{
		W:     nn.NewParam(mat.NewDense[T](mat.WithShape(out, in))),
		WRec:  nn.NewParam(mat.NewDense[T](mat.WithShape(out, out))),
		B:     nn.NewParam(mat.NewDense[T](mat.WithShape(out))),
		BPart: nn.NewParam(mat.NewDense[T](mat.WithShape(out))),
		Alpha: nn.NewParam(mat.NewDense[T](mat.WithShape(out))),
		Beta1: nn.NewParam(mat.NewDense[T](mat.WithShape(out))),
		Beta2: nn.NewParam(mat.NewDense[T](mat.WithShape(out))),
	}
}

// Forward performs the forward step for each input node and returns the result.
func (m *Model) Forward(xs ...ag.DualValue) []ag.DualValue {
	ys := make([]ag.DualValue, len(xs))
	var s *State = nil
	for i, x := range xs {
		s = m.Next(s, x)
		ys[i] = s.Y
	}
	return ys
}

// Next performs a single forward step, producing a new state.
//
// d1 = beta1 * w (dot) x + beta2 * wRec (dot) yPrev
// d2 = alpha * w (dot) x * wRec (dot) yPrev
// c = tanh(d1 + d2 + bc)
// p = sigmoid(w (dot) x + bp)
// y = f(p * c + (1 - p) * yPrev)
func (m *Model) Next(state *State, x ag.DualValue) (s *State) {
	s = new(State)

	var yPrev ag.DualValue = nil
	if state != nil {
		yPrev = state.Y
	}

	wx := ag.Mul(m.W, x)
	if yPrev == nil {
		s.D1 = ag.Prod(m.Beta1, wx)
		s.C = ag.Tanh(ag.Add(s.D1, m.B))
		s.P = ag.Sigmoid(ag.Add(wx, m.BPart))
		s.Y = ag.Tanh(ag.Prod(s.P, s.C))
		return
	}
	wyRec := ag.Mul(m.WRec, yPrev)
	s.D1 = ag.Add(ag.Prod(m.Beta1, wx), ag.Prod(m.Beta2, wyRec))
	s.D2 = ag.Prod(ag.Prod(m.Alpha, wx), wyRec)
	s.C = ag.Tanh(ag.Add(ag.Add(s.D1, s.D2), m.B))
	s.P = ag.Sigmoid(ag.Add(wx, m.BPart))
	one := s.P.Value().NewScalar(1)
	s.Y = ag.Tanh(ag.Add(ag.Prod(s.P, s.C), ag.Prod(ag.ReverseSub(s.P, one), yPrev)))
	return
}
