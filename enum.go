package enums

import (
	"sync"
)

type numeric interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64
}

func Register[S ~string, N numeric](dst *Enum[S, N], id string, definition map[S]N, options ...Option[S, N]) error {
	if dst == nil {
		return errFailedToRegister(id, "destination is required!")
	}
	if dst.Once == nil {
		return errFailedToRegister(id, "destination is required!")
	}
	// Setup destination
	if len(id) == 0 {
		return errFailedToRegister(id, "id is required!")
	}
	if len(definition) == 0 {
		return errFailedToRegister(id, "definition is required!")
	}
	i, names, indexes := 0, make([]S, len(definition)), make([]N, len(definition))
	for name, index := range definition {
		names[i] = name
		indexes[i] = index
		i++
	}
	if !isUnique(indexes) {
		return errFailedToRegister(id, "indexes must be unique!")
	}
	// Index
	dst.index.id = id
	dst.index.values = indexes
	dst.index.typeName = typeNameOf(indexes[0], true)
	// Description
	dst.desc.id = id
	dst.desc.typeName = typeNameOf(names[0], true)
	dst.desc.values = names
	// Options
	for n := range options {
		options[n](dst)
	}
	return nil
}

type Enum[S ~string, N numeric] struct {
	Once  *sync.Once
	index Index[N]
	desc  Description[S]
}

func (e Enum[_, N]) Index() Index[N] { return e.index }

func (e Enum[S, N]) IndexOf(desc S) N {
	for i := range e.desc.values {
		if desc == e.desc.values[i] {
			return e.index.values[i]
		}
	}
	return e.index.undefined
}

func (e Enum[S, _]) Desc() Description[S] { return e.desc }

func (e Enum[S, N]) Describe(index N) S {
	for i := range e.index.values {
		if index == e.index.values[i] {
			return e.desc.values[i]
		}
	}
	return e.desc.undefined
}
