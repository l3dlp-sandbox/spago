// Copyright 2019 spaGO Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ag

import (
	"fmt"
	"testing"

	"github.com/nlpodyssey/spago/mat"
	"github.com/nlpodyssey/spago/mat/float"
	"github.com/nlpodyssey/spago/mat/mattest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewVariable(t *testing.T) {
	t.Run("float32", testNewVariable[float32])
	t.Run("float64", testNewVariable[float64])
}

func testNewVariable[T float.DType](t *testing.T) {
	testCases := []struct {
		value        mat.Matrix
		requiresGrad bool
	}{
		{mat.NewScalar[T](42), true},
		{mat.NewScalar[T](42), false},
	}

	for _, tc := range testCases {
		name := fmt.Sprintf("Var(%g, %v)", tc.value, tc.requiresGrad)
		t.Run(name, func(t *testing.T) {
			v := Var(tc.value).WithGrad(tc.requiresGrad)
			require.NotNil(t, v)
			assert.Same(t, tc.value, v.Value())
			assert.Nil(t, v.Grad())
			assert.False(t, v.HasGrad())
			assert.Equal(t, tc.requiresGrad, v.RequiresGrad())
		})
	}
}

func TestNewVariableWithName(t *testing.T) {
	t.Run("float32", testNewVariableWithName[float32])
	t.Run("float64", testNewVariableWithName[float64])
}

func testNewVariableWithName[T float.DType](t *testing.T) {
	testCases := []struct {
		value        mat.Matrix
		requiresGrad bool
		name         string
	}{
		{mat.NewScalar[T](42), true, "foo"},
		{mat.NewScalar[T](42), false, "bar"},
	}

	for _, tc := range testCases {
		name := fmt.Sprintf("NewVariableWithName(%g, %v, %#v)", tc.value, tc.requiresGrad, tc.name)
		t.Run(name, func(t *testing.T) {
			v := Var(tc.value).WithGrad(tc.requiresGrad)
			require.NotNil(t, v)
			assert.Same(t, tc.value, v.Value())
			assert.Nil(t, v.Grad())
			assert.False(t, v.HasGrad())
			assert.Equal(t, tc.requiresGrad, v.RequiresGrad())
		})
	}
}

func TestNewScalar(t *testing.T) {
	t.Run("float32", testNewScalar[float32])
	t.Run("float64", testNewScalar[float64])
}

func testNewScalar[T float.DType](t *testing.T) {
	v := Var(mat.NewScalar(T(42)))
	require.NotNil(t, v)
	mattest.AssertMatrixEquals(t, mat.NewScalar[T](42), v.Value())
	assert.Nil(t, v.Grad())
	assert.False(t, v.HasGrad())
	assert.False(t, v.RequiresGrad())
}

func TestNewScalarWithName(t *testing.T) {
	t.Run("float32", testNewScalarWithName[float32])
	t.Run("float64", testNewScalarWithName[float64])
}

func testNewScalarWithName[T float.DType](t *testing.T) {
	v := Var(mat.NewScalar(T(42)))
	require.NotNil(t, v)
	mattest.AssertMatrixEquals(t, mat.NewScalar[T](42), v.Value())
	assert.Nil(t, v.Grad())
	assert.False(t, v.HasGrad())
	assert.False(t, v.RequiresGrad())
}

func TestConstant(t *testing.T) {
	t.Run("float32", testConstant[float32])
	t.Run("float64", testConstant[float64])
}

func testConstant[T float.DType](t *testing.T) {
	v := Var(mat.NewScalar(T(42)))
	require.NotNil(t, v)
	mattest.AssertMatrixEquals(t, mat.NewScalar[T](42), v.Value())
	assert.Nil(t, v.Grad())
	assert.False(t, v.HasGrad())
	assert.False(t, v.RequiresGrad())
}

func TestVariable_Gradients(t *testing.T) {
	t.Run("float32", testVariableGradients[float32])
	t.Run("float64", testVariableGradients[float64])
}

func testVariableGradients[T float.DType](t *testing.T) {
	t.Run("with requires gradient true", func(t *testing.T) {
		v := Var(mat.NewScalar[T](42)).WithGrad(true)
		require.Nil(t, v.Grad())
		assert.False(t, v.HasGrad())

		v.AccGrad(mat.NewScalar[T](5))
		mattest.RequireMatrixEquals(t, mat.NewScalar[T](5), v.Grad())
		assert.True(t, v.HasGrad())

		v.AccGrad(mat.NewScalar[T](10))
		mattest.RequireMatrixEquals(t, mat.NewScalar[T](15), v.Grad())
		assert.True(t, v.HasGrad())

		v.ZeroGrad()
		require.Nil(t, v.Grad())
		assert.False(t, v.HasGrad())
	})

	t.Run("with requires gradient false", func(t *testing.T) {
		v := Var(mat.NewScalar[T](42)).WithGrad(false)
		require.Nil(t, v.Grad())
		assert.False(t, v.HasGrad())

		v.AccGrad(mat.NewScalar[T](5))
		require.Nil(t, v.Grad())
		assert.False(t, v.HasGrad())

		v.ZeroGrad()
		require.Nil(t, v.Grad())
		assert.False(t, v.HasGrad())
	})
}
