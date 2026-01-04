package serialize

import (
	"bytes"
	"encoding/gob"
	"fmt"
)

// Serialize converts a value to []byte using Gob encoding.
func Serialize(value interface{}) ([]byte, error) {
	if value == nil {
		return nil, ErrSerializeNilValue
	}

	var buffer bytes.Buffer
	enc := gob.NewEncoder(&buffer)

	if err := enc.Encode(value); err != nil {
		return nil, fmt.Errorf("failed to serialize value: %w", err)
	}

	return buffer.Bytes(), nil
}

// Deserialize converts []byte to a value using Gob decoding.
func Deserialize(data []byte, resultPtr interface{}) error {
	if len(data) == 0 {
		return ErrDeserializeEmptyData
	}

	if resultPtr == nil {
		return ErrDeserializeResultNil
	}

	buffer := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buffer)

	if err := dec.Decode(resultPtr); err != nil {
		return fmt.Errorf("failed to deserialize data: %w", err)
	}

	return nil
}
