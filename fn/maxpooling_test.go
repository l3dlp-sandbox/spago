// Copyright 2019 spaGO Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fn

import (
	"testing"

	"github.com/nlpodyssey/spago/mat"
	"github.com/nlpodyssey/spago/mat/float"
	"github.com/stretchr/testify/assert"
)

func TestMaxPool_Forward(t *testing.T) {
	t.Run("float32", testMaxPoolForward[float32])
	t.Run("float64", testMaxPoolForward[float64])
}

func testMaxPoolForward[T float.DType](t *testing.T) {
	x := &variable{
		value: mat.NewDense(4, 4, []T{
			0.4, 0.1, -0.9, -0.5,
			-0.4, 0.3, 0.7, -0.3,
			0.8, 0.2, 0.6, 0.7,
			0.2, -0.1, 0.6, -0.2,
		}),
		grad:         nil,
		requiresGrad: true,
	}
	f := NewMaxPooling(x, 2, 2)
	assert.Equal(t, []*variable{x}, f.Operands())

	y := f.Forward()

	assert.InDeltaSlice(t, []T{
		0.4, 0.7,
		0.8, 0.7,
	}, y.Data(), 1.0e-6)

	if y.Rows() != 2 || y.Columns() != 2 {
		t.Error("The rows and columns of the resulting matrix are not correct")
	}

	f.Backward(mat.NewDense(2, 2, []T{
		0.5, -0.7,
		0.8, -0.7,
	}))

	assert.InDeltaSlice(t, []T{
		0.5, 0.0, 0.0, 0.0,
		0.0, 0.0, -0.7, 0.0,
		0.8, 0.0, 0.0, -0.7,
		0.0, 0.0, 0.0, 0.0,
	}, x.grad.Data(), 1.0e-6)

	if x.grad.Rows() != 4 || x.grad.Columns() != 4 {
		t.Error("The rows and columns of the resulting x-gradients matrix are not correct")
	}
}
