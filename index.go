package enums

type Index[N numeric] struct {
	id, typeName string
	undefined    N
	values       []N
}

func (x Index[N]) Id() string { return x.id }

func (x Index[N]) Is(index, target N, or ...N) bool { return is(index, target, or...) }

func (x Index[N]) List() []N { return x.values }

func (x Index[N]) Validate(index N) error {
	for i := range x.values {
		if index == x.values[i] {
			return nil
		}
	}
	return errInvalidValue(x.id, x.values, index)
}
