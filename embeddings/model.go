// Copyright 2022 spaGO Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package embeddings

import (
	"encoding/gob"
	"fmt"
	"sync"

	"github.com/nlpodyssey/spago/ag"
	"github.com/nlpodyssey/spago/embeddings/store"
	"github.com/nlpodyssey/spago/mat"
	"github.com/nlpodyssey/spago/nn"
)

// Config provides configuration settings for an embeddings Model.
type Config struct {
	// Size of the embedding vectors.
	Size int
	// Whether to return the `ZeroEmbedding` in case the key doesn't exist in
	// the embeddings store.
	// If it is false, nil is returned instead, so the caller has more
	// responsibility but also more control.
	UseZeroEmbedding bool
	// The name of the store.Store get from a store.Repository for the
	// data handled by the embeddings model.
	StoreName string
	// A trainable model allows its Embedding parameters to have gradient
	// values that can be propagated. When set to false, gradients handling is
	// disabled.
	Trainable bool
}

// A Model for handling embeddings.
type Model[T mat.DType, K Key] struct {
	nn.BaseModel[T]
	Config
	// ZeroEmbedding is used as a fallback value for missing embeddings.
	//
	// If Config.UseZeroEmbedding is true, ZeroEmbedding is initialized
	// as a zero-vector of size Config.Size, otherwise it is set to nil.
	ZeroEmbedding nn.Param[T] `spago:"type:weights"`
	// Database where embeddings are stored.
	Store store.Store
	// EmbeddingsWithGrad is filled with all those embedding-parameters that
	// have a gradient value attached.
	//
	// This map is public, so that all parameters with a gradient can be
	// easily traversed, which is useful in contexts such as optimizers.
	// This value is managed by the Model and its derived Embedding parameters
	// and should be considered read-only for external access.
	//
	// Parameters whose gradient is "zeroed" (see ag.GradValue.ZeroGrad) are
	// automatically removed from this map, thus freeing resources that would
	// otherwise be kept in memory indefinitely.
	//
	// In many simple use cases, as long as gradients are regularly zeroed,
	// the automatic mechanism described above will prevent the memory from
	// being cluttered with too many unused values.
	//
	// For other special or peculiar usages, for example if gradients are not
	// cleared regularly or at all, you might need to clear this map explicitly,
	// by calling the dedicated method ClearEmbeddingsWithGrad (again: never
	// modify this value directly).
	//
	// The type of this field, ParamsMap, is defined in order to be traversable,
	// but to produce no data when serialized.
	EmbeddingsWithGrad ParamsMap[T]
	// This map is maintained in parallel with EmbeddingsWithGrad.
	// An Embedding parameter doesn't keep any internal value; everything is
	// rather delegated to the model, or the model's store.
	// The Embedding values stored in EmbeddingsWithGrad don't contain any
	// gradient value; instead the Model provides private methods allowing
	// reading and writing gradients by key, which are stored here.
	grads map[string]mat.Matrix[T]
	mu    sync.RWMutex
}

func init() {
	gob.Register(&Model[float32, string]{})
	gob.Register(&Model[float32, []byte]{})
	gob.Register(&Model[float64, string]{})
	gob.Register(&Model[float64, []byte]{})
}

// New returns a new embeddings Model.
//
// It panics in case of errors getting the Store from the Repository.
func New[T mat.DType, K Key](conf Config, repo store.Repository) *Model[T, K] {
	st, err := repo.Store(conf.StoreName)
	if err != nil {
		panic(fmt.Errorf("embeddings: error getting Store %#v: %w", conf.StoreName, err))
	}

	var zeroEmb nn.Param[T] = nil
	if conf.UseZeroEmbedding {
		zeroEmb = nn.NewParam[T](mat.NewEmptyVecDense[T](conf.Size), nn.RequiresGrad[T](false))
	}
	return &Model[T, K]{
		Config:        conf,
		ZeroEmbedding: zeroEmb,
		Store:         &store.PreventStoreMarshaling{Store: st},
	}
}

// Count counts how many embedding key/value pairs are currently stored.
// It panics in case of reading errors.
func (m *Model[_, _]) Count() int {
	n, err := m.Store.KeysCount()
	if err != nil {
		panic(fmt.Errorf("embeddings: error counting keys in store: %w", err))
	}
	return n
}

