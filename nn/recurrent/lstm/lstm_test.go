// Copyright 2019 spaGO Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package lstm

import (
	"github.com/nlpodyssey/spago/mat/rand"
	"testing"

	"github.com/nlpodyssey/spago/ag"
	"github.com/nlpodyssey/spago/losses"
	"github.com/nlpodyssey/spago/mat"
	"github.com/nlpodyssey/spago/mat/float"
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

	x := mat.NewDense[T](mat.WithBacking([]T{-0.8, -0.9, -0.9, 1.0}), mat.WithGrad(true))
	st := model.Next(nil, x)

	assert.InDeltaSlice(t, []T{-0.15, -0.114, -0.459, 0.691, -0.401}, st.Cell.Value().Data(), 0.005)

	assert.InDeltaSlice(t, []T{-0.13, -0.05, -0.05, 0.31, -0.09}, st.Y.Value().Data(), 0.005)

	// == Backward

	gold := mat.NewDense[T](mat.WithBacking([]T{0.57, 0.75, -0.15, 1.64, 0.45}))
	loss := losses.MSE(st.Y, gold, false)
	ag.Backward(loss)

	assert.InDeltaSlice(t, []T{0.12, -0.14, 0.03, 0.02}, x.Grad().Data(), 0.005)

	assert.InDeltaSlice(t, []T{
		-0.04, -0.05, -0.05, 0.05,
		-0.02, -0.03, -0.03, 0.03,
		0.0, 0.0, 0.0, 0.0,
		0.07, 0.08, 0.08, -0.09,
		-0.02, -0.02, -0.02, 0.02,
	}, model.WIn.Grad().Data(), 0.005)

	assert.InDeltaSlice(t, []T{
		0.05, 0.03, 0.0, -0.09, 0.02,
	}, model.BIn.Grad().Data(), 0.005)

	assert.InDeltaSlice(t, []T{
		-0.01, -0.01, -0.01, 0.01,
		-0.02, -0.02, -0.02, 0.02,
		0.0, 0.0, 0.0, 0.0,
		0.16, 0.18, 0.18, -0.2,
		-0.03, -0.03, -0.03, 0.04,
	}, model.WOut.Grad().Data(), 0.005)

	assert.InDeltaSlice(t, []T{
		0.01, 0.02, 0.0, -0.2, 0.04,
	}, model.BOut.Grad().Data(), 0.005)

	assert.InDeltaSlice(t, []T{
		0.16, 0.18, 0.18, -0.2,
		0.05, 0.06, 0.06, -0.07,
		0.0, 0.0, 0.0, 0.0,
		0.01, 0.01, 0.01, -0.01,
		0.01, 0.01, 0.01, -0.01,
	}, model.WCand.Grad().Data(), 0.005)

	assert.InDeltaSlice(t, []T{
		-0.20, -0.07, 0.0, -0.01, -0.01,
	}, model.BCand.Grad().Data(), 0.005)

	if model.WInRec.HasGrad() {
		t.Error("WInRec doesn't match the expected values")
	}

	if model.WOutRec.HasGrad() {
		t.Error("WOutRec doesn't match the expected values")
	}

	if model.WForRec.HasGrad() {
		t.Error("WForRec doesn't match the expected values")
	}

	if model.WCandRec.HasGrad() {
		t.Error("WCandRec doesn't match the expected values")
	}

	if model.WFor.HasGrad() {
		t.Error("WFor doesn't match the expected values")
	}

	if model.BFor.HasGrad() {
		t.Error("BFor doesn't match the expected values")
	}
}

//gocyclo:ignore
func TestModel_ForwardWithPrev(t *testing.T) {
	t.Run("float32", testModelForwardWithPrev[float32])
	t.Run("float64", testModelForwardWithPrev[float64])
}

