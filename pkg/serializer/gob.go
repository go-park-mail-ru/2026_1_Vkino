package serializer

import (
	"bytes"
	"encoding/gob"
	"errors"
)

var (
	ErrDeserialize = errors.New("deserialization failed")
	ErrSerialize   = errors.New("serialization failed")
)

func RegisterType(value any) {
	gob.Register(value)
}

func Serialize[T any](v T) ([]byte, error) {
	var buf bytes.Buffer

	enc := gob.NewEncoder(&buf)

	if err := enc.Encode(v); err != nil {
		return nil, ErrSerialize
	}

	return buf.Bytes(), nil
}

func Deserialize[T any](data []byte, v *T) error {
	if v == nil {
		return ErrDeserialize
	}

	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)

	if err := dec.Decode(v); err != nil {
		return ErrDeserialize
	}

	return nil
}
