// Copyright 2021 spaGO Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package convolution2d

import (
	"testing"

	"github.com/nlpodyssey/spago/ag"
	"github.com/nlpodyssey/spago/mat"
	"github.com/nlpodyssey/spago/mat/float"
	"github.com/nlpodyssey/spago/nn/activation"
	"github.com/stretchr/testify/assert"
)

//gocyclo:ignore
func TestModel_Forward(t *testing.T) {
	t.Run("float32", testModelForward[float32])
	t.Run("float64", testModelForward[float64])
}

func testModelForward[T float.DType](t *testing.T) {
	model := newTestModel[T]()

	// == Forward

	x1 := mat.NewDense(4, 4, []T{
		0.2, 0.1, 0.5, 0.8,
		0.4, -0.3, -0.2, -0.3,
		0.5, -0.6, -0.4, 0.6,
		-0.3, 0.9, 0.5, 0.5,
	}, mat.WithGrad(true))
	x2 := mat.NewDense(4, 4, []T{
		-0.2, 0.1, 0.5, 0.8,
		0.4, -0.3, -0.2, -0.9,
		0.5, 0.2, 0.2, 0.9,
		0.9, 0.3, 0.2, 0.7,
	}, mat.WithGrad(true))
	x3 := mat.NewDense(4, 4, []T{
		0.2, 0.5, 0.9, 0.8,
		0.4, -0.5, -0.3, -0.2,
		0.5, 0.6, -0.9, 0.0,
		0.3, 0.9, 0.2, 0.1,
	}, mat.WithGrad(true))
	y := model.Forward(x1, x2, x3)

	assert.InDeltaSlice(t, []T{
		0.6291451614, 0.4218990053, 0.0399786803,
		0.8956928738, -0.0698858903, 0.8004990218,
		0.9892435057, 0.8956928738, 0.8144140938,
	}, y[0].Value().Data(), 1.0e-05)

	assert.InDeltaSlice(t, []T{
		0.7778880666, 0.9775871874, 0.9968122755,
		0.4853810906, 0.0299910032, 0.049958375,
		0.9934620209, -0.0996679946, 0.7931990971,
	}, y[1].Value().Data(), 1.0e-05)

	y[0].AccGrad(mat.NewDense(3, 3, []T{
		-0.3, 0.5, 0.6,
		0.9, 0.1, 0.0,
		0.3, 0.4, -1.0,
	}))

	y[1].AccGrad(mat.NewDense(3, 3, []T{
		-0.3, 0.5, -0.6,
		-0.2, 0.0, 0.1,
		0.3, 0.6, 0.0,
	}))

	ag.Backward(y...)

	assert.InDeltaSlice(t, []T{
		0.4361460918, 0.3557904551,
		-0.385442345, -0.4771584238,
	}, model.K[0].Grad().Data(), 1.0e-05)

	assert.InDeltaSlice(t, []T{
		0.3698844136, 0.3073631249,
		-0.2445673659, -0.7294329628,
	}, model.K[1].Grad().Data(), 1.0e-05)

	assert.InDeltaSlice(t, []T{
		1.083537722, 0.5723401861,
		-0.3032622381, -0.1473428208,
	}, model.K[2].Grad().Data(), 1.0e-05)

	assert.InDeltaSlice(t, []T{
		0.8550443848,
	}, model.B[0].Grad().Data(), 1.0e-05)

	assert.InDeltaSlice(t, []T{
		0.8550443848,
	}, model.B[1].Grad().Data(), 1.0e-05)

	assert.InDeltaSlice(t, []T{
		0.8550443848,
	}, model.B[2].Grad().Data(), 1.0e-05)

	assert.InDeltaSlice(t, []T{
		-0.4589582127, -0.227843921,
		0.3638506439, 0.4843712647,
	}, model.K[3].Grad().Data(), 1.0e-05)

	assert.InDeltaSlice(t, []T{
		0.0636604256, 0.0718576652,
		0.0719689855, 0.2137251507,
	}, model.K[4].Grad().Data(), 1.0e-05)

	assert.InDeltaSlice(t, []T{
		0.2512514644, -0.5181427625,
		0.3122710023, 0.083947163,
	}, model.K[5].Grad().Data(), 1.0e-05)

	assert.InDeltaSlice(t, []T{
		0.4446945415,
	}, model.B[3].Grad().Data(), 1.0e-05)

	assert.InDeltaSlice(t, []T{
		0.4446945415,
	}, model.B[4].Grad().Data(), 1.0e-05)

	assert.InDeltaSlice(t, []T{
		0.4446945415,
	}, model.B[5].Grad().Data(), 1.0e-05)

	assert.InDeltaSlice(t, []T{
		-0.1380132736, 0.1920924921, 0.1513219132, -0.2426717471,
		0.0800724775, -0.1421413609, 0.3154099326, 0.2579849708,
		0.1957547684, 0.2998123793, 0.215307598, 0.17459204,
		-0.0015932117, -0.5074179426, 0.1603251177, -0.1010189052,
	}, x1.Grad().Data(), 1.0e-05)

	assert.InDeltaSlice(t, []T{
		0.0906264549, -0.3191097037, -0.1651395044, 0.1778027207,
		-0.1607711201, -0.0878535422, 0.5073356622, 0.590921715,
		-0.0134817352, 0.2208414849, 0.6085984037, -0.1508941132,
		0.0024568264, 0.1978529598, -0.2931814848, -0.3030567155,
	}, x2.Grad().Data(), 1.0e-05)

	assert.InDeltaSlice(t, []T{
		-0.1080412779, 0.0455926386, 0.3750678293, 0.1774208035,
		-0.0346239638, -0.1154286619, 0.4431526591, 0.4162195256,
		0.0087566253, 0.2184951473, 0.3251171452, -0.0212185723,
		0.00206583, 0.1416061797, 0.455342109, -0.2020378103,
	}, x3.Grad().Data(), 1.0e-05)
}