func testModelForwardWithPrev[T float.DType](t *testing.T) {
	model := newTestModel[T]()

	// == Forward
	s0 := &State{
		Cell: mat.NewDense[T](mat.WithBacking([]T{0.8, -0.6, 1.0, 0.1, 0.1}), mat.WithGrad(true)),
		Y:    mat.NewDense[T](mat.WithBacking([]T{-0.2, 0.2, -0.3, -0.9, -0.8}), mat.WithGrad(true)),
	}
	x := mat.NewDense[T](mat.WithBacking([]T{-0.8, -0.9, -0.9, 1.0}), mat.WithGrad(true))
	s1 := model.Next(s0, x)

	assert.InDeltaSlice(t, []T{0.5649, -0.2888, 0.3185, 0.9031, -0.4346}, s1.Cell.Value().Data(), 0.005)
	assert.InDeltaSlice(t, []T{0.47, -0.05, 0.01, 0.48, -0.16}, s1.Y.Value().Data(), 0.005)

	// == Backward

	gold := mat.NewDense[T](mat.WithBacking([]T{0.57, 0.75, -0.15, 1.64, 0.45}))
	loss := losses.MSE(s1.Y, gold, false)
	ag.Backward(loss)

	assert.InDeltaSlice(t, []T{0.106, -0.055, 0.002, 0.058}, x.Grad().Data(), 0.005)

	assert.InDeltaSlice(t, []T{
		-0.003, -0.003, -0.003, 0.003,
		0.007, 0.007, 0.007, -0.008,
		0.001, 0.002, 0.002, -0.002,
		0.044, 0.05, 0.05, -0.055,
		-0.036, -0.041, -0.041, 0.046,
	}, model.WIn.Grad().Data(), 0.005)

	assert.InDeltaSlice(t, []T{
		0.003, -0.008, -0.002, -0.055, 0.046,
	}, model.BIn.Grad().Data(), 0.005)

	assert.InDeltaSlice(t, []T{
		0.003, 0.004, 0.004, -0.004,
		-0.027, -0.03, -0.03, 0.033,
		-0.002, -0.002, -0.002, 0.002,
		0.146, 0.164, 0.164, -0.182,
		-0.047, -0.053, -0.053, 0.059,
	}, model.WOut.Grad().Data(), 0.005)

	assert.InDeltaSlice(t, []T{
		-0.004, 0.033, 0.002, -0.182, 0.059,
	}, model.BOut.Grad().Data(), 0.005)

	assert.InDeltaSlice(t, []T{
		0.038, 0.043, 0.043, -0.048,
		0.024, 0.027, 0.027, -0.03,
		0.00, 0.00, 0.00, 0.00,
		0.005, 0.006, 0.006, -0.006,
		0.012, 0.013, 0.013, -0.015,
	}, model.WCand.Grad().Data(), 0.005)

	assert.InDeltaSlice(t, []T{
		-0.048, -0.03, 0.0, -0.006, -0.015,
	}, model.BCand.Grad().Data(), 0.005)

	assert.InDeltaSlice(t, []T{
		-0.001, 0.001, -0.001, -0.003, -0.003,
		0.002, -0.002, 0.002, 0.007, 0.007,
		0.0, 0.0, 0.001, 0.002, 0.001,
		0.011, -0.011, 0.017, 0.05, 0.044,
		-0.009, 0.009, -0.014, -0.041, -0.036,
	}, model.WInRec.Grad().Data(), 0.005)

	assert.InDeltaSlice(t, []T{
		0.001, -0.001, 0.001, 0.004, 0.003,
		-0.007, 0.007, -0.01, -0.03, -0.027,
		0.0, 0.0, -0.001, -0.002, -0.002,
		0.036, -0.036, 0.055, 0.164, 0.146,
		-0.012, 0.012, -0.018, -0.053, -0.047,
	}, model.WOutRec.Grad().Data(), 0.005)

	assert.InDeltaSlice(t, []T{
		0.001, -0.001, 0.001, 0.004, 0.004,
		-0.004, 0.004, -0.006, -0.017, -0.015,
		0.0, 0.0, 0.0, -0.001, -0.001,
		0.001, -0.001, 0.001, 0.003, 0.003,
		0.001, -0.001, 0.001, 0.004, 0.004,
	}, model.WForRec.Grad().Data(), 0.005)

	assert.InDeltaSlice(t, []T{
		0.01, -0.01, 0.014, 0.043, 0.038,
		0.006, -0.006, 0.009, 0.027, 0.024,
		0.0, 0.0, 0.0, 0.0, 0.0,
		0.001, -0.001, 0.002, 0.006, 0.005,
		0.003, -0.003, 0.004, 0.013, 0.012,
	}, model.WCandRec.Grad().Data(), 0.005)

	assert.InDeltaSlice(t, []T{
		0.004, 0.004, 0.004, -0.005,
		-0.015, -0.017, -0.017, 0.019,
		-0.001, -0.001, -0.001, 0.001,
		0.003, 0.003, 0.003, -0.003,
		0.004, 0.004, 0.004, -0.005,
	}, model.WFor.Grad().Data(), 0.005)

	assert.InDeltaSlice(t, []T{
		-0.005, 0.019, 0.001, -0.003, -0.005,
	}, model.BFor.Grad().Data(), 0.005)
}

