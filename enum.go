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
	enum.id = id
	enum.index.values = indexes
	enum.index.typeName = typeNameOf(indexes[0], true)
	enum.name.typeName = typeNameOf(names[0], true)
	enum.name.values = names
	for n := range options {
		options[n](&enum)
	}
	return enum
}

type Enum[N numeric, S ~string] struct {
	id    string
	index struct {
		typeName  string
		values    []N
		undefined N
	}
	name struct {
		typeName  string
		values    []S
		undefined S
	}
}

func (e Enum[N, _]) Equals(index, target N, or ...N) bool { return equals(index, target, or...) }

func (e Enum[N, _]) Indexes() []N { return e.index.values }

func (e Enum[N, S]) IndexOf(name S) N {
	for i := range e.name.values {
		if name == e.name.values[i] {
			return e.index.values[i]
		}
	}
	return e.index.undefined
}

func (e Enum[N, _]) Validate(index N) error {
	for i := range e.index.values {
		if index == e.index.values[i] {
			return nil
		}
	}
	return errInvalidValue(e.id, e.index.values, index)
}

func (e Enum[_, S]) NameEquals(name, target S, or ...S) bool { return equals(name, target, or...) }

func (e Enum[_, S]) Names() []S { return e.name.values }

func (e Enum[N, S]) NameOf(index N) S {
	for i := range e.index.values {
		if index == e.index.values[i] {
			return e.name.values[i]
		}
	}
	return e.name.undefined
}

func (e Enum[_, S]) ValidateName(name S) error {
	for i := range e.name.values {
		if name == e.name.values[i] {
			return nil
		}
	}
	return errInvalidValue(e.id, e.name.values, name)
}
