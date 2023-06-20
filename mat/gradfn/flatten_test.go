// Copyright 2019 spaGO Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gradfn

import (
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/nlpodyssey/spago/mat"
	"github.com/nlpodyssey/spago/mat/float"
)

func TestFlatten_Forward(t *testing.T) {
	t.Run("float32", testFlattenForward[float32])
	t.Run("float64", testFlattenForward[float64])
}

func testFlattenForward[T float.DType](t *testing.T) {
	x := mat.NewDense[T](mat.WithShape(3, 4), mat.WithBacking([]T{
		0.1, 0.2, 0.3, 0.0,
		0.4, 0.5, -0.6, 0.7,
		-0.5, 0.8, -0.8, -0.1,
	}), mat.WithGrad(true))

	f := NewFlatten(x)
	assert.Equal(t, []mat.Tensor{x}, f.Operands())

	y, err := f.Forward()
	assert.Nil(t, err)

	assert.InDeltaSlice(t, []T{
		0.1, 0.2, 0.3,
		0.0, 0.4, 0.5,
		-0.6, 0.7, -0.5,
		0.8, -0.8, -0.1,
	}, y.Data(), 1.0e-6)

	if y.Shape()[0] != 1 || y.Shape()[1] != 12 {
		t.Error("The rows and columns of the resulting matrix are not correct")
	}

	err = f.Backward(mat.NewDense[T](mat.WithBacking([]T{
		0.1, 0.2, 0.3,
		0.0, 0.4, 0.5,
		-0.6, 0.7, -0.5,
		0.8, -0.8, -0.1})))
	assert.Nil(t, err)

	assert.InDeltaSlice(t, []T{
		0.1, 0.2, 0.3,
		0.0, 0.4, 0.5,
		-0.6, 0.7, -0.5,
		0.8, -0.8, -0.1,
	}, x.Grad().Data(), 1.0e-6)

	if x.Grad().Shape()[0] != 3 || x.Grad().Shape()[1] != 4 {
		t.Error("The rows and columns of the resulting x-gradients matrix are not correct")
	}
}
