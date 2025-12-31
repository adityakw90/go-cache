package key

import (
	"errors"
	"fmt"
)

type CustomKeyFunction interface {
	Name() string
	Call(params map[string]interface{}) (string, error)
	Callable(args ...interface{}) string
}

// CustomKeyFunction defines custom key generation logic.
type customKeyFunction struct {
	name     string                           // Name is the function name for registration.
	callable func(args ...interface{}) string // Callable is the key generation function.
	params   []string                         // Params are parameter names for documentation.
}

func (ckf *customKeyFunction) Name() string {
	return ckf.name
}

// Call calls the custom key function with the given parameters.
func (ckf *customKeyFunction) Call(params map[string]interface{}) (string, error) {
	var listArgs []interface{}

	for _, param := range ckf.params {
		if val, exists := params[param]; exists {
			listArgs = append(listArgs, val)
		} else {
			return "", fmt.Errorf("required param %s missing", param)
		}
	}

	return ckf.callable(listArgs...), nil
}

func (ckf *customKeyFunction) Callable(args ...interface{}) string {
	return ckf.callable(args...)
}

func NewCustomKeyFunction(
	name string,
	callable func(args ...interface{}) string,
	params []string,
) (CustomKeyFunction, error) {
	if name == "" {
		return nil, errors.New("name is required")
	}
	if callable == nil {
		return nil, errors.New("callable is required")
	}
	if len(params) == 0 {
		return nil, errors.New("params is required")
	}
	return &customKeyFunction{
		name:     name,
		callable: callable,
		params:   params,
	}, nil
}
