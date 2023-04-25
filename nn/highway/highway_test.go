// Copyright 2019 spaGO Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package highway

import (
	"testing"

	"github.com/nlpodyssey/spago/ag"
	"github.com/nlpodyssey/spago/mat"
	"github.com/nlpodyssey/spago/mat/float"
	"github.com/nlpodyssey/spago/nn/activation"
	"github.com/stretchr/testify/assert"
)

func TestModel_Forward(t *testing.T) {
	t.Run("float32", testModelForward[float32])
	t.Run("float64", testModelForward[float64])
}

func testModelForward[T float.DType](t *testing.T) {
	model := newTestModel[T]()

	// == Forward

	x := mat.NewVecDense([]T{-0.8, -0.9, -0.9, 1.0}, mat.WithGrad(true))
	y := model.Forward(x)[0]

	assert.InDeltaSlice(t, []T{-0.456097, -0.855358, -0.79552, 0.844718}, y.Value().Data(), 1.0e-05)

	// == Backward

	y.AccGrad(mat.NewVecDense([]T{0.57, 0.75, -0.15, 1.64}))
	ag.Backward(y)

	assert.InDeltaSlice(t, []T{0.822396, 0.132595, -0.437002, 0.446894}, x.Grad().Data(), 1.0e-06)

	assert.InDeltaSlice(t, []T{
		-0.327765, -0.368736, -0.368736, 0.409706,
		-0.094803, -0.106653, -0.106653, 0.118504,
		0.013931, 0.015672, 0.015672, -0.017413,
		-0.346622, -0.389949, -0.389949, 0.433277,
	}, model.WIn.Grad().Data(), 1.0e-06)

	assert.InDeltaSlice(t, []T{
		0.409706, 0.118504, -0.017413, 0.433277,
	}, model.BIn.Grad().Data(), 1.0e-06)

	assert.InDeltaSlice(t, []T{
		-0.023020, -0.025897, -0.025897, 0.028775,
		-0.015190, -0.017088, -0.017088, 0.018987,
		0.011082, 0.012467, 0.012467, -0.013853,
		0.097793, 0.110017, 0.110017, -0.122241,
	}, model.WT.Grad().Data(), 1.0e-06)

	assert.InDeltaSlice(t, []T{
		0.028775, 0.018987, -0.013853, -0.122241,
	}, model.BT.Grad().Data(), 1.0e-06)
}

func newTestModel[T float.DType]() *Model {
	model := New[T](4, activation.Tanh)

	mat.SetData[T](model.WIn.Value(), []T{
		0.5, 0.6, -0.8, -0.6,
		0.7, -0.4, 0.1, -0.8,
		0.7, -0.7, 0.3, 0.5,
		0.8, -0.9, 0.0, -0.1,
	})

	mat.SetData[T](model.WT.Value(), []T{
		0.1, 0.4, -1.0, 0.4,
		0.7, -0.2, 0.1, 0.0,
		0.7, 0.8, -0.5, -0.3,
		-0.9, 0.9, -0.3, -0.3,
	})

	mat.SetData[T](model.BIn.Value(), []T{0.4, 0.0, -0.3, 0.8})
	mat.SetData[T](model.BT.Value(), []T{0.9, 0.2, -0.9, 0.2})

	return model
}
