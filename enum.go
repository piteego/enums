package enums

type Enum[S ~string, N numeric] struct {
	named   named[S]
	indexed indexed[N]
}

func (e Enum[_, N]) Index() Enumerator[N] { return e.indexed }

func (e Enum[S, N]) IndexOf(name S) N {
	for i := range e.named.values {
		if name == e.named.values[i] {
			return e.indexed.values[i]
		}
	}
	return e.indexed.unknown
}

func (e Enum[S, _]) Name() Enumerator[S] { return e.named }

func (e Enum[S, N]) NameOf(index N) S {
	for i := range e.indexed.values {
		if index == e.indexed.values[i] {
			return e.named.values[i]
		}
	}
	return e.named.unknown
}
