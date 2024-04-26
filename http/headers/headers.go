package headers

import (
	"github.com/indigo-web/indigo/internal/keyvalue"
)

type (
	Header  = keyvalue.Pair
	Headers = *keyvalue.Storage
)

func New() Headers {
	return NewPrealloc(0)
}

func NewPrealloc(n int) Headers {
	return keyvalue.NewPreAlloc(n)
}

func NewFromMap(m map[string][]string) Headers {
	return keyvalue.NewFromMap(m)
}
