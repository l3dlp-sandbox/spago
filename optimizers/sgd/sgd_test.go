// Copyright 2019 spaGO Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package sgd

import (
	"testing"

	"github.com/nlpodyssey/spago/mat"
	"github.com/nlpodyssey/spago/mat/float"
	"github.com/stretchr/testify/assert"
)

func TestSGD_Update(t *testing.T) {
	t.Run("float32", testSGDUpdate[float32])
	t.Run("float64", testSGDUpdate[float64])
}

func testSGDUpdate[T float.DType](t *testing.T) {
	updater := New[T](NewConfig(
		0.001, // learning rate
		0.0,   // momentum
		false, // nesterov
	))

	params := mat.NewVecDense([]T{0.4, 0.4, 0.5, 1.0, 0.8})
	grads := mat.NewVecDense([]T{0.9, 0.7, 0.4, 0.8, 0.1})
	supp := updater.NewState(params.Shape()...).([]mat.Matrix)

	params.SubInPlace(updater.calcDelta(grads, supp))

	assert.InDeltaSlice(t, []T{0.3991, 0.3993, 0.4996, 0.9992, 0.7999}, params.Data(), 1.0e-6)
}

func TestSGDMomentum_Update(t *testing.T) {
	t.Run("float32", testSGDMomentumUpdate[float32])
	t.Run("float64", testSGDMomentumUpdate[float64])
}

func testSGDMomentumUpdate[T float.DType](t *testing.T) {
	updater := New[T](NewConfig(
		0.001, // learning rate
		0.9,   // momentum
		false, // nesterov
	))

	params := mat.NewVecDense([]T{0.4, 0.4, 0.5, 1.0, 0.8})
	grads := mat.NewVecDense([]T{0.9, 0.7, 0.4, 0.8, 0.1})

	supp := updater.NewState(params.Shape()...).([]mat.Matrix)
	mat.SetData[T](supp[v], []T{0.7, 0.8, 0.5, 0.3, 0.2})

	params.SubInPlace(updater.calcDelta(grads, supp))

	assert.InDeltaSlice(t, []T{-0.2309, -0.3207, 0.0496, 0.7292, 0.6199}, params.Data(), 1.0e-6)
}

func TestSGDMomentum_Update2(t *testing.T) {
	t.Run("float32", testSGDMomentumUpdate2[float32])
	t.Run("float64", testSGDMomentumUpdate2[float64])
}

func testSGDMomentumUpdate2[T float.DType](t *testing.T) {
	updater := New[T](NewConfig(
		0.001, // learning rate
		0.9,   // momentum
		false, // nesterov
	))

	params := mat.NewDense(3, 3, []T{
		1.4, 1.3, 0,
		-0.8, 0.16, 0.65,
		0.7, -0.4, 0.2,
	})

	grads := mat.NewDense(3, 3, []T{
		0.5, 0.3, -0.1,
		-0.6, -0.4, -1.0,
		0.5, -0.6, 0.1,
	})

	supp := updater.NewState(params.Shape()...).([]mat.Matrix)

	// === First iteration

	params.SubInPlace(updater.calcDelta(grads, supp))

	assert.InDeltaSlice(t, []T{
		0.0005, 0.0003, -0.0001,
		-0.0006, -0.0004, -0.001,
		0.0005, -0.0006, 0.0001,
	}, supp[v].Data(), 1.0e-6)

	assert.InDeltaSlice(t, []T{
		1.3995, 1.2997, 0.0001,
		-0.7994, 0.1604, 0.651,
		0.6995, -0.3994, 0.1999,
	}, params.Data(), 1.0e-6)

	// === Second iteration

	grads2 := mat.NewDense(3, 3, []T{
		0.7, 0.44, -0.66,
		-0.56, 0.4, 1.4,
		0.44, 1.44, 2.44,
	})

	params.SubInPlace(updater.calcDelta(grads2, supp))

	assert.InDeltaSlice(t, []T{
		0.00115, 0.00071, -0.00075,
		-0.0011, 4e-05, 0.0005,
		0.00089, 0.0009, 0.00253,
	}, supp[v].Data(), 1.0e-6)

	assert.InDeltaSlice(t, []T{
		1.39835, 1.29899, 0.00085,
		-0.7983, 0.16036, 0.6505,
		0.69861, -0.4003, 0.19737,
	}, params.Data(), 1.0e-5)
}

func TestSGDNesterovMomentum_Update(t *testing.T) {
	t.Run("float32", testSGDNesterovMomentumUpdate[float32])
	t.Run("float64", testSGDNesterovMomentumUpdate[float64])
}

func testSGDNesterovMomentumUpdate[T float.DType](t *testing.T) {
	updater := New[T](NewConfig(
		0.001, // learning rate
		0.9,   // momentum
		true,  // nesterov
	))

	params := mat.NewVecDense([]T{0.4, 0.4, 0.5, 1.0, 0.8})
	grads := mat.NewVecDense([]T{0.9, 0.7, 0.4, 0.8, 0.1})

	supp := updater.NewState(params.Shape()...).([]mat.Matrix)
	mat.SetData[T](supp[v], []T{0.7, 0.8, 0.5, 0.3, 0.2})

	params.SubInPlace(updater.calcDelta(grads, supp))

	assert.InDeltaSlice(t, []T{-0.16871, -0.24933, 0.09424, 0.75548, 0.63781}, params.Data(), 1.0e-6)
}

func TestSGDNesterovMomentum_Update2(t *testing.T) {
	t.Run("float32", testSGDNesterovMomentumUpdate2[float32])
	t.Run("float64", testSGDNesterovMomentumUpdate2[float64])
}

func testSGDNesterovMomentumUpdate2[T float.DType](t *testing.T) {
	updater := New[T](NewConfig(
		0.001, // learning rate
		0.9,   // momentum
		true,  // nesterov
	))

	params := mat.NewDense(3, 3, []T{
		1.4, 1.3, 0,
		-0.8, 0.16, 0.65,
		0.7, -0.4, 0.2,
	})

	grads := mat.NewDense(3, 3, []T{
		0.5, 0.3, -0.1,
		-0.6, -0.4, -1.0,
		0.5, -0.6, 0.1,
	})

	supp := updater.NewState(params.Shape()...).([]mat.Matrix)

	// === First iteration

	params.SubInPlace(updater.calcDelta(grads, supp))

	assert.InDeltaSlice(t, []T{
		0.0005, 0.0003, -0.0001,
		-0.0006, -0.0004, -0.001,
		0.0005, -0.0006, 0.0001,
	}, supp[v].Data(), 1.0e-6)

	assert.InDeltaSlice(t, []T{
		1.39905, 1.29943, 0.00019,
		-0.79886, 0.16076, 0.6519,
		0.69905, -0.39886, 0.19981,
	}, params.Data(), 1.0e-6)

	// === Second iteration

	grads2 := mat.NewDense(3, 3, []T{
		0.7, 0.44, -0.66,
		-0.56, 0.4, 1.4,
		0.44, 1.44, 2.44,
	})

	params.SubInPlace(updater.calcDelta(grads2, supp))

	assert.InDeltaSlice(t, []T{
		0.00115, 0.00071, -0.00075,
		-0.0011, 4e-05, 0.0005,
		0.00089, 0.0009, 0.00253,
	}, supp[v].Data(), 1.0e-6)

	assert.InDeltaSlice(t, []T{
		1.397315, 1.298351, 0.001525,
		-0.79731, 0.160324, 0.65005,
		0.697809, -0.40111, 0.195093,
	}, params.Data(), 1.0e-6)
}
