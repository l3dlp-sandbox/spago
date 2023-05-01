// Copyright 2021 spaGO Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package lamb

import (
	"math"

	"github.com/nlpodyssey/spago/mat"
	"github.com/nlpodyssey/spago/mat/float"
	"github.com/nlpodyssey/spago/nn"
	"github.com/nlpodyssey/spago/optimizer"
)

var _ optimizer.StrategyConfig = &Config{}

// Config provides configuration settings for Lamb optimizer.
type Config struct {
	optimizer.StrategyConfig
	StepSize float64
	Beta1    float64
	Beta2    float64
	Epsilon  float64
	Lambda   float64
}

// NewConfig returns a new Lamb Config.
func NewConfig(stepSize, beta1, beta2, epsilon, lambda float64) Config {
	if !(beta1 >= 0.0 && beta1 < 1.0) {
		panic("lamb: `beta1` must be in the range [0.0, 1.0)")
	}
	if !(beta2 >= 0.0 && beta2 < 1.0) {
		panic("lamb: `beta2` must be in the range [0.0, 1.0)")
	}
	return Config{
		StepSize: stepSize,
		Beta1:    beta1,
		Beta2:    beta2,
		Epsilon:  epsilon,
		Lambda:   lambda,
	}
}

// NewDefaultConfig returns a new Config with generically reasonable default values.
func NewDefaultConfig() Config {
	return Config{
		StepSize: 0.001,
		Beta1:    0.9,
		Beta2:    0.999,
		Epsilon:  1.0e-8,
		Lambda:   0.1,
	}
}

var _ optimizer.Strategy = &Lamb[float32]{}

// Lamb implements the Lamb gradient descent optimization method.
type Lamb[T float.DType] struct {
	Config
	Alpha    float64
	TimeStep int
}

// New returns a new Lamb optimizer, initialized according to the given configuration.
func New[T float.DType](c Config) *Lamb[T] {
	lamb := &Lamb[T]{
		Config: c,
		Alpha:  c.StepSize,
	}
	lamb.IncExample() // initialize 'alpha' coefficient
	return lamb
}

// Label returns the enumeration-like value which identifies this gradient descent method.
func (o *Lamb[_]) Label() int {
	return optimizer.Lamb
}

const (
	v    int = 0
	m    int = 1
	buf1 int = 2 // contains 'grads.ProdScalar(1.0 - beta1)'
	buf2 int = 3 // contains 'grads.Prod(grads).ProdScalar(1.0 - beta2)'
	buf3 int = 4
)

// NewSupport returns a new support structure with the given dimensions.
func (o *Lamb[T]) NewPayload(r, c int) *nn.OptimizerPayload {
	supp := make([]mat.Matrix, 5)
	supp[v] = mat.NewEmptyDense[T](r, c)
	supp[m] = mat.NewEmptyDense[T](r, c)
	supp[buf1] = mat.NewEmptyDense[T](r, c)
	supp[buf2] = mat.NewEmptyDense[T](r, c)
	supp[buf3] = mat.NewEmptyDense[T](r, c)
	return &nn.OptimizerPayload{
		Label: o.Label(),
		Data:  supp,
	}
}

// IncExample beats the occurrence of a new example.
func (o *Lamb[_]) IncExample() {
	o.TimeStep++
	o.updateAlpha()
}

func (o *Lamb[T]) updateAlpha() {
	ts := float64(o.TimeStep)
	o.Alpha = o.StepSize * math.Sqrt(1.0-math.Pow(o.Beta2, ts)) / (1.0 - math.Pow(o.Beta1, ts))
}

// CalcDelta returns the difference between the current params and where the method wants it to be.
func (o *Lamb[T]) CalcDelta(param *nn.Param) mat.Matrix {
	return o.calcDelta(param.Grad(), optimizer.GetOrSetPayload(param, o).Data, param.Value())
}

// v = v*beta1 + grads*(1.0-beta1)
// m = m*beta2 + (grads*grads)*(1.0-beta2)
// weights = ||params|| / || (v / (sqrt(m) + eps)) + (lambda * weights)
// d = (v / (sqrt(m) + eps)) + (lambda * weights) * alpha
func (o *Lamb[T]) calcDelta(grads mat.Matrix, supp []mat.Matrix, weights mat.Matrix) mat.Matrix {
	updateV(grads, supp, o.Beta1)
	updateM(grads, supp, o.Beta2)
	buf := supp[m].Sqrt().AddScalarInPlace(o.Epsilon)
	suppDiv := supp[v].Div(buf)
	if o.Lambda != 0.0 {
		scaledW := weights.ProdScalar(o.Lambda)
		suppDiv.AddInPlace(scaledW)
	}
	weightsNorm := norm(weights)
	adamStepNorm := norm(suppDiv)
	var trustRatio float64 = 1
	if !(weightsNorm == 0.0 || adamStepNorm == 0.0) {
		trustRatio = weightsNorm / adamStepNorm
	}
	supp[buf3].ProdMatrixScalarInPlace(suppDiv, o.Alpha*trustRatio)
	return supp[buf3]
}

// v = v*beta1 + grads*(1.0-beta1)
func updateV(grads mat.Matrix, supp []mat.Matrix, beta1 float64) {
	supp[v].ProdScalarInPlace(beta1)
	supp[buf1].ProdMatrixScalarInPlace(grads, 1.0-beta1)
	supp[v].AddInPlace(supp[buf1])
}

// m = m*beta2 + (grads*grads)*(1.0-beta2)
func updateM(grads mat.Matrix, supp []mat.Matrix, beta2 float64) {
	supp[m].ProdScalarInPlace(beta2)
	sqGrad := grads.Prod(grads)
	supp[buf2].ProdMatrixScalarInPlace(sqGrad, 1.0-beta2)
	supp[m].AddInPlace(supp[buf2])
}

func norm(grads mat.Matrix) float64 {
	prod := grads.Prod(grads)
	sum := prod.Sum()
	return math.Sqrt(sum.Scalar().F64())
}
