package enums

type Option[S ~string, N numeric] func(*Enum[S, N])

func SetUnknown[S ~string, N numeric](name S, index N) Option[S, N] {
	return func(e *Enum[S, N]) {
		e.named.unknown = name
		e.indexed.unknown = index
	}
}
