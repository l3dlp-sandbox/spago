// Copyright 2019 spaGO Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package crf

import (
	"testing"

	"github.com/nlpodyssey/spago/ag"
	"github.com/nlpodyssey/spago/mat"
	"github.com/nlpodyssey/spago/mat/float"
	"github.com/stretchr/testify/assert"
)

func TestModel_Decode(t *testing.T) {
	t.Run("float32", testModelDecode[float32])
	t.Run("float64", testModelDecode[float64])
}

func testModelDecode[T float.DType](t *testing.T) {
	model := newTestModel[T]()

	w1 := mat.NewVecDense([]T{1.7, 0.2, -0.3, 0.5}, mat.WithGrad(true))
	w2 := mat.NewVecDense([]T{2.0, -3.5, 0.1, 2.0}, mat.WithGrad(true))
	w3 := mat.NewVecDense([]T{-2.5, 3.2, -0.2, -0.3}, mat.WithGrad(true))
	w4 := mat.NewVecDense([]T{3.3, -0.9, 2.7, -2.7}, mat.WithGrad(true))
	w5 := mat.NewVecDense([]T{0.5, 0.2, 0.4, 1.4}, mat.WithGrad(true))
	y := model.Decode([]ag.Node{w1, w2, w3, w4, w5})

	gold := []int{3, 3, 1, 0, 3}

	assert.Equal(t, gold, y)
}

func TestModel_GoldScore(t *testing.T) {
	t.Run("float32", testModelGoldScore[float32])
	t.Run("float64", testModelGoldScore[float64])
}

func testModelGoldScore[T float.DType](t *testing.T) {
	model := newTestModel[T]()

	w1 := mat.NewVecDense([]T{1.7, 0.2, -0.3, 0.5}, mat.WithGrad(true))
	w2 := mat.NewVecDense([]T{2.0, -3.5, 0.1, 2.0}, mat.WithGrad(true))
	w3 := mat.NewVecDense([]T{-2.5, 3.2, -0.2, -0.3}, mat.WithGrad(true))
	w4 := mat.NewVecDense([]T{3.3, -0.9, 2.7, -2.7}, mat.WithGrad(true))
	w5 := mat.NewVecDense([]T{0.5, 0.2, 0.4, 1.4}, mat.WithGrad(true))
	gold := []int{0, 0, 1, 0, 3}
	y := model.goldScore([]ag.Node{w1, w2, w3, w4, w5}, gold)

	assert.InDeltaSlice(t, []T{14.27}, y.Value().Data(), 0.00001)
}

func TestModel_TotalScore(t *testing.T) {
	t.Run("float32", testModelTotalScore[float32])
	t.Run("float64", testModelTotalScore[float64])
}

func testModelTotalScore[T float.DType](t *testing.T) {
	model := newTestModel[T]()

	w1 := mat.NewVecDense([]T{1.7, 0.2, -0.3, 0.5}, mat.WithGrad(true))
	w2 := mat.NewVecDense([]T{2.0, -3.5, 0.1, 2.0}, mat.WithGrad(true))
	w3 := mat.NewVecDense([]T{-2.5, 3.2, -0.2, -0.3}, mat.WithGrad(true))
	w4 := mat.NewVecDense([]T{3.3, -0.9, 2.7, -2.7}, mat.WithGrad(true))
	w5 := mat.NewVecDense([]T{0.5, 0.2, 0.4, 1.4}, mat.WithGrad(true))
	y := model.totalScore([]ag.Node{w1, w2, w3, w4, w5})

	assert.InDeltaSlice(t, []T{16.64258}, y.Value().Data(), 0.00001)
}

func TestModel_Loss(t *testing.T) {
	t.Run("float32", testModelLoss[float32])
	t.Run("float64", testModelLoss[float64])
}

func testModelLoss[T float.DType](t *testing.T) {
	model := newTestModel[T]()

	w1 := mat.NewVecDense([]T{1.7, 0.2, -0.3, 0.5}, mat.WithGrad(true))
	w2 := mat.NewVecDense([]T{2.0, -3.5, 0.1, 2.0}, mat.WithGrad(true))
	w3 := mat.NewVecDense([]T{-2.5, 3.2, -0.2, -0.3}, mat.WithGrad(true))
	w4 := mat.NewVecDense([]T{3.3, -0.9, 2.7, -2.7}, mat.WithGrad(true))
	w5 := mat.NewVecDense([]T{0.5, 0.2, 0.4, 1.4}, mat.WithGrad(true))
	gold := []int{0, 0, 1, 0, 3}
	loss := model.NegativeLogLoss([]ag.Node{w1, w2, w3, w4, w5}, gold)

	ag.Backward(loss)
	assert.InDeltaSlice(t, []T{2.37258}, loss.Value().Data(), 0.00001)
}

func newTestModel[T float.DType]() *Model {
	model := New[T](4)
	mat.SetData[T](model.TransitionScores.Value(), []T{
		0.0, 0.6, 0.8, 1.2, 1.6,
		0.2, 0.5, 0.02, 0.03, 0.45,
		0.3, 0.2, 0.6, 0.01, 0.19,
		0.4, 0.02, 0.02, 0.7, 0.26,
		0.9, 0.1, 0.02, 0.08, 0.8,
	})
	return model
}
