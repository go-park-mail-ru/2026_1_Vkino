package serializer

import (
	"bytes"
	"encoding/gob"
	"fmt"
)

func RegisterType(value any) {
	gob.Register(value)
}

func Serialize(value any) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)

	if err := enc.Encode(value); err != nil {
		return nil, fmt.Errorf("serialization failed: %w", err)
	}

	return buf.Bytes(), nil
}

func Deserialize(data []byte, value any) error {
	if value == nil {
		return fmt.Errorf("deserialization failed: nil pointer")
	}

	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)

	if err := dec.Decode(value); err != nil {
		return fmt.Errorf("deserialization failed: %w", err)
	}

	return nil
}
