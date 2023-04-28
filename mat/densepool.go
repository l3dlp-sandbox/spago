// Copyright 2022 spaGO Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// Part of this code is a modified version of Gonum matrix
// pool handling:
// https://github.com/gonum/gonum/blob/master/mat/pool.go
//
// Copyright ©2014 The Gonum Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file available
// at https://github.com/gonum/gonum/blob/master/LICENSE.

package mat

import (
	"fmt"
	"math"
	"math/bits"
	"runtime"
	"sync"

	"github.com/nlpodyssey/spago/mat/float"
)

// densePoolType provides pools for slice lengths from 0 to 64 bits
// (0 to MaxUint64).
type densePoolType[T float.DType] [65]sync.Pool

var (
	densePoolFloat32 = newDensePool[float32]()
	densePoolFloat64 = newDensePool[float64]()
)

// newDensePool creates a new pool for handling matrices of a specific DType.
func newDensePool[T float.DType]() *densePoolType[T] {
	dp := new(densePoolType[T])
	for i := range dp {
		dp[i].New = dp.makeNewFunc(i)
	}
	return dp
}

// densePool returns the global (sort-of singleton) pre-instantiated pool
// for a specific DType.
func densePool[T float.DType]() *densePoolType[T] {
	switch any(T(0)).(type) {
	case float32:
		return any(densePoolFloat32).(*densePoolType[T])
	case float64:
		return any(densePoolFloat64).(*densePoolType[T])
	default:
		panic(fmt.Sprintf("mat: no Dense pool for type %T", T(0)))
	}
}

func (dp *densePoolType[T]) makeNewFunc(bitsLen int) func() any {
	var length uint
	if bitsLen >= 64 {
		length = math.MaxUint64
	} else {
		length = 1<<bitsLen - 1
	}
	return func() any {
		return &Dense[T]{
			rows:  -1,
			cols:  -1,
			isNew: true,
			data:  make([]T, length),
		}
	}
}

// Get returns a Dense matrix from the pool, with size rows×cols, and
// a raw data slice with a cap in the range rows*cols < cap <= 2*rows*cols.
//
// Warning: the values may not be all zeros. To ensure that all elements
// are initialized to zero, see GetEmptyDense.
func (dp *densePoolType[T]) Get(rows, cols int) *Dense[T] {
	return dp.get(rows, cols)
}

// GetEmpty returns a Dense matrix from the pool, with size rows×cols, and
// a raw data slice with a cap in the range rows*cols < cap <= 2*rows*cols.
//
// All elements are guaranteed to be initialized to zero.
func (dp *densePoolType[T]) GetEmpty(rows, cols int) *Dense[T] {
	d := dp.get(rows, cols)
	if !d.isNew {
		zeros(d.data)
	}
	return d
}

// zeros sets all elements of a slice to zero.
func zeros[T float.DType](s []T) {
	for i := range s {
		s[i] = 0
	}
}

func (dp *densePoolType[T]) get(rows, cols int) *Dense[T] {
	if rows < 0 || cols < 0 {
		panic("mat: negative values for rows and cols are not allowed")
	}
	length := uint(rows * cols)
	bitsLen := bits.Len(length)
	d := dp[bitsLen].Get().(*Dense[T])
	d.data = d.data[:length]
	d.rows = rows
	d.cols = cols

	runtime.SetFinalizer(d, func(o any) {
		d.grad = nil
		d.isNew = false
		dp[bitsLen].Put(d)
	})

	return d
}
