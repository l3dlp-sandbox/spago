// Copyright 2019 spaGO Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ag

import (
	"fmt"
	"math"
	"sync"
)

// Map returns a transformed version of xs with all its components modified according to the mapping function.
// It is useful for applying an operator to a sequence of nodes. Keep in mind that using this function has an overhead
// because of the callback, however insignificant compared to mathematical computations.
func Map(mapping func(Node) Node, xs []Node) []Node {
	ys := make([]Node, len(xs))
	for i, x := range xs {
		ys[i] = mapping(x)
	}
	return ys
}

// MapConcurrent is the concurrent version of Map.
func MapConcurrent(mapping func(Node) Node, xs []Node) []Node {
	var wg sync.WaitGroup
	wg.Add(len(xs))
	ys := make([]Node, len(xs))
	for i, x := range xs {
		i, x := i, x
		go func() {
			ys[i] = mapping(x)
			wg.Done()
		}()
	}
	wg.Wait()
	return ys
}

// Map2 takes two arguments and applies a mapping function (that must take two arguments) to the items from the two node-slices in parallel.
// It panics if one slice is shorter than the other.
func Map2(mapping func(a Node, b Node) Node, xs1 []Node, xs2 []Node) []Node {
	if len(xs1) != len(xs2) {
		panic(fmt.Sprintf("ag: arguments must have the same size (%d != %d)", len(xs1), len(xs2)))
	}
	ys := make([]Node, len(xs1))
	for i, x1 := range xs1 {
		ys[i] = mapping(x1, xs2[i])
	}
	return ys
}

// Map2Concurrent is the concurrent version of Map2.
func Map2Concurrent(mapping func(a Node, b Node) Node, xs1 []Node, xs2 []Node) []Node {
	if len(xs1) != len(xs2) {
		panic(fmt.Sprintf("ag: arguments must have the same size (%d != %d)", len(xs1), len(xs2)))
	}
	var wg sync.WaitGroup
	wg.Add(len(xs1))
	ys := make([]Node, len(xs1))
	for i, x1 := range xs1 {
		i, x1 := i, x1
		go func() {
			ys[i] = mapping(x1, xs2[i])
			wg.Done()
		}()
	}
	wg.Wait()
	return ys
}

// Pad down/up samples the input to the given size.
func Pad(xs []Node, seqLen int, padding func(i int) Node) []Node {
	if len(xs) == seqLen {
		return xs
	}
	if len(xs) > seqLen {
		return xs[:seqLen]
	}
	padded := make([]Node, seqLen)
	copy(padded[:len(xs)], xs)
	for i := len(xs); i < len(padded); i++ {
		padded[i] = padding(i)
	}
	return padded
}

// SeparateMatrix returns a matrix of Node(s) represented as a slice of slice containing the elements extracted from the input.
// The dimensions of the resulting matrix are the same of the input.
func SeparateMatrix(x Node) [][]Node {
	rows, cols := x.Value().Dims()
	ys := make([][]Node, rows)
	for i := range ys {
		row := make([]Node, cols)
		for j := range row {
			row[j] = At(x, i, j)
		}
		ys[i] = row
	}
	return ys
}

// SeparateVec returns a slice of Node(s) containing the elements extracted from the input.
// The size of the vector equals the number of input elements.
// You can think of this method as the inverse of the ag.Concat operator.
func SeparateVec(x Node) []Node {
	size := x.Value().Size()
	ys := make([]Node, size)
	for i := 0; i < size; i++ {
		ys[i] = AtVec(x, i)
	}
	return ys
}

// SplitVec splits the x Node into multiple chunks.
func SplitVec(x Node, chunks int) []Node {
	if x.Value().Size()%chunks != 0 {
		panic("nn: incompatible chunks size")
	}
	l := 0
	size := int(math.Ceil(float64(x.Value().Size()) / float64(chunks)))
	ys := make([]Node, chunks)
	for i := 0; i < chunks; i++ {
		ys[i] = Slice(x, l, 0, l+size, 1)
		l += size
	}
	return ys
}

// Sum returns the value that describes the sum of the sample.
// It panics if the input is empty.
func Sum(xs ...Node) Node {
	sumVector := xs[0]
	for i := 1; i < len(xs); i++ {
		sumVector = Add(sumVector, xs[i])
	}
	return sumVector
}

// Mean returns the value that describes the average of the sample.
func Mean(xs []Node) Node {
	sumVector := xs[0]
	for i := 1; i < len(xs); i++ {
		sumVector = Add(sumVector, xs[i])
	}
	ln := sumVector.Value().NewScalar(float64(len(xs)))
	return DivScalar(sumVector, ln)
}

// Maximum returns the value that describes the maximum of the sample.
func Maximum(xs []Node) Node {
	maxVector := xs[0]
	for i := 1; i < len(xs); i++ {
		maxVector = Max(maxVector, xs[i])
	}
	return maxVector
}

// Minimum returns the value that describes the minimum of the sample.
func Minimum(xs []Node) Node {
	minVector := xs[0]
	for i := 1; i < len(xs); i++ {
		minVector = Min(minVector, xs[i])
	}
	return minVector
}

// BiLinear performs a bilinear transformation of the type (x_1 W x_2)
func BiLinear(w, x1, x2 Node) Node {
	return Mul(Mul(T(x1), w), x2)
}

// BiAffine performs a biaffine transformation.
func BiAffine(w, u, v, b, x1, x2 Node) Node {
	return Add(Add(Add(BiLinear(w, x1, x2), Mul(T(u), x1)), Mul(T(v), x2)), b)
}

// PositiveELU returns a new operator node as a result of ELU(x) + 1.
func PositiveELU(x Node) Node {
	one := x.Value().NewScalar(1)
	return AddScalar(ELU(x, one), one)
}

// LogSoftmax returns a new operator node as a result of Log(Softmax(x)).
func LogSoftmax(x Node) Node {
	return Log(Softmax(x))
}

// LogSumExp "trick" computes the log of the sum of exponentials of input elements.
// When the input is one, this must be a vector. Alternatively, the calculation
// is conducted on a list of scalars.
func LogSumExp(xs ...Node) Node {
	if len(xs) == 1 {
		x := xs[0]
		max := ReduceMax(x)
		sum := ReduceSum(Exp(SubScalar(x, max)))
		return Add(max, Log(sum))
	}

	max := ScalarMax(xs)
	var sum Node
	for _, v := range xs {
		sum = Add(sum, Exp(Sub(v, max)))
	}
	return Add(max, Log(sum))
}

// RowViews calls RowView for each row of x, returning a new slice
// of row-view Nodes.
func RowViews(x Node) []Node {
	ys := make([]Node, x.Value().Rows())
	for i := range ys {
		ys[i] = RowView(x, i)
	}
	return ys
}

// ColViews calls ColView for each column of x, returning a new slice
// of column-view Nodes.
func ColViews(x Node) []Node {
	ys := make([]Node, x.Value().Columns())
	for i := range ys {
		ys[i] = ColView(x, i)
	}
	return ys
}