func newTestModel[T float.DType]() *Model {
	model := New[T](4, 5)
	mat.SetData[T](model.WIn.Value(), []T{
		0.5, 0.6, -0.8, -0.6,
		0.7, -0.4, 0.1, -0.8,
		0.7, -0.7, 0.3, 0.5,
		0.8, -0.9, 0.0, -0.1,
		0.4, 1.0, -0.7, 0.8,
	})
	mat.SetData[T](model.WInRec.Value(), []T{
		0.0, 0.8, 0.8, -1.0, -0.7,
		-0.7, -0.8, 0.2, -0.7, 0.7,
		-0.9, 0.9, 0.7, -0.5, 0.5,
		0.0, -0.1, 0.5, -0.2, -0.8,
		-0.6, 0.6, 0.8, -0.1, -0.3,
	})
	mat.SetData[T](model.BIn.Value(), []T{0.4, 0.0, -0.3, 0.8, -0.4})
	mat.SetData[T](model.WOut.Value(), []T{
		0.1, 0.4, -1.0, 0.4,
		0.7, -0.2, 0.1, 0.0,
		0.7, 0.8, -0.5, -0.3,
		-0.9, 0.9, -0.3, -0.3,
		-0.7, 0.6, -0.6, -0.8,
	})
	mat.SetData[T](model.WOutRec.Value(), []T{
		0.1, -0.6, -1.0, -0.1, -0.4,
		0.5, -0.9, 0.0, 0.8, 0.3,
		-0.3, -0.9, 0.3, 1.0, -0.2,
		0.7, 0.2, 0.3, -0.4, -0.6,
		-0.2, 0.5, -0.2, -0.9, 0.4,
	})
	mat.SetData[T](model.BOut.Value(), []T{0.9, 0.2, -0.9, 0.2, -0.9})
	mat.SetData[T](model.WFor.Value(), []T{
		-1.0, 0.2, 0.0, 0.2,
		-0.7, 0.7, -0.3, -0.3,
		0.3, -0.6, 0.0, 0.7,
		-1.0, -0.6, 0.9, 0.8,
		0.5, 0.8, -0.9, -0.8,
	})
	mat.SetData[T](model.WForRec.Value(), []T{
		0.2, -0.3, -0.3, -0.5, -0.7,
		0.4, -0.1, -0.6, -0.4, -0.8,
		0.6, 0.6, 0.1, 0.7, -0.4,
		-0.8, 0.9, 0.1, -0.1, -0.2,
		-0.5, -0.3, -0.6, -0.6, 0.1,
	})
	mat.SetData[T](model.BFor.Value(), []T{0.5, -0.5, 1.0, 0.4, 0.9})
	mat.SetData[T](model.WCand.Value(), []T{
		0.2, 0.6, 0.0, 0.1,
		0.1, -0.3, -0.8, -0.5,
		-0.1, 0.0, 0.4, -0.4,
		-0.8, -0.3, -0.7, 0.3,
		-0.4, 0.9, 0.8, -0.3,
	})
	mat.SetData[T](model.WCandRec.Value(), []T{
		-0.3, 0.3, -0.1, 0.6, -0.7,
		-0.2, -0.8, -0.6, -0.5, -0.4,
		-0.4, 0.8, -0.5, -0.1, 0.9,
		0.3, 0.7, 0.3, 0.0, -0.4,
		-0.3, 0.3, -0.7, 0.0, 0.7,
	})
	mat.SetData[T](model.BCand.Value(), []T{0.2, -0.9, -0.9, 0.5, 0.1})
	return model
}

//gocyclo:ignore
func TestModel_ForwardSeq(t *testing.T) {
	t.Run("float32", testModelForwardSeq[float32])
	t.Run("float64", testModelForwardSeq[float64])
}

