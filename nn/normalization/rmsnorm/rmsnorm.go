// Copyright 2019 spaGO Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package rmsnorm implements the Root Mean Square Layer Normalization method.
//
// Reference: "Root Mean Square Layer Normalization" by Biao Zhang and Rico Sennrich (2019).
// (https://arxiv.org/pdf/1910.07467.pdf)
package rmsnorm

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
	W *nn.Param
	B *nn.Param
}

func init() {
	gob.Register(&Model{})
}

// New returns a new model with parameters initialized to zeros.
func New[T float.DType](size int) *Model {
	return &Model{
		W: nn.NewParam(mat.NewDense[T](mat.WithShape(size))),
		B: nn.NewParam(mat.NewDense[T](mat.WithShape(size))),
	}
}

// Forward performs the forward step for each input node and returns the result.
func (m *Model) Forward(xs ...ag.DualValue) []ag.DualValue {
	if len(xs) == 0 {
		return nil
	}
	eps := xs[0].Value().NewScalar(1e-10)
	ys := make([]ag.DualValue, len(xs))
	for i, x := range xs {
		rms := ag.Sqrt(ag.ReduceMean(ag.Square(x)))
		ys[i] = ag.Add(ag.Prod(ag.DivScalar(x, ag.AddScalar(rms, eps)), m.W), m.B)
	}
	return ys
}
