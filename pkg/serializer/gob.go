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

func Serialize(value any) ([]byte, error) {
	var buf bytes.Buffer

	enc := gob.NewEncoder(&buf)

	if err := enc.Encode(value); err != nil {
		return nil, ErrSerialize
	}

	return buf.Bytes(), nil
}

func Deserialize(data []byte, value any) error {
	if value == nil {
		return ErrDeserialize
	}

	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)

	if err := dec.Decode(value); err != nil {
		return ErrDeserialize
	}

	return nil
}
