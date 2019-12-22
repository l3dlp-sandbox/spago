// Copyright 2019 spaGO Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package losses

import (
	"brillion.io/spago/pkg/ml/ag"
)

// MAE measures the mean absolute error (a.k.a. L1 Loss) between each element in the input x and target y.
func MAE(g *ag.Graph, x ag.Node, y ag.Node, reduceMean bool) ag.Node {
	loss := g.Abs(g.Sub(x, y))
	if reduceMean {
		return g.ReduceMean(loss)
	} else {
		return g.ReduceSum(loss)
	}
}

// MSE measures the mean squared error (squared L2 norm) between each element in the input x and target y.
func MSE(g *ag.Graph, x ag.Node, y ag.Node, reduceMean bool) ag.Node {
	loss := g.ProdScalar(g.Square(g.Sub(x, y)), g.NewScalar(0.5))
	if reduceMean {
		return g.ReduceMean(loss)
	} else {
		return g.ReduceSum(loss)
	}
}

// NLL returns the loss of the input x respect to the target y.
// The target is expected to be a one-hot vector.
func NLL(g *ag.Graph, x ag.Node, y ag.Node) ag.Node {
	return g.Neg(g.ReduceSum(g.Prod(y, g.Log(x))))
}

// c is the index of the gold class
func CrossEntropy(g *ag.Graph, x ag.Node, c int) ag.Node {
	return g.Add(g.Neg(g.AtVec(x, c)), g.Log(g.ReduceSum(g.Exp(x))))
}

func Perplexity(g *ag.Graph, x ag.Node, c int) ag.Node {
	return g.Exp(CrossEntropy(g, x, c))
}

func ZeroOneQuantization(g *ag.Graph, x ag.Node) ag.Node {
	return g.ReduceSum(g.Prod(g.Square(x), g.Square(g.ReverseSub(x, g.NewScalar(1.0)))))
}

func Norm2Quantization(g *ag.Graph, x ag.Node) ag.Node {
	return g.Square(g.SubScalar(g.ReduceSum(g.Square(x)), g.NewScalar(1.0)))
}

// q is the quantization regularizer weight (suggested  0.00001)
func OneHotQuantization(g *ag.Graph, x ag.Node, q float64) ag.Node {
	return g.ProdScalar(g.Add(ZeroOneQuantization(g, x), Norm2Quantization(g, x)), g.NewScalar(q))
}

func Distance(g *ag.Graph, x ag.Node, target float64) ag.Node {
	return g.Abs(g.Sub(g.NewScalar(target), x))
}

func MSESeq(g *ag.Graph, predicted []ag.Node, target []ag.Node, reduceMean bool) ag.Node {
	loss := MSE(g, predicted[0], target[0], false)
	for i := 1; i < len(predicted); i++ {
		loss = g.Add(loss, MSE(g, predicted[i], target[i], false))
	}
	if reduceMean {
		loss = g.DivScalar(loss, g.NewScalar(float64(len(predicted))))
	}
	return loss
}

func CrossEntropySeq(g *ag.Graph, predicted []ag.Node, target []int, reduceMean bool) ag.Node {
	loss := CrossEntropy(g, predicted[0], target[0])
	for i := 1; i < len(predicted); i++ {
		loss = g.Add(loss, CrossEntropy(g, predicted[i], target[i]))
	}
	if reduceMean {
		loss = g.DivScalar(loss, g.NewScalar(float64(len(predicted))))
	}
	return loss
}
