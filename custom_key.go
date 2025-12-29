package cache

import "fmt"

// CustomKeyFunction defines custom key generation logic.
type CustomKeyFunction struct {
	Name     string                           // Name is the function name for registration.
	Callable func(args ...interface{}) string // Callable is the key generation function.
	Params   []string                         // Params are parameter names for documentation.
}

// Call calls the custom key function with the given parameters.
func (ckf *CustomKeyFunction) Call(params map[string]interface{}) (string, error) {
	var listArgs []interface{}

	for _, param := range ckf.Params {
		if val, exists := params[param]; exists {
			listArgs = append(listArgs, val)
		} else {
			return "", fmt.Errorf("required param %s missing", param)
		}
	}

	return ckf.Callable(listArgs...), nil
}
