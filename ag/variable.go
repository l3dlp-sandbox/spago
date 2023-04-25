// Copyright 2019 spaGO Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ag

import (
	"sync"

	"github.com/nlpodyssey/spago/mat"
	"github.com/nlpodyssey/spago/mat/float"
)

var _ Node = &Variable{}

// Variable is a simple type of Node, primarily consisting of a value and
// optional gradients.
type Variable struct {
	value        mat.Matrix
	grad         mat.Matrix
	gradMu       sync.RWMutex
	requiresGrad bool
}

// Var creates a new Variable Node.
// Use WithGrad() to set whether the variable requires gradients (default false).
func Var(value mat.Matrix) *Variable {
	return &Variable{
		value:        value,
		grad:         nil,
		requiresGrad: false,
	}
}

// Scalar creates a new Variable from a scalar value.
// Use WithGrad() to set whether the variable requires gradients (default false).
func Scalar[T float.DType](value T) *Variable {
	return &Variable{
		value:        mat.NewScalar(value),
		grad:         nil,
		requiresGrad: false,
	}
}

// WithGrad sets whether the variable requires gradients.
func (r *Variable) WithGrad(value bool) *Variable {
	r.requiresGrad = value
	return r
}

// Value returns the value of the variable itself.
func (r *Variable) Value() mat.Matrix {
	return r.value
}

// Grad returns the gradients accumulated during the backward pass.
func (r *Variable) Grad() mat.Matrix {
	r.gradMu.RLock()
	defer r.gradMu.RUnlock()
	return r.grad
}

// AccGrad accumulates the gradients into the Variable.
func (r *Variable) AccGrad(grad mat.Matrix) {
	if !r.requiresGrad {
		return
	}
	r.gradMu.Lock()
	defer r.gradMu.Unlock()
	if r.grad == nil {
		r.grad = grad.Clone()
		return
	}
	r.grad.AddInPlace(grad)
}

// HasGrad reports whether there are accumulated gradients.
func (r *Variable) HasGrad() bool {
	return r.Grad() != nil
}

// RequiresGrad reports whether the Variable requires gradients.
func (r *Variable) RequiresGrad() bool {
	return r.requiresGrad
}

// ZeroGrad zeroes the gradients, setting the value of Grad to nil.
func (r *Variable) ZeroGrad() {
	if r.Grad() == nil {
		return
	}
	r.gradMu.Lock()
	defer r.gradMu.Unlock()
	mat.ReleaseMatrix(r.grad)
	r.grad = nil
}