func testModelForwardSeq[T float.DType](t *testing.T) {
	model := newTestModel2[T]()

	// == Forward
	s0 := &State{
		Cell: mat.NewDense[T](mat.WithBacking([]T{0.0, 0.0}), mat.WithGrad(true)),
		Y:    mat.NewDense[T](mat.WithBacking([]T{0.0, 0.0}), mat.WithGrad(true)),
	}
	x := mat.NewDense[T](mat.WithBacking([]T{3.5, 4.0, -0.1}), mat.WithGrad(true))
	s1 := model.Next(s0, x)

	assert.InDeltaSlice(t, []T{-0.07229, 0.97534}, s1.Cell.Value().Data(), 1.0e-05)
	assert.InDeltaSlice(t, []T{-0.00568, 0.64450}, s1.Y.Value().Data(), 1.0e-05)

	x2 := mat.NewDense[T](mat.WithBacking([]T{3.3, -2.0, 0.1}), mat.WithGrad(true))
	s2 := model.Next(s1, x2)

	assert.InDeltaSlice(t, []T{0.39238, 0.99174}, s2.Cell.Value().Data(), 1.0e-05)
	assert.InDeltaSlice(t, []T{0.01688, 0.57555}, s2.Y.Value().Data(), 1.0e-05)

	// == Backward

	s1.Y.AccGrad(mat.NewDense[T](mat.WithBacking([]T{-0.045417243, 0.363337947})))
	s2.Y.AccGrad(mat.NewDense[T](mat.WithBacking([]T{-0.043997875, 0.351983003})))

	ag.Backward(s2.Y)

	assert.InDeltaSlice(t, []T{0.017677422, 0.001052328, -0.013964347}, x.Grad().Data(), 1.0e-05)
	assert.InDeltaSlice(t, []T{0.073384228, 0.058574837, -0.065843263}, x2.Grad().Data(), 1.0e-05)

	assert.InDeltaSlice(t, []T{
		-0.0007430347, 0.0013851366, -5.39851412321307e-005,
		0.0261676128, 0.0125034523, -0.000161823,
	}, model.WIn.Grad().Data(), 1.0e-05)

	assert.InDeltaSlice(t, []T{
		-0.0002344176, 0.0076487617,
	}, model.BIn.Grad().Data(), 1.0e-05)

	assert.InDeltaSlice(t, []T{
		-0.0020964178, 0.0016978467, -7.79118487459307e-005,
		0.2587774428, 0.0141980025, 0.0020842003,
	}, model.WOut.Grad().Data(), 1.0e-05)

	assert.InDeltaSlice(t, []T{
		-0.0006395087, 0.0767240127,
	}, model.BOut.Grad().Data(), 1.0e-05)

	assert.InDeltaSlice(t, []T{
		-0.0009563009, -0.0002035453, -2.61630702758865e-006,
		0.3072443936, -0.1849936394, 0.0092695324,
	}, model.WCand.Grad().Data(), 1.0e-05)

	assert.InDeltaSlice(t, []T{
		-0.0002820345, 0.0930923312,
	}, model.BCand.Grad().Data(), 1.0e-05)

	assert.InDeltaSlice(t, []T{
		0.000002199, -0.000249511,
		-0.000017127, 0.0019433607,
	}, model.WInRec.Grad().Data(), 1.0e-05)

	assert.InDeltaSlice(t, []T{
		0.000004029, -0.0004571578,
		-0.000277093, 0.0314410032,
	}, model.WOutRec.Grad().Data(), 1.0e-05)

	assert.InDeltaSlice(t, []T{
		-1.25423475053617e-007, 1.42314671841757e-005,
		-0.0001255625, 0.0142472443,
	}, model.WForRec.Grad().Data(), 1.0e-05)

	assert.InDeltaSlice(t, []T{
		8.7529846089037e-007, -9.931778175653e-005,
		-0.0005276474, 0.0598707467,
	}, model.WCandRec.Grad().Data(), 1.0e-05)

	assert.InDeltaSlice(t, []T{
		7.28678234345139e-005, -4.41623172330387e-005, 2.20811586165194e-006,
		0.0729486052, -0.0442112759, 0.0022105638,
	}, model.WFor.Grad().Data(), 1.0e-05)

	assert.InDeltaSlice(t, []T{
		2.20811586165194e-005, 0.0221056379,
	}, model.BFor.Grad().Data(), 1.0e-05)
}

func newTestModel2[T float.DType]() *Model {
	model := New[T](3, 2)
	mat.SetData[T](model.WIn.Value(), []T{
		-0.2, -0.3, 0.5,
		0.8, 0.2, 0.01,
	})
	mat.SetData[T](model.WInRec.Value(), []T{
		0.5, 0.3,
		0.2, -0.1,
	})
	mat.SetData[T](model.BIn.Value(), []T{-0.2, 0.1})
	mat.SetData[T](model.WOut.Value(), []T{
		-0.7, 0.2, 0.1,
		0.5, 0.0, -0.5,
	})
	mat.SetData[T](model.WOutRec.Value(), []T{
		0.2, 0.7,
		0.1, -0.7,
	})
	mat.SetData[T](model.BOut.Value(), []T{-0.8, 0.0})
	mat.SetData[T](model.WFor.Value(), []T{
		0.3, 0.2, -0.4,
		0.4, 0.1, -0.6,
	})
	mat.SetData[T](model.WForRec.Value(), []T{
		-0.5, 0.22,
		0.8, -0.6,
	})
	mat.SetData[T](model.BFor.Value(), []T{0.5, 0.3})
	mat.SetData[T](model.WCand.Value(), []T{
		-0.001, -0.3, 0.5,
		0.4, 0.6, -0.3,
	})
	mat.SetData[T](model.WCandRec.Value(), []T{
		0.2, 0.7,
		0.1, -0.1,
	})
	mat.SetData[T](model.BCand.Value(), []T{0.4, 0.3})
	return model
}

func testModelInit[T float.DType](t *testing.T) {
	model := New[T](3, 2).Init(rand.NewLockedRand(42))

	assert.InDeltaSlice(t, []T{
		1.0, 1.0,
	}, model.BFor.Value().Data(), 1.0e-05)

	// TODO: add tests for the other weights and biases
}

//gocyclo:ignore
func TestModel_Init(t *testing.T) {
	t.Run("float32", testModelInit[float32])
	t.Run("float64", testModelInit[float64])
}
