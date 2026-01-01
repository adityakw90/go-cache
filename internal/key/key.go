package key

type CustomKeyFunction interface {
	Name() string
	Call(params map[string]interface{}) (string, error)
	Callable(args ...interface{}) string
}

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
