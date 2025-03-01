package enums

type numeric interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64
}

func New[N numeric, S ~string](id string, definition map[N]S, options ...Option[N, S]) Enum[N, S] {
	if len(id) == 0 {
		panic("[Enum] failed to create new enum: id is required")
	}
	if len(definition) == 0 {
		panic("[Enum] failed to create new enum: definition is required")
	}
	i, names, indexes := 0, make([]S, len(definition)), make([]N, len(definition))
	for k, v := range definition {
		indexes[i] = k
		names[i] = v
		i++
	}
	if !isUnique(names) {
		panic("[Enum] failed to create new enum: names must be unique")
	}
	var enum Enum[N, S]
	// Index
	enum.index.id = id
	enum.index.values = indexes
	enum.index.typeName = typeNameOf(indexes[0], true)
	// Description
	enum.desc.id = id
	enum.desc.typeName = typeNameOf(names[0], true)
	enum.desc.values = names
	// Options
	for n := range options {
		options[n](&enum)
	}
	return enum
}

type Enum[N numeric, S ~string] struct {
	index Index[N]
	desc  Description[S]
}

func (e Enum[N, _]) Index() Index[N] { return e.index }

func (e Enum[N, S]) IndexOf(desc S) N {
	for i := range e.desc.values {
		if desc == e.desc.values[i] {
			return e.index.values[i]
		}
	}
	return e.index.undefined
}

func (e Enum[N, S]) Desc() Description[S] { return e.desc }

func (e Enum[N, S]) Describe(index N) S {
	for i := range e.index.values {
		if index == e.index.values[i] {
			return e.desc.values[i]
		}
	}
	return e.desc.undefined
}
