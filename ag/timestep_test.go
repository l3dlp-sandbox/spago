// Copyright 2022 spaGO Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ag

import (
	"testing"

	"github.com/nlpodyssey/spago/mat"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTimeStepHandler(t *testing.T) {
	t.Run("float32", testTimeStepHandler[float32])
	t.Run("float64", testTimeStepHandler[float64])
}

func testTimeStepHandler[T mat.DType](t *testing.T) {
	tsh := NewTimeStepHandler()
	require.Equal(t, 0, tsh.CurrentTimeStep())

	// Simulate some parameters, with no associated time step (default 0)
	paramA := NewVariableWithName[T](mat.NewScalar[T](1), true, "Param 0")
	paramB := NewVariableWithName[T](mat.NewScalar[T](2), true, "Param 1")

	// Perform an operation while still on initial time step 0
	paramsSum := Sum(paramA, paramB)

	// Time step 1
	tsh.IncTimeStep()
	require.Equal(t, 1, tsh.CurrentTimeStep())
	in1 := NewVariableWithName[T](mat.NewScalar[T](3), true, "Input 0")

	a1 := Add(paramA, in1)
	b1 := Add(a1, paramB)
	out1 := Add(b1, paramsSum)

	// Time step 2
	tsh.IncTimeStep()
	require.Equal(t, 2, tsh.CurrentTimeStep())
	in2 := NewVariableWithName[T](mat.NewScalar[T](4), true, "Input 1")

	a2 := Add(paramA, in2)
	b2 := Add(a2, paramB)
	c2 := Add(paramsSum, out1) // note: this is not linked to the input
	out2 := Add(b2, c2)

	// Time step 2
	tsh.IncTimeStep()
	in3 := NewVariableWithName[T](mat.NewScalar[T](4), true, "Input 3")

	a3 := Add(paramA, in3)
	b3 := Add(a3, paramB)
	c3 := Add(paramsSum, out2) // note: this is not linked to the input
	out3 := Add(b3, c3)

	assert.Equal(t, 0, tsh.NodeTimeStep(paramA))
	assert.Equal(t, 0, tsh.NodeTimeStep(paramB))
	assert.Equal(t, 0, tsh.NodeTimeStep(paramsSum))

	assert.Equal(t, 0, tsh.NodeTimeStep(in1))
	assert.Equal(t, 0, tsh.NodeTimeStep(in2))
	assert.Equal(t, 0, tsh.NodeTimeStep(in3))

	assert.Equal(t, 1, tsh.NodeTimeStep(a1))
	assert.Equal(t, 1, tsh.NodeTimeStep(b1))
	assert.Equal(t, 1, tsh.NodeTimeStep(out1))

	assert.Equal(t, 2, tsh.NodeTimeStep(a2))
	assert.Equal(t, 2, tsh.NodeTimeStep(b2))
	assert.Equal(t, 2, tsh.NodeTimeStep(c2))
	assert.Equal(t, 2, tsh.NodeTimeStep(out2))

	assert.Equal(t, 3, tsh.NodeTimeStep(a3))
	assert.Equal(t, 3, tsh.NodeTimeStep(b3))
	assert.Equal(t, 3, tsh.NodeTimeStep(c3))
	assert.Equal(t, 3, tsh.NodeTimeStep(out3))
}
