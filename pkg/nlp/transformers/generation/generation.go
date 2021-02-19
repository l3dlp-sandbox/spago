// Copyright 2021 spaGO Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// Additional copyright notes in the package README.

// Package generation implements a generation search algorithm for conditional generation.
package generation

import (
	mat "github.com/nlpodyssey/spago/pkg/mat32"
	"github.com/nlpodyssey/spago/pkg/ml/ag"
	"github.com/nlpodyssey/spago/pkg/utils/processingqueue"
	"math"
	"sort"
	"sync"
)

// Generator is an implementation of a generation search algorithm for conditional generation.
type Generator struct {
	config          GeneratorConfig
	model           EncoderDecoder
	processingQueue processingqueue.ProcessingQueue
}

// NewGenerator creates a new Generator object.
func NewGenerator(config GeneratorConfig, model EncoderDecoder) *Generator {
	return &Generator{
		config:          config,
		model:           model,
		processingQueue: processingqueue.New(config.MaxConcurrentComputations),
	}
}

// Generate generates sequences for models with a language modeling head, using
// generation-search decoding.
func (b *Generator) Generate(inputIDs []int) []int {
	if !b.config.IsEncoderDecoder {
		panic("generator: unsupported architecture")
	}

	encodedInput := b.model.Encode(inputIDs)
	if !b.config.IncrementalForward {
		b.performForward()
	}

	return b.beamSearch(NewScorer(b.config), encodedInput)
}

func (b *Generator) beamSearch(scorer *Scorer, encodedInput []ag.Node) []int {
	var (
		numBeams         = b.config.NumBeams
		beamScores       = b.makeInitBeamScores()
		decodingInputIDs = b.makeStartDecodingInputForBeamDecoding()
		scores           = make([]Scores, numBeams)
		cache            = make([]Cache, numBeams)
		curLen           = len(decodingInputIDs[0])
	)

	for curLen < b.config.MaxLength {
		scores, cache = b.generateNext(encodedInput, decodingInputIDs, cache)
		nextTokenScores := b.inhibitInvalidTokens(decodingInputIDs, scores)
		updateTokensScores(nextTokenScores, beamScores)
		scoredTokens := b.makeScoredTokens(nextTokenScores)
		beamOutputs := scorer.Process(decodingInputIDs, scoredTokens)
		beamScores = beamOutputs.nextBeamScores
		decodingInputIDs = makeNewInputIDs(decodingInputIDs, beamOutputs)
		cache = reorderCache(cache, beamOutputs.nextBeamIndices)

		if scorer.IsDone() {
			break
		}
		curLen++
	}

	return scorer.Finalize(decodingInputIDs, beamScores)
}

func (b *Generator) generateNext(
	encodedInput []ag.Node,
	decodingInputIDs [][]int,
	pastCache []Cache,
) ([]Scores, []Cache) {
	numBeams := b.config.NumBeams
	logProbs := make([]ag.Node, numBeams)
	scores := make([]Scores, numBeams)
	nextCache := make([]Cache, numBeams)

	var wg sync.WaitGroup
	wg.Add(numBeams)
	for i := 0; i < numBeams; i++ {
		i := i // redefine `i` in the inner scope, for using it in the goroutine
		b.processingQueue.Go(func() {
			defer wg.Done()
			logProbs[i], nextCache[i] = b.model.Decode(encodedInput, decodingInputIDs[i], pastCache[i])
		})
	}
	wg.Wait()

	if !b.config.IncrementalForward {
		b.performForward()
	}

	for i := 0; i < numBeams; i++ {
		scores[i] = b.model.Graph().GetCopiedValue(logProbs[i])
	}

	return scores, nextCache
}

func (b *Generator) performForward() {
	g := b.model.Graph()
	g.Forward(ag.Range(g.TimeStep(), -1))
	g.IncTimeStep() // mark the next block to be computed from here on
}

func (b *Generator) makeScoredTokens(tokensScores []Scores) ScoredTokens {
	resultSize := b.config.NumBeams * 2
	result := make(ScoredTokens, 0, resultSize+1)

	var currentMinValue mat.Float = -math.MaxFloat32
	var currentMinIndex int

	for beamIndex, n := range tokensScores {
		for tokenIndex, score := range n.Data() {
			if len(result) < resultSize || score > currentMinValue {
				result = append(result, &ScoredToken{
					BeamIndex:  beamIndex,
					TokenIndex: tokenIndex,
					Score:      score,
				})
			}
			if len(result) > resultSize {
				result = append(result[:currentMinIndex], result[currentMinIndex+1:]...)
			}
			currentMinValue = math.MaxFloat32
			for ri, rv := range result {
				if rv.Score < currentMinValue {
					currentMinValue = rv.Score
					currentMinIndex = ri
				}
			}
		}
	}

	sort.SliceStable(result, func(i, j int) bool {
		return result[i].Score > result[j].Score
	})
	return result
}

func updateTokensScores(tokensScores []Scores, beamScores []mat.Float) {
	for i, bs := range beamScores {
		v := tokensScores[i]
		for j, f := range v.Data() {
			v.SetVec(j, f+bs)
		}
	}
}

func makeNewInputIDs(inputIDs [][]int, scorerOut ScorerProcessOutput) [][]int {
	newInputIDs := make([][]int, len(inputIDs))
	for i, beamIndex := range scorerOut.nextBeamIndices {
		prevValue := inputIDs[beamIndex]
		newInputIDs[i] = make([]int, 0, len(prevValue)+1)
		newInputIDs[i] = append(newInputIDs[i], prevValue...)
		newInputIDs[i] = append(newInputIDs[i], scorerOut.nextBeamTokens[i])
	}
	return newInputIDs
}

func reorderCache(cache []Cache, nextBeamIndices []int) []Cache {
	reorderedCache := make([]Cache, len(cache))
	for i, beamIndex := range nextBeamIndices {
		reorderedCache[i] = cache[beamIndex]
	}
	return reorderedCache
}

func (b *Generator) makeStartDecodingInputForBeamDecoding() [][]int {
	beamInputIDs := make([][]int, b.config.NumBeams)
	for i := range beamInputIDs {
		beamInputIDs[i] = []int{b.config.DecoderStartTokenID}
	}
	return beamInputIDs
}

func (b *Generator) makeInitBeamScores() []mat.Float {
	numBeams := b.config.NumBeams
	beamScores := make([]mat.Float, numBeams)
	beamScores[0] = 0
	for i := 1; i < numBeams; i++ {
		beamScores[i] = -1e9
	}
	return beamScores
}