// Embedding returns the Embedding parameter associated with the given key,
// also reporting whether the key was found in the store.
//
// Even if an embedding parameter is not found in the store, a usable value
// is still returned; it's sufficient to set some data on it (value, payload)
// to trigger its creation on the store.
//
// It panics in case of errors reading from the underlying store.
func (m *Model[T, K]) Embedding(key K) (nn.Param[T], bool) {
	if e, ok := m.EmbeddingsWithGrad[stringifyKey(key)]; ok {
		return e, true
	}

	exists, err := m.Store.Contains(encodeKey(key))
	if err != nil {
		panic(err)
	}
	e := &Embedding[T, K]{
		model: m,
		key:   key,
	}
	return e, exists
}

// Encode returns the embedding values associated with the input keys.
//
// The value are returned as Node(s) already inserted in the graph.
//
// Missing embedding values can be either nil or ZeroEmbedding, according
// to the Model's Config.
func (m *Model[T, K]) Encode(keys []K) []ag.Node[T] {
	nodes := make([]ag.Node[T], len(keys))

	// reuse the same node for the same key
	cache := make(map[string]ag.Node[T], len(keys))

	for i, key := range keys {
		strKey := stringifyKey(key)

		if v, ok := cache[strKey]; ok {
			nodes[i] = v
			continue
		}

		var n ag.Node[T]
		if e, ok := m.Embedding(key); ok {
			n = ag.NewWrap[T](e)
		} else {
			n = m.ZeroEmbedding
		}

		nodes[i] = n
		cache[strKey] = n
	}
	return nodes
}

// ClearEmbeddingsWithGrad empties the memory of visited embeddings with
// non-null gradient value.
func (m *Model[_, _]) ClearEmbeddingsWithGrad() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.grads = nil
	m.EmbeddingsWithGrad = nil
}

// UseRepository allows the Model to use a Store from the given Repository.
//
// It only works if a store is not yet present. This can only happen in
// special situations, for example upon an Embedding model being deserialized,
// or when manually instantiating and handling a Model (i.e. bypassing New).
func (m *Model[_, _]) UseRepository(repo store.Repository) error {
	st, err := repo.Store(m.StoreName)
	if err != nil {
		return err
	}
	if m.storeExists() {
		return fmt.Errorf("a Store is already set on this embeddings.Model")
	}
	m.Store = &store.PreventStoreMarshaling{Store: st}
	return nil
}

func (m *Model[_, _]) storeExists() bool {
	switch s := m.Store.(type) {
	case nil:
		return false
	case store.PreventStoreMarshaling:
		return s.Store != nil
	case *store.PreventStoreMarshaling:
		return s.Store != nil
	default:
		return true
	}
}

func (m *Model[T, K]) getGrad(key K) (grad mat.Matrix[T], exists bool) {
	if !m.Trainable {
		return nil, false
	}

	m.mu.RLock()
	grad, exists = m.grads[stringifyKey(key)]
	m.mu.RUnlock()
	return
}

func (m *Model[T, K]) accGrad(e *Embedding[T, K], gx mat.Matrix[T]) {
	if !m.Trainable {
		return
	}
	key := stringifyKey(e.key)

	m.mu.Lock()
	defer m.mu.Unlock()

	grad, exists := m.grads[key]
	if exists {
		grad.AddInPlace(gx)
		return
	}

	if m.grads == nil {
		m.grads = make(map[string]mat.Matrix[T])
		m.EmbeddingsWithGrad = make(ParamsMap[T])
	}
	m.grads[key] = gx.Clone()
	m.EmbeddingsWithGrad[key] = e
}

func (m *Model[T, K]) zeroGrad(k K) {
	if !m.Trainable {
		return
	}
	key := stringifyKey(k)

	m.mu.Lock()
	defer m.mu.Unlock()

	grad, exists := m.grads[key]
	if !exists {
		return
	}

	mat.ReleaseMatrix(grad)
	delete(m.grads, key)
	delete(m.EmbeddingsWithGrad, key)
}
