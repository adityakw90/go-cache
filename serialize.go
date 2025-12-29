package cache

import (
	"bytes"
	"encoding/gob"
	"fmt"
)

// serialize converts a value to []byte using Gob encoding.
func (c *Cache) serialize(value interface{}) ([]byte, error) {
	if value == nil {
		return nil, fmt.Errorf("cannot serialize nil value")
	}

	var buffer bytes.Buffer
	enc := gob.NewEncoder(&buffer)

	if err := enc.Encode(value); err != nil {
		return nil, fmt.Errorf("failed to serialize value: %w", err)
	}

	return buffer.Bytes(), nil
}

// deserialize converts []byte to a value using Gob decoding.
func (c *Cache) deserialize(data []byte, result interface{}) error {
	if len(data) == 0 {
		return fmt.Errorf("cannot deserialize empty data")
	}

	if result == nil {
		return fmt.Errorf("result interface cannot be nil")
	}

	buffer := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buffer)

	if err := dec.Decode(result); err != nil {
		return fmt.Errorf("failed to deserialize data: %w", err)
	}

	return nil
}

