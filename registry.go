package enums

import "sync"

type Registry[S ~string, N numeric] struct {
	sync.Once
	Enum Enum[S, N]
}

func Register[S ~string, N numeric](dst *Enum[S, N], id string, definition map[S]N, options ...Option[S, N]) error {
	if dst == nil {
		return errFailedToRegister(id, "destination is required!")
	}
	if len(id) == 0 {
		return errFailedToRegister(id, "id is required!")
	}
	if len(definition) == 0 {
		return errFailedToRegister(id, "definition is required!")
	}
	i, names, indexes := 0, make([]S, len(definition)), make([]N, len(definition))
	for name, value := range definition {
		names[i] = name
		indexes[i] = value
		i++
	}
	if !isUnique(indexes) {
		return errFailedToRegister(id, "indexes must be unique!")
	}
	// Name
	dst.named.id = id
	dst.named.typeName = typeNameOf(names[0], true)
	dst.named.values = names
	// Index
	dst.indexed.id = id
	dst.indexed.values = indexes
	dst.indexed.typeName = typeNameOf(indexes[0], true)
	// Options
	for n := range options {
		options[n](dst)
	}
	return nil
}
