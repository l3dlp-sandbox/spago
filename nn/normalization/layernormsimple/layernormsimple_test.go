// Copyright 2019 spaGO Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package layernormsimple

import (
	"testing"

	"github.com/nlpodyssey/spago/ag"
	"github.com/nlpodyssey/spago/mat"
	"github.com/nlpodyssey/spago/mat/float"
	"github.com/stretchr/testify/assert"
)

func TestModel_Forward(t *testing.T) {
	t.Run("float32", testModelForward[float32])
	t.Run("float64", testModelForward[float64])
}

func testModelForward[T float.DType](t *testing.T) {
	model := New()

	// == Forward
	x1 := mat.NewDense[T](mat.WithBacking([]T{1.0, 2.0, 0.0, 4.0}), mat.WithGrad(true))
	x2 := mat.NewDense[T](mat.WithBacking([]T{3.0, 2.0, 1.0, 6.0}), mat.WithGrad(true))
	x3 := mat.NewDense[T](mat.WithBacking([]T{6.0, 2.0, 5.0, 1.0}), mat.WithGrad(true))
	y := model.Forward(x1, x2, x3)

	assert.InDeltaSlice(t, []T{-0.5070925528, 0.1690308509, -1.1832159566, 1.5212776585}, y[0].Value().Data(), 1.0e-06)
	assert.InDeltaSlice(t, []T{0.0, -0.5345224838, -1.0690449676, 1.6035674515}, y[1].Value().Data(), 1.0e-06)
	assert.InDeltaSlice(t, []T{1.2126781252, -0.7276068751, 0.7276068751, -1.2126781252}, y[2].Value().Data(), 1.0e-06)

	// == Backward

	y[0].AccGrad(mat.NewDense[T](mat.WithBacking([]T{-1.0, -0.2, 0.4, 0.6})))
	y[1].AccGrad(mat.NewDense[T](mat.WithBacking([]T{-0.3, 0.1, 0.7, 0.9})))
	y[2].AccGrad(mat.NewDense[T](mat.WithBacking([]T{0.3, -0.4, 0.7, -0.8})))
	ag.Backward(y...)

	assert.InDeltaSlice(t, []T{-0.5640800969, -0.1274975561, 0.4868088507, 0.2047688023}, x1.Grad().Data(), 1.0e-06)
	assert.InDeltaSlice(t, []T{-0.3474396144, -0.0878144080, 0.2787152951, 0.1565387274}, x2.Grad().Data(), 1.0e-06)
	assert.InDeltaSlice(t, []T{-0.1440946948, 0.0185468419, 0.1754816581, -0.0499338051}, x3.Grad().Data(), 1.0e-06)
}
