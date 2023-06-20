// Copyright 2019 spaGO Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package birnn

import (
	"encoding/gob"
	"sync"

	"github.com/nlpodyssey/spago/ag"
	"github.com/nlpodyssey/spago/mat"
	"github.com/nlpodyssey/spago/nn"
)

// MergeType is the enumeration-like type used for the set of merging methods
// which a BiRNN model Processor can perform.
type MergeType int

const (
	// Concat merging method: the outputs are concatenated together (the default)
	Concat MergeType = iota
	// Sum merging method: the outputs are added together
	Sum
	// Prod merging method: the outputs multiplied element-wise together
	Prod
	// Avg merging method: the average of the outputs is taken
	Avg
)

var _ nn.Model = &Model{}

// Model contains the serializable parameters.
type Model struct {
	nn.Module
	Positive  nn.StandardModel // positive time direction a.k.a. left-to-right
	Negative  nn.StandardModel // negative time direction a.k.a. right-to-left
	MergeMode MergeType
}

func init() {
	gob.Register(&Model{})
	gob.Register(&Model{})
}

// New returns a new model with parameters initialized to zeros.
func New(positive, negative nn.StandardModel, merge MergeType) *Model {
	return &Model{
		Positive:  positive,
		Negative:  negative,
		MergeMode: merge,
	}
}

// Forward performs the forward step for each input node and returns the result.
func (m *Model) Forward(xs ...mat.Tensor) []mat.Tensor {
	var pos []mat.Tensor
	var neg []mat.Tensor
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		pos = m.Positive.Forward(xs...)
	}()
	go func() {
		defer wg.Done()
		neg = m.Negative.Forward(reversed(xs)...)
	}()
	wg.Wait()
	out := make([]mat.Tensor, len(pos))
	for i := range out {
		out[i] = m.merge(pos[i], neg[len(out)-1-i])
	}
	return out
}

func reversed(ns []mat.Tensor) []mat.Tensor {
	r := make([]mat.Tensor, len(ns))
	copy(r, ns)
	for i := 0; i < len(r)/2; i++ {
		j := len(r) - i - 1
		r[i], r[j] = r[j], r[i]
	}
	return r
}

func (m *Model) merge(a, b mat.Tensor) mat.Tensor {
	switch m.MergeMode {
	case Concat:
		return ag.Concat(a, b)
	case Sum:
		return ag.Add(a, b)
	case Prod:
		return ag.Prod(a, b)
	case Avg:
		return ag.ProdScalar(ag.Add(a, b), a.Value().(mat.Matrix).NewScalar(0.5))
	default:
		panic("birnn: invalid merge mode")
	}
}
