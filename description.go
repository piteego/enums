package enums

type Description[S ~string] struct {
	id, typeName string
	values       []S
	undefined    S
}

func (d Description[S]) Id() string { return d.id }

func (d Description[S]) Is(desc, target S, or ...S) bool { return is(desc, target, or...) }

func (d Description[S]) List() []S { return d.values }

func (d Description[S]) Validate(name S) error {
	for i := range d.values {
		if name == d.values[i] {
			return nil
		}
	}
	return errInvalidValue(d.id, d.values, name)
}
