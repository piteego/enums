package enums

type Option[N numeric, S ~string] func(*Enum[N, S])

func WithUndefinedValues[N numeric, S ~string](index N, name S) Option[N, S] {
	return func(e *Enum[N, S]) {
		e.index.undefined = index
		e.desc.undefined = name
	}
}
