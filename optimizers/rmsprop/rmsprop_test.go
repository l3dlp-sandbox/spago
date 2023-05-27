// Copyright 2019 spaGO Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package rmsprop

import (
	"testing"

	"github.com/nlpodyssey/spago/mat"
	"github.com/nlpodyssey/spago/mat/float"
	"github.com/stretchr/testify/assert"
)

func Test_Update(t *testing.T) {
	t.Run("float32", testUpdate[float32])
	t.Run("float64", testUpdate[float64])
}

func testUpdate[T float.DType](t *testing.T) {
	updater := New[T](NewConfig(0.001, 1e-06, 0.9))
	params := mat.NewDense[T](mat.WithBacking([]T{0.4, 0.4, 0.5, 1.0, 0.8}))
	grads := mat.NewDense[T](mat.WithBacking([]T{0.9, 0.7, 0.4, 0.8, 0.1}))
	supp := updater.NewState(params.Shape()...).([]mat.Matrix)
	mat.SetData[T](supp[v], []T{1.0, 0.4, 0.7, 0.0, 0.2})

	params.SubInPlace(updater.calcDelta(grads, supp))

	assert.InDeltaSlice(t, []T{0.399091, 0.398905, 0.499502, 0.996838, 0.799765}, params.Data(), 1.0e-6)
}

func Test_Update2(t *testing.T) {
	t.Run("float32", testUpdate2[float32])
	t.Run("float64", testUpdate2[float64])
}

func testUpdate2[T float.DType](t *testing.T) {
	updater := New[T](NewConfig(
		0.001, // learning rate
		1e-08, // epsilon
		0.9,   // decay
	))

	params := mat.NewDense[T](mat.WithShape(3, 3), mat.WithBacking([]T{
		1.4, 1.3, 0,
		-0.8, 0.16, 0.65,
		0.7, -0.4, 0.2,
	}))

	grads := mat.NewDense[T](mat.WithShape(3, 3), mat.WithBacking([]T{
		0.5, 0.3, -0.1,
		-0.6, -0.4, -1.0,
		0.5, -0.6, 0.1,
	}))

	supp := updater.NewState(params.Shape()...).([]mat.Matrix)

	// === First iteration

	params.SubInPlace(updater.calcDelta(grads, supp))

	assert.InDeltaSlice(t, []T{
		0.025, 0.009, 0.001,
		0.036, 0.016, 0.1,
		0.025, 0.036, 0.001,
	}, supp[v].Data(), 1.0e-6)

	assert.InDeltaSlice(t, []T{
		1.39683772253983, 1.29683772267316, 0.003162276660169,
		-0.796837722506498, 0.163162277410168, 0.653162277560168,
		0.696837722539832, -0.396837722506498, 0.196837723339831,
	}, params.Data(), 1.0e-6)

	// === Second iteration

	grads2 := mat.NewDense[T](mat.WithShape(3, 3), mat.WithBacking([]T{
		0.7, 0.44, -0.66,
		-0.56, 0.4, 1.4,
		0.44, 1.44, 2.44,
	}))

	params.SubInPlace(updater.calcDelta(grads2, supp))

	assert.InDeltaSlice(t, []T{
		0.0715, 0.02746, 0.04446,
		0.06376, 0.0304, 0.286,
		0.04186, 0.23976, 0.59626,
	}, supp[v].Data(), 1.0e-6)

	assert.InDeltaSlice(t, []T{
		1.39421987106571, 1.29418249122086, 0.006292383674455,
		-0.79461996603293, 0.160868120203042, 0.650544426037096,
		0.694687155213517, -0.399778580934813, 0.19367783320647,
	}, params.Data(), 1.0e-5)
}
