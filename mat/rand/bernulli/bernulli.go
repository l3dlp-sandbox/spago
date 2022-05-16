// Copyright 2019 spaGO Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bernulli

import (
	"math"

	"github.com/nlpodyssey/spago/mat"
	"github.com/nlpodyssey/spago/mat/float"
	"github.com/nlpodyssey/spago/mat/rand"
	"github.com/nlpodyssey/spago/mat/rand/uniform"
)

// Distribution creates a new matrix initialized with Bernoulli distribution.
func Distribution[T float.DType](r, c int, prob T, generator *rand.LockedRand) mat.Matrix {
	out := mat.NewEmptyDense[T](r, c)
	dist := uniform.New(0.0, 1.0, generator)
	for i := 0; i < r; i++ {
		for j := 0; j < c; j++ {
			val := T(dist.Next())
			fl := math.Floor(float64(val))
			if val < prob {
				out.SetScalar(i, j, float.Float(fl))
			} else {
				out.SetScalar(i, j, float.Float(fl+1))
			}
		}
	}
	return out
}
