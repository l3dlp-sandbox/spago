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

func TestTranspose_Forward(t *testing.T) {
	t.Run("float32", testTransposeForward[float32])
	t.Run("float64", testTransposeForward[float64])
}

func testTransposeForward[T float.DType](t *testing.T) {
	x := &variable{
		value: mat.NewDense(3, 4, []T{
			0.1, 0.2, 0.3, 0.0,
			0.4, 0.5, -0.6, 0.7,
			-0.5, 0.8, -0.8, -0.1,
		}),
		grad:         nil,
		requiresGrad: true,
	}

	f := NewTranspose(x)
	assert.Equal(t, []*variable{x}, f.Operands())

	y := f.Forward()

	assert.InDeltaSlice(t, []T{
		0.1, 0.4, -0.5,
		0.2, 0.5, 0.8,
		0.3, -0.6, -0.8,
		0.0, 0.7, -0.1,
	}, y.Data(), 1.0e-6)

	if y.Rows() != 4 || y.Columns() != 3 {
		t.Error("The rows and columns of the resulting matrix are not right")
	}

	f.Backward(mat.NewDense(4, 3, []T{
		0.1, 0.2, 0.3,
		0.0, 0.4, 0.5,
		-0.6, 0.7, -0.5,
		0.8, -0.8, -0.1,
	}))

	assert.InDeltaSlice(t, []T{
		0.1, 0.0, -0.6, 0.8,
		0.2, 0.4, 0.7, -0.8,
		0.3, 0.5, -0.5, -0.1,
	}, x.grad.Data(), 1.0e-6)

	if x.grad.Rows() != 3 || x.grad.Columns() != 4 {
		t.Error("The rows and columns of the resulting x-gradients matrix are not correct")
	}
}
