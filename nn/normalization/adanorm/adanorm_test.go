// Copyright 2019 spaGO Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package adanorm

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
	m := New[T](0.8)

	// == Forward
	x1 := mat.NewDense[T](mat.WithBacking([]T{1.0, 2.0, 0.0, 4.0}), mat.WithGrad(true))
	x2 := mat.NewDense[T](mat.WithBacking([]T{3.0, 2.0, 1.0, 6.0}), mat.WithGrad(true))
	x3 := mat.NewDense[T](mat.WithBacking([]T{6.0, 2.0, 5.0, 1.0}), mat.WithGrad(true))
	y := m.Forward(x1, x2, x3)

	assert.InDeltaSlice(t, []T{-0.4262454708, 0.1329389665, -1.0585727653, 1.0318792697}, y[0].Value().Data(), 1.0e-06)
	assert.InDeltaSlice(t, []T{0, -0.4504751299, -0.9466645455, 1.0771396755}, y[1].Value().Data(), 1.0e-06)
	assert.InDeltaSlice(t, []T{0.8524954413, -0.6244384413, 0.5397325589, -1.087789559}, y[2].Value().Data(), 1.0e-06)

	// == Backward
	y[0].AccGrad(mat.NewDense[T](mat.WithBacking([]T{-1.0, -0.2, 0.4, 0.6})))
	y[1].AccGrad(mat.NewDense[T](mat.WithBacking([]T{-0.3, 0.1, 0.7, 0.9})))
	y[2].AccGrad(mat.NewDense[T](mat.WithBacking([]T{0.3, -0.4, 0.7, -0.8})))
	ag.Backward(y...)

	assert.InDeltaSlice(t, []T{-0.4779089755, -0.0839735551, 0.4004185091, 0.1614640214}, x1.Grad().Data(), 1.0e-06)
	assert.InDeltaSlice(t, []T{-0.2710945487, -0.0790678529, 0.2259110116, 0.12425139}, x2.Grad().Data(), 1.0e-06)
	assert.InDeltaSlice(t, []T{-0.1154695275, 0.0283184423, 0.1372573, -0.050106214}, x3.Grad().Data(), 1.0e-06)
}
