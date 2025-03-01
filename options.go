package enums

type Option[S ~string, N numeric] func(*Enum[S, N])

func UndefinedValues[S ~string, N numeric](name S, index N) Option[S, N] {
	return func(e *Enum[S, N]) {
		e.desc.undefined = name
		e.index.undefined = index
	}
}
