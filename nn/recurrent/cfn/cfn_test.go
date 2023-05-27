// Copyright 2019 spaGO Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cfn

import (
	"testing"

	"github.com/nlpodyssey/spago/ag"
	"github.com/nlpodyssey/spago/losses"
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

	x := mat.NewDense[T](mat.WithBacking([]T{-0.8, -0.9, -0.9, 1.0}), mat.WithGrad(true))
	y := model.Forward(x)[0]

	assert.InDeltaSlice(t, []T{0.268, -0.025, 0.381, 0.613, -0.364}, y.Value().Data(), 0.0005)

	// == Backward

	gold := mat.NewDense[T](mat.WithBacking([]T{0.57, 0.75, -0.15, 1.64, 0.45}))
	loss := losses.MSE(y, gold, false)
	ag.Backward(loss)

	assert.InDeltaSlice(t, []T{0.318, 0.01, -0.027, 0.302}, x.Grad().Data(), 0.005)

	assert.InDeltaSlice(t, []T{
		0.039, 0.044, 0.044, -0.049,
		-0.012, -0.013, -0.013, 0.015,
		-0.081, -0.091, -0.091, 0.101,
		0.149, 0.167, 0.167, -0.186,
		-0.13, -0.146, -0.146, 0.162,
	}, model.WIn.Grad().Data(), 0.005)

	assert.InDeltaSlice(t, []T{-0.049, 0.015, 0.101, -0.186, 0.16}, model.BIn.Grad().Data(), 0.005)

	assert.InDeltaSlice(t, []T{
		0.052, 0.059, 0.059, -0.065,
		0.154, 0.174, 0.174, -0.193,
		-0.089, -0.1, -0.1, 0.111,
		0.142, 0.159, 0.159, -0.177,
		0.104, 0.117, 0.117, -0.130,
	}, model.WCand.Grad().Data(), 0.005)

	if model.WInRec.HasGrad() {
		t.Error("WInRec doesn't match the expected values")
	}
	if model.WForRec.HasGrad() {
		t.Error("WForRec doesn't match the expected values")
	}

	if model.WFor.HasGrad() {
		t.Error("WFor doesn't match the expected values")
	}
}

func TestModel_ForwardWithPrev(t *testing.T) {
	t.Run("float32", testModelForwardWithPrev[float32])
	t.Run("float64", testModelForwardWithPrev[float64])
}

func testModelForwardWithPrev[T float.DType](t *testing.T) {
	model := newTestModel[T]()

	s0 := &State{Y: mat.NewDense[T](mat.WithBacking([]T{-0.2, 0.2, -0.3, -0.9, -0.8}), mat.WithGrad(true))}

	// == Forward

	x := mat.NewDense[T](mat.WithBacking([]T{-0.8, -0.9, -0.9, 1.0}), mat.WithGrad(true))
	s1 := model.Next(s0, x)

	assert.InDeltaSlice(t, []T{0.308, 0.011, 0.405, 0.230, -0.689}, s1.Y.Value().Data(), 0.0005)

	// == Backward

	gold := mat.NewDense[T](mat.WithBacking([]T{0.57, 0.75, -0.15, 1.64, 0.45}))
	loss := losses.MSE(s1.Y, gold, false)
	ag.Backward(loss)

	assert.InDeltaSlice(t, []T{0.111, 0.37, -0.281, 0.126}, x.Grad().Data(), 0.005)

	assert.InDeltaSlice(t, []T{
		0.028, 0.032, 0.032, -0.035,
		-0.011, -0.012, -0.012, 0.014,
		-0.084, -0.094, -0.094, 0.105,
		0.144, 0.162, 0.162, -0.18,
		-0.182, -0.205, -0.205, 0.228,
	}, model.WIn.Grad().Data(), 0.005)

	assert.InDeltaSlice(t, []T{-0.035, 0.014, 0.105, -0.18, 0.228}, model.BIn.Grad().Data(), 0.005)

	assert.InDeltaSlice(t, []T{
		0.082, 0.093, 0.093, -0.103,
		0.146, 0.164, 0.164, -0.183,
		-0.102, -0.115, -0.115, 0.128,
		0.226, 0.255, 0.255, -0.283,
		0.172, 0.194, 0.194, -0.215,
	}, model.WCand.Grad().Data(), 0.005)

	assert.InDeltaSlice(t, []T{
		0.007, -0.007, 0.011, 0.032, 0.028,
		-0.003, 0.003, -0.004, -0.012, -0.011,
		-0.021, 0.021, -0.031, -0.094, -0.084,
		0.036, -0.036, 0.054, 0.162, 0.144,
		-0.046, 0.046, -0.068, -0.205, -0.182,
	}, model.WInRec.Grad().Data(), 0.005)

	assert.InDeltaSlice(t, []T{
		-0.003, -0.004, -0.004, 0.004,
		0.017, 0.019, 0.019, -0.022,
		0.006, 0.007, 0.007, -0.007,
		-0.177, -0.199, -0.199, 0.222,
		-0.144, -0.162, -0.162, 0.18,
	}, model.WFor.Grad().Data(), 0.005)

	assert.InDeltaSlice(t, []T{0.004, -0.022, -0.007, 0.222, 0.18}, model.BFor.Grad().Data(), 0.005)

	assert.InDeltaSlice(t, []T{
		-0.001, 0.001, -0.001, -0.004, -0.003,
		0.004, -0.004, 0.006, 0.019, 0.017,
		0.001, -0.001, 0.002, 0.007, 0.006,
		-0.044, 0.044, -0.066, -0.199, -0.177,
		-0.036, 0.036, -0.054, -0.162, -0.144,
	}, model.WForRec.Grad().Data(), 0.005)
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
	mat.SetData[T](model.WFor.Value(), []T{
		0.1, 0.4, -1.0, 0.4,
		0.7, -0.2, 0.1, 0.0,
		0.7, 0.8, -0.5, -0.3,
		-0.9, 0.9, -0.3, -0.3,
		-0.7, 0.6, -0.6, -0.8,
	})
	mat.SetData[T](model.WForRec.Value(), []T{
		0.1, -0.6, -1.0, -0.1, -0.4,
		0.5, -0.9, 0.0, 0.8, 0.3,
		-0.3, -0.9, 0.3, 1.0, -0.2,
		0.7, 0.2, 0.3, -0.4, -0.6,
		-0.2, 0.5, -0.2, -0.9, 0.4,
	})
	mat.SetData[T](model.BFor.Value(), []T{0.9, 0.2, -0.9, 0.2, -0.9})
	mat.SetData[T](model.WCand.Value(), []T{
		-1.0, 0.2, 0.0, 0.2,
		-0.7, 0.7, -0.3, -0.3,
		0.3, -0.6, 0.0, 0.7,
		-1.0, -0.6, 0.9, 0.8,
		0.5, 0.8, -0.9, -0.8,
	})
	return model
}

