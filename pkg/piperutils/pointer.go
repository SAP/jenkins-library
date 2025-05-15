package piperutils

func SafeDereference[T any](p *T) T {
	if p == nil {
		var zeroValue T
		return zeroValue
	}

	return *p
}
