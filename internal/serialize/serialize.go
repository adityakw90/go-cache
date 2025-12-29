package serialize

import (
	"bytes"
	"encoding/gob"

	"github.com/adityakw90/go-cache/internal/errs"
)

// Serialize converts a value to []byte using Gob encoding.
func Serialize(value interface{}) ([]byte, error) {
	if value == nil {
		return nil, errs.ErrSerializeNilValue
	}

	var buffer bytes.Buffer
	enc := gob.NewEncoder(&buffer)

	if err := enc.Encode(value); err != nil {
		return nil, errs.Wrap(err, "failed to serialize value")
	}

	return buffer.Bytes(), nil
}

// Deserialize converts []byte to a value using Gob decoding.
func Deserialize(data []byte, result interface{}) error {
	if len(data) == 0 {
		return errs.ErrDeserializeEmptyData
	}

	if result == nil {
		return errs.ErrDeserializeResultNil
	}

	buffer := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buffer)

	if err := dec.Decode(result); err != nil {
		return errs.Wrap(err, "failed to deserialize data")
	}

	return nil
}
