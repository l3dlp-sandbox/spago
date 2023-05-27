// Copyright 2019 spaGO Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package radam

import (
	"math"

	"github.com/nlpodyssey/spago/mat"
	"github.com/nlpodyssey/spago/mat/float"
	"github.com/nlpodyssey/spago/nn"
	"github.com/nlpodyssey/spago/optimizers"
)

var _ optimizers.StrategyConfig = &Config{}

// Config provides configuration settings for a RAdam optimizer.
type Config struct {
	optimizers.StrategyConfig
	StepSize float64
	Beta1    float64
	Beta2    float64
	Epsilon  float64
}

// NewConfig returns a new RAdam Config.
// It panics if beta1 or beta2 are not in the range [0.0, 1.0).
func NewConfig(stepSize, beta1, beta2, epsilon float64) Config {
	if !(beta1 >= 0.0 && beta1 < 1.0) {
		panic("adam: `beta1` must be in the range [0.0, 1.0)")
	}
	if !(beta2 >= 0.0 && beta2 < 1.0) {
		panic("adam: `beta2` must be in the range [0.0, 1.0)")
	}
	return Config{
		StepSize: stepSize,
		Beta1:    beta1,
		Beta2:    beta2,
		Epsilon:  epsilon,
	}
}

// NewDefaultConfig returns a new Config with generically reasonable default values.
func NewDefaultConfig() Config {
	return Config{
		StepSize: 0.001,
		Beta1:    0.9,
		Beta2:    0.999,
		Epsilon:  1.0e-8,
	}
}

var _ optimizers.Strategy = &RAdam[float32]{}

// RAdam implements the RAdam gradient descent optimization method.
type RAdam[T float.DType] struct {
	Config
	RoMax    float64 // The maximum length of the approximated SMA.
	TimeStep int
}

// New returns a new RAdam optimizer, initialized according to the given configuration.
func New[T float.DType](c Config) *RAdam[T] {
	adam := &RAdam[T]{
		Config:   c,
		RoMax:    2.0/(1.0-c.Beta2) - 1.0,
		TimeStep: 1.0,
	}
	return adam
}

// Label returns the enumeration-like value which identifies this gradient descent method.
func (o *RAdam[_]) Label() int {
	return optimizers.RAdam
}

const (
	m    int = 0
	v    int = 1
	buf1 int = 2
	buf2 int = 3
	buf3 int = 4
)

// NewState returns a new state.
func (o *RAdam[T]) NewState(shape ...int) any {
	r, c := shape[0], shape[1]
	supp := make([]mat.Matrix, 5)
	supp[m] = mat.NewDense[T](mat.WithShape(r, c))
	supp[v] = mat.NewDense[T](mat.WithShape(r, c))
	supp[buf1] = mat.NewDense[T](mat.WithShape(r, c))
	supp[buf2] = mat.NewDense[T](mat.WithShape(r, c))
	supp[buf3] = mat.NewDense[T](mat.WithShape(r, c))
	return supp
}

// IncBatch beats the occurrence of a new batch.
func (o *RAdam[_]) IncBatch() {
	o.TimeStep++
}

// CalcDelta returns the difference between the current params and where the method wants it to be.
func (o *RAdam[T]) CalcDelta(param *nn.Param) mat.Matrix {
	grads := param.Grad()
	supp := param.GetOrSetState(o.NewState).([]mat.Matrix)
	return o.calcDelta(grads, supp)
}

func (o *RAdam[T]) calcDelta(grads mat.Matrix, supp []mat.Matrix) mat.Matrix {
	updateM(grads, supp, o.Beta1)
	updateV(grads, supp, o.Beta2)
	sqrtB2T := math.Sqrt(1.0 - math.Pow(o.Beta2, float64(o.TimeStep)))
	alpha := o.calcAlpha()
	buf := supp[v].Sqrt().AddScalarInPlace(o.Epsilon * sqrtB2T)
	suppDiv := supp[m].Div(buf)
	supp[buf3].ProdMatrixScalarInPlace(suppDiv, alpha)
	return supp[buf3]
}

// m = m*beta1 + grads*(1.0-beta1)
func updateM(grads mat.Matrix, supp []mat.Matrix, beta1 float64) {
	supp[m].ProdScalarInPlace(beta1)
	supp[buf1].ProdMatrixScalarInPlace(grads, 1.0-beta1)
	supp[m].AddInPlace(supp[buf1])
}

// v = v*beta2 + (grads*grads)*(1.0-beta2)
func updateV(grads mat.Matrix, supp []mat.Matrix, beta2 float64) {
	supp[v].ProdScalarInPlace(beta2)
	sqGrad := grads.Prod(grads)
	supp[buf2].ProdMatrixScalarInPlace(sqGrad, 1.0-beta2)
	supp[v].AddInPlace(supp[buf2])
}

func (o *RAdam[T]) calcAlpha() float64 {
	timeStep := float64(o.TimeStep)
	b1T := math.Pow(o.Beta1, timeStep)
	b2T := math.Pow(o.Beta2, timeStep)
	ro := o.RoMax - 2.0*timeStep*b2T/(1.0-b2T)
	var rect float64 = 1
	if ro > 4.0 { // i.e. if the variance is tractable
		rect = math.Sqrt((ro - 4.0) * (ro - 2.0) * o.RoMax / ((o.RoMax - 4.0) * (o.RoMax - 2.0) * ro))
	}
	return o.StepSize * rect * mat.Sqrt(1.0-b2T) / (1.0 - b1T)
}
