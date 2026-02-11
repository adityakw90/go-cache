package key

// KeyGeneratorFunc is a function type that generates a cache key from a map of data.
// It is used by the cache system to generate consistent, predictable keys based on
// configurable parameters like prefix, namespace, version, and custom key data.
type KeyGeneratorFunc func(data map[string]string) (string, error)

// CustomKeyFunction defines an interface for custom cache key generation logic.
// Implementations can provide specialized key generation based on function arguments
// and named parameters.
type CustomKeyFunction interface {
	// Name returns the unique identifier for this custom key function.
	Name() string

	// Call generates a cache key using the provided parameter values.
	// Parameters are matched by name to the function's expected parameters.
	Call(params map[string]interface{}) (string, error)

	// Callable generates a cache key directly from function arguments.
	// This is used when the key function is called internally by the cache system.
	Callable(args ...interface{}) string
}

// NewCustomKeyFunction creates a new CustomKeyFunction with the specified name,
// callable function, and parameter names.
// Parameters:
//   - name: unique identifier for the custom key function
//   - callable: function that generates a cache key from arguments
//   - params: ordered list of parameter names expected by the callable function
//
// Returns:
//   - CustomKeyFunction implementation
//   - error if name is empty, callable is nil, or params is empty
func NewCustomKeyFunction(
	name string,
	callable func(args ...interface{}) string,
	params []string,
) (CustomKeyFunction, error) {
	if name == "" {
		return nil, ErrNameRequired
	}
	if callable == nil {
		return nil, ErrCallableRequired
	}
	if len(params) == 0 {
		return nil, ErrParamsRequired
	}
	return &customKeyFunction{
		name:     name,
		callable: callable,
		params:   params,
	}, nil
}