func TestDepthwise_Forward(t *testing.T) {
	t.Run("float32", testDepthwiseForward[float32])
	t.Run("float64", testDepthwiseForward[float64])
}

func testDepthwiseForward[T float.DType](t *testing.T) {
	model := newTestModel2[T]()

	// == Forward

	x1 := mat.NewDense(4, 4, []T{
		0.2, 0.1, 0.5, 0.8,
		0.4, -0.3, -0.2, -0.3,
		0.5, -0.6, -0.4, 0.6,
		-0.3, 0.9, 0.5, 0.5,
	}, mat.WithGrad(true))
	x2 := mat.NewDense(4, 4, []T{
		-0.2, 0.1, 0.5, 0.8,
		0.4, -0.3, -0.2, -0.9,
		0.5, 0.2, 0.2, 0.9,
		0.9, 0.3, 0.2, 0.7,
	}, mat.WithGrad(true))
	x3 := mat.NewDense(4, 4, []T{
		0.2, 0.5, 0.9, 0.8,
		0.4, -0.5, -0.3, -0.2,
		0.5, 0.6, -0.9, 0.0,
		0.3, 0.9, 0.2, 0.1,
	}, mat.WithGrad(true))
	y := model.Forward(x1, x2, x3)

	assert.InDeltaSlice(t, []T{
		0.09, -0.3, -0.22,
		0.29, -0.37, 0.08,
		0.67, 0.28, -0.14,
	}, y[0].Value().Data(), 1.0e-05)

	assert.InDeltaSlice(t, []T{
		0.14, 0.06, -0.66,
		0.19, 0.51, 0.88,
		0.46, 0.4, 1.04,
	}, y[1].Value().Data(), 1.0e-05)

	assert.InDeltaSlice(t, []T{
		0.51, 0.69, 0.92,
		0.97, -0.21, 0.14,
		1.48, 0.77, 0.24,
	}, y[2].Value().Data(), 1.0e-05)
}

func newTestModel[T float.DType]() *Model {
	model := New[T](Config{
		KernelSizeX:    2,
		KernelSizeY:    2,
		XStride:        1,
		YStride:        1,
		InputChannels:  3,
		OutputChannels: 2,
		Mask:           []int{1, 1, 1},
		DepthWise:      false,
		Activation:     activation.Tanh,
	})
	mat.SetData[T](model.K[0].Value(), []T{
		0.5, -0.4,
		0.3, 0.3,
	})
	mat.SetData[T](model.K[1].Value(), []T{
		-0.5, 0.3,
		0.2, 0.9,
	})
	mat.SetData[T](model.K[2].Value(), []T{
		0.4, 0.3,
		0.2, 0.6,
	})
	mat.SetData[T](model.B[0].Value(), []T{0.0})
	mat.SetData[T](model.B[1].Value(), []T{0.2})
	mat.SetData[T](model.B[2].Value(), []T{0.5})
	mat.SetData[T](model.K[3].Value(), []T{
		0.4, 0.8,
		-0.9, 0.4,
	})
	mat.SetData[T](model.K[4].Value(), []T{
		0.0, 0.5,
		0.3, -0.5,
	})
	mat.SetData[T](model.K[5].Value(), []T{
		0.3, 0.6,
		0.2, 0.8,
	})
	mat.SetData[T](model.B[3].Value(), []T{0.4})
	mat.SetData[T](model.B[4].Value(), []T{0.1})
	mat.SetData[T](model.B[5].Value(), []T{0.5})
	return model
}

func newTestModel2[T float.DType]() *Model {
	model := New[T](Config{
		KernelSizeX:    2,
		KernelSizeY:    2,
		XStride:        1,
		YStride:        1,
		InputChannels:  3,
		OutputChannels: 3,
		Mask:           []int{1, 1, 1},
		DepthWise:      true,
		Activation:     activation.Identity,
	})
	mat.SetData[T](model.K[0].Value(), []T{
		0.5, -0.4,
		0.3, 0.3,
	})
	mat.SetData[T](model.K[1].Value(), []T{
		-0.5, 0.3,
		0.2, 0.9,
	})
	mat.SetData[T](model.K[2].Value(), []T{
		0.4, 0.3,
		0.2, 0.6,
	})
	mat.SetData[T](model.B[0].Value(), []T{0.0})
	mat.SetData[T](model.B[1].Value(), []T{0.2})
	mat.SetData[T](model.B[2].Value(), []T{0.5})
	return model
}
