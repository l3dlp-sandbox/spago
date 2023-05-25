// Copyright 2019 spaGO Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package hyperbolic

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHyperbolic_Decay(t *testing.T) {
	fn := New(0.01, 0.001, 0.5)

	assert.InDelta(t, 0.01, fn.Decay(0.01, 1), 1.0e-06)
	assert.InDelta(t, 0.005, fn.Decay(0.01, 2), 1.0e-06)
	assert.InDelta(t, 0.004, fn.Decay(0.00774263682, 3), 1.0e-06)
	assert.InDelta(t, 0.001, fn.Decay(0.001, 10), 1.0e-06)
}

func TestNew(t *testing.T) {
	assert.NotPanics(t, func() { New(0.01, 0.001, 0.5) }, "The New did panic unexpectedly")
	assert.Panics(t, func() { New(0.001, 0.01, 0.5) }, "The New had to panic with init lr < final lr")
}
