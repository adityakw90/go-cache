package serialize

import "errors"

var (
	ErrSerializeNilValue    = errors.New("cannot serialize nil value")
	ErrDeserializeEmptyData = errors.New("cannot deserialize empty data")
	ErrDeserializeResultNil = errors.New("result interface cannot be nil")
)
