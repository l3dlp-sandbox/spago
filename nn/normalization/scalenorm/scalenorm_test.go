// Copyright 2019 spaGO Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package scalenorm

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
	model := newTestModel[T]()

	// == Forward
	x1 := mat.NewVecDense([]T{1.0, 2.0, 0.0, 4.0}, mat.WithGrad(true))
	x2 := mat.NewVecDense([]T{3.0, 2.0, 1.0, 6.0}, mat.WithGrad(true))
	x3 := mat.NewVecDense([]T{6.0, 2.0, 5.0, 1.0}, mat.WithGrad(true))
	y := model.Forward(x1, x2, x3)

	assert.InDeltaSlice(t, []T{0.1091089451, -0.0872871560, 0.0, 0.6982972487}, y[0].Value().Data(), 1.0e-06)
	assert.InDeltaSlice(t, []T{0.2121320343, -0.0565685424, 0.0424264068, 0.6788225099}, y[1].Value().Data(), 1.0e-06)
	assert.InDeltaSlice(t, []T{0.3692744729, -0.0492365963, 0.1846372364, 0.0984731927}, y[2].Value().Data(), 1.0e-06)

	// == Backward
	y[0].AccGrad(mat.NewVecDense([]T{-1.0, -0.2, 0.4, 0.6}))
	y[1].AccGrad(mat.NewVecDense([]T{-0.3, 0.1, 0.7, 0.9}))
	y[2].AccGrad(mat.NewVecDense([]T{0.3, -0.4, 0.7, -0.8}))
	ag.BackwardMany(y...)

	assert.InDeltaSlice(t, []T{-0.1246959373, -0.0224452687, 0.0261861468, 0.0423966187}, x1.Grad().Data(), 1.0e-06)
	assert.InDeltaSlice(t, []T{-0.0554937402, -0.0256821183, 0.0182716392, 0.033262303}, x2.Grad().Data(), 1.0e-06)
	assert.InDeltaSlice(t, []T{0.0020142244, 0.0043641529, 0.0121412971, -0.0815201374}, x3.Grad().Data(), 1.0e-06)
}

func newTestModel[T float.DType]() *Model {
	model := New[T](4)
	mat.SetData[T](model.Gain.Value(), []T{0.5, -0.2, 0.3, 0.8})
	return model
}