func TestModel_ForwardSeq(t *testing.T) {
	t.Run("float32", testModelForwardSeq[float32])
	t.Run("float64", testModelForwardSeq[float64])
}

func testModelForwardSeq[T float.DType](t *testing.T) {
	model := newTestModel2[T]()

	s0 := &State{Y: mat.NewDense[T](mat.WithBacking([]T{0.0, 0.0}), mat.WithGrad(true))}

	// == Forward

	x := mat.NewDense[T](mat.WithBacking([]T{3.5, 4.0, -0.1}), mat.WithGrad(true))
	s1 := model.Next(s0, x)

	assert.InDeltaSlice(t, []T{-0.0886045623, 0.9749300057}, s1.Y.Value().Data(), 1.0e-05)

	x2 := mat.NewDense[T](mat.WithBacking([]T{3.3, -2.0, 0.1}), mat.WithGrad(true))
	s2 := model.Next(s1, x2)

	assert.InDeltaSlice(t, []T{0.2205790544, 0.5834192006}, s2.Y.Value().Data(), 1.0e-05)

	// == Backward

	s1.Y.AccGrad(mat.NewDense[T](mat.WithBacking([]T{-0.0522186536, 0.4177492291})))
	s2.Y.AccGrad(mat.NewDense[T](mat.WithBacking([]T{-0.0436513876, 0.3492111007})))

	ag.Backward(s2.Y)

	assert.InDeltaSlice(t, []T{0.0087725508, 0.0021613524, 0.000922185}, x.Grad().Data(), 1.0e-05)

	assert.InDeltaSlice(t, []T{0.1519336751, 0.2004414748, -0.1394754602}, x2.Grad().Data(), 1.0e-05)

	assert.InDeltaSlice(t, []T{
		0.0021317013, -0.0012919402, 0.000064597,
		0.1915774162, -0.116107525, 0.0058053762,
	}, model.WFor.Grad().Data(), 1.0e-05)

	assert.InDeltaSlice(t, []T{
		0.0006459701, 0.0580537625,
	}, model.BFor.Grad().Data(), 1.0e-05)

	assert.InDeltaSlice(t, []T{
		-0.0086247377, 0.0259952369, -0.0009604806,
		0.0488884721, 0.0399603087, -0.0008611542,
	}, model.WIn.Grad().Data(), 1.0e-05)

	assert.InDeltaSlice(t, []T{
		-0.0028191819, 0.0141256817,
	}, model.BIn.Grad().Data(), 1.0e-05)

	assert.InDeltaSlice(t, []T{
		-0.053950159, 0.0250862936, -0.0013786491,
		1.0348122585, -0.62172667, 0.0311750786,
	}, model.WCand.Grad().Data(), 1.0e-05)

	assert.InDeltaSlice(t, []T{
		-5.72358971272346e-005, 0.0006297756,
		-0.0051438282, 0.056598355,
	}, model.WForRec.Grad().Data(), 1.0e-05)

	assert.InDeltaSlice(t, []T{
		0.000550411, -0.0060562594,
		-0.000244289, 0.00268795,
	}, model.WInRec.Grad().Data(), 1.0e-05)
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
	return model
}
