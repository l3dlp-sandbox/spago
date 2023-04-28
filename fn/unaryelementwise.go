// Copyright 2019 spaGO Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fn

import (
	"fmt"

	"github.com/nlpodyssey/spago/mat"
)

// UnaryElementwise is a single-input element-wise function.
type UnaryElementwise[O DualValue] struct {
	x  O
	f  func(i, j int, v float64) float64 // function
	df func(i, j int, v float64) float64 // derivative
}

// Operands returns the list of operands.
func (r *UnaryElementwise[O]) Operands() []O {
	return []O{r.x}
}

// Forward computes the output of this node.
func (r *UnaryElementwise[O]) Forward() (mat.Matrix, error) {
	return r.x.Value().Apply(r.f), nil
}

// Backward computes the backward pass.
func (r *UnaryElementwise[O]) Backward(gy mat.Matrix) error {
	if !mat.SameDims(r.x.Value(), gy) {
		return fmt.Errorf("fn: matrices have incompatible dimensions")
	}
	if r.x.RequiresGrad() {
		gx := r.x.Value().Apply(r.df)
		gx.ProdInPlace(gy)
		r.x.AccGrad(gx)
	}
	return nil
}
