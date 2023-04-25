// Copyright 2019 spaGO Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ran

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

	x := mat.NewVecDense([]T{-0.8, -0.9, -0.9, 1.0}, mat.WithGrad(true))
	st := model.Next(nil, x)

	assert.InDeltaSlice(t, []T{0.39652, 0.25162, 0.5, 0.70475, 0.45264}, st.InG.Value().Data(), 1.0e-05)
	assert.InDeltaSlice(t, []T{0.85321, 0.43291, 0.11609, 0.51999, 0.24232}, st.ForG.Value().Data(), 1.0e-05)
	assert.InDeltaSlice(t, []T{1.02, -0.1, 0.1, 2.03, -1.41}, st.Cand.Value().Data(), 1.0e-05)
	assert.InDeltaSlice(t, []T{0.38375, -0.02516, 0.04996, 0.8918, -0.56369}, st.Y.Value().Data(), 1.0e-05)

	// == Backward

	gold := mat.NewVecDense([]T{0.57, 0.75, -0.15, 1.64, 0.45})
	loss := losses.MSE(st.Y, gold, false)
	ag.Backward(loss)

	assert.InDeltaSlice(t, []T{0.21996, -0.12731, 0.10792, 0.49361}, x.Grad().Data(), 1.0e-05)

	assert.InDeltaSlice(t, []T{
		0.03101, 0.03489, 0.03489, -0.03877,
		-0.01167, -0.01313, -0.01313, 0.01459,
		-0.00399, -0.00449, -0.00449, 0.00499,
		0.05175, 0.05822, 0.05822, -0.06469,
		-0.19328, -0.21744, -0.21744, 0.2416,
	}, model.WIn.Grad().Data(), 1.0e-05)

	assert.InDeltaSlice(t, []T{-0.03877, 0.01459, 0.00499, -0.06469, 0.2416}, model.BIn.Grad().Data(), 1.0e-05)

	assert.InDeltaSlice(t, []T{
		0.05038, 0.05668, 0.05668, -0.06298,
		0.15594, 0.17543, 0.17543, -0.19492,
		-0.07978, -0.08976, -0.08976, 0.09973,
		0.08635, 0.09714, 0.09714, -0.10794,
		0.25044, 0.28174, 0.28174, -0.31304,
	}, model.WCand.Grad().Data(), 1.0e-05)

	assert.InDeltaSlice(t, []T{-0.06298, -0.19492, 0.09973, -0.10794, -0.31304}, model.BCand.Grad().Data(), 1.0e-05)

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

	// == Forward

	s0 := &State{
		C: mat.NewVecDense([]T{-0.2, 0.2, -0.3, -0.9, -0.8}, mat.WithGrad(true)),
		Y: mat.NewVecDense([]T{-0.2, 0.2, -0.3, -0.9, -0.8}, mat.WithGrad(true)),
	}
	x := mat.NewVecDense([]T{-0.8, -0.9, -0.9, 1.0}, mat.WithGrad(true))
	s1 := model.Next(s0, x)

	assert.InDeltaSlice(t, []T{0.72312, 0.24974, 0.54983, 0.82054, 0.53494}, s1.InG.Value().Data(), 1.0e-05)
	assert.InDeltaSlice(t, []T{0.91133, 0.18094, 0.04834, 0.67481, 0.38936}, s1.ForG.Value().Data(), 1.0e-05)
	assert.InDeltaSlice(t, []T{1.02, -0.1, 0.1, 2.03, -1.41}, s1.Cand.Value().Data(), 1.0e-05)
	assert.InDeltaSlice(t, []T{0.5045, 0.01121, 0.04046, 0.78504, -0.78786}, s1.Y.Value().Data(), 1.0e-05)

	// == Backward

	gold := mat.NewVecDense([]T{0.57, 0.75, -0.15, 1.64, 0.45})
	loss := losses.MSE(s1.Y, gold, false)
	ag.Backward(loss)

	assert.InDeltaSlice(t, []T{0.19694, 0.11428, -0.14008, 0.15609}, x.Grad().Data(), 1.0e-05)

	assert.InDeltaSlice(t, []T{
		0.00798, 0.00898, 0.00898, -0.00997,
		-0.01107, -0.01246, -0.01246, 0.01384,
		-0.00377, -0.00424, -0.00424, 0.00471,
		0.07845, 0.08826, 0.08826, -0.09807,
		-0.13175, -0.14822, -0.14822, 0.16469,
	}, model.WIn.Grad().Data(), 1.0e-05)

	assert.InDeltaSlice(t, []T{-0.00997, 0.01384, 0.00471, -0.09807, 0.16469}, model.BIn.Grad().Data(), 1.0e-05)

	assert.InDeltaSlice(t, []T{
		0.02825, 0.03178, 0.03178, -0.03531,
		0.14759, 0.16603, 0.16603, -0.18448,
		-0.08364, -0.09409, -0.09409, 0.10455,
		0.21535, 0.24227, 0.24227, -0.26919,
		0.20092, 0.22604, 0.22604, -0.25115,
	}, model.WCand.Grad().Data(), 1.0e-05)

	assert.InDeltaSlice(t, []T{-0.03531, -0.18448, 0.10455, -0.26919, -0.25115}, model.BCand.Grad().Data(), 1.0e-05)

	assert.InDeltaSlice(t, []T{
		0.00199, -0.00199, 0.00299, 0.00898, 0.00798,
		-0.00277, 0.00277, -0.00415, -0.01246, -0.01107,
		-0.00094, 0.00094, -0.00141, -0.00424, -0.00377,
		0.01961, -0.01961, 0.02942, 0.08826, 0.07845,
		-0.03294, 0.03294, -0.04941, -0.14822, -0.13175,
	}, model.WInRec.Grad().Data(), 1.0e-05)

	assert.InDeltaSlice(t, []T{
		-0.00016, 0.00016, -0.00024, -0.00071, -0.00063,
		0.00438, -0.00438, 0.00657, 0.01971, 0.01752,
		0.00052, -0.00052, 0.00079, 0.00236, 0.00210,
		-0.01296, 0.01296, -0.01944, -0.05831, -0.05183,
		-0.01786, 0.01786, -0.02679, -0.08037, -0.07144,
	}, model.WForRec.Grad().Data(), 1.0e-05)

	assert.InDeltaSlice(t, []T{
		-0.00063, -0.00071, -0.00071, 0.00079,
		0.01752, 0.01971, 0.01971, -0.02189,
		0.00210, 0.00236, 0.00236, -0.00262,
		-0.05183, -0.05831, -0.05831, 0.06479,
		-0.07144, -0.08037, -0.08037, 0.08930,
	}, model.WFor.Grad().Data(), 1.0e-05)
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
	mat.SetData[T](model.BCand.Value(), []T{0.2, 0.0, -0.9, 0.7, -0.3})
	return model
}
