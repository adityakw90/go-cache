package key

import "errors"

// custom key function errors
var ErrNameRequired = errors.New("name is required")
var ErrCallableRequired = errors.New("callable is required")
var ErrParamsRequired = errors.New("params is required")

// key generator errors
var ErrKeyGeneratorParamsPrefixRequired = errors.New("missing 'prefix' in key generator data")
var ErrKeyGeneratorParamsKeyRequired = errors.New("missing 'key' in key generator data")
var ErrKeyGeneratorParamsNamespaceRequired = errors.New("missing 'namespace' in key generator data")
var ErrKeyGeneratorParamsVersionRequired = errors.New("missing 'version' in key generator data")
