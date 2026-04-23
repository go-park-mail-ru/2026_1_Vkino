package errmap

type Matcher[S any, K comparable] func(subject S, key K) bool

type Mapper[S any, K comparable, R any] struct {
	order []K
	rules map[K]R
	match Matcher[S, K]
}

func New[S any, K comparable, R any](order []K, rules map[K]R, match Matcher[S, K]) *Mapper[S, K, R] {
	return &Mapper[S, K, R]{
		order: order,
		rules: rules,
		match: match,
	}
}

func (m *Mapper[S, K, R]) Resolve(subject S) (R, bool) {
	var zero R

	for _, key := range m.order {
		if m.match(subject, key) {
			rule, ok := m.rules[key]
			if !ok {
				return zero, false
			}

			return rule, true
		}
	}

	return zero, false
}
