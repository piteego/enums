package enums

type (
	numeric interface {
		~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64
	}
	enumable               interface{ numeric | ~string }
	Enumerator[E enumable] interface {
		List() []E
		Validate(E) error
	}
)

type named[S ~string] struct {
	id, typeName string
	unknown      S
	values       []S
}

func (x named[S]) List() []S { return x.values }

func (x named[S]) Validate(name S) error {
	for i := range x.values {
		if name == x.values[i] {
			return nil
		}
	}
	return errInvalidValue(x.id, x.values, name)
}

type indexed[N numeric] struct {
	id, typeName string
	unknown      N
	values       []N
}

func (x indexed[N]) List() []N { return x.values }

func (x indexed[N]) Validate(index N) error {
	for i := range x.values {
		if index == x.values[i] {
			return nil
		}
	}
	return errInvalidValue(x.id, x.values, index)
}
