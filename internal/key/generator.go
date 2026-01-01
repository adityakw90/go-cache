package key

import "fmt"

// KeyGenerator generates a simple cache key.
// Template: {prefix}:{key}
func KeyGenerator(data map[string]string) (string, error) {
	prefix, ok := data["prefix"]
	if !ok {
		return "", ErrKeyGeneratorParamsPrefixRequired
	}
	key, ok := data["key"]
	if !ok {
		return "", ErrKeyGeneratorParamsKeyRequired
	}
	return prefix + ":" + key, nil
}

// KeyVersionGenerator generates a versioned cache key.
// Template: {prefix}:{namespace}:v{version}-{key}.gob
func KeyVersionGenerator(data map[string]string) (string, error) {
	prefix, ok := data["prefix"]
	if !ok {
		return "", ErrKeyGeneratorParamsPrefixRequired
	}
	namespace, ok := data["namespace"]
	if !ok {
		return "", ErrKeyGeneratorParamsNamespaceRequired
	}
	version, ok := data["version"]
	if !ok {
		return "", ErrKeyGeneratorParamsVersionRequired
	}
	key, ok := data["key"]
	if !ok {
		return "", ErrKeyGeneratorParamsKeyRequired
	}
	return fmt.Sprintf("%s:%s:v%s-%s.gob", prefix, namespace, version, key), nil
}

// VersionGenerator generates a version key.
// Template: {prefix}:{namespace}:version
func VersionGenerator(data map[string]string) (string, error) {
	prefix, ok := data["prefix"]
	if !ok {
		return "", ErrKeyGeneratorParamsPrefixRequired
	}
	namespace, ok := data["namespace"]
	if !ok {
		return "", ErrKeyGeneratorParamsNamespaceRequired
	}
	return fmt.Sprintf("%s:%s:version", prefix, namespace), nil
}

// LockGenerator generates a lock key.
// Template: {prefix}:{namespace}:lock
func LockGenerator(data map[string]string) (string, error) {
	prefix, ok := data["prefix"]
	if !ok {
		return "", ErrKeyGeneratorParamsPrefixRequired
	}
	namespace, ok := data["namespace"]
	if !ok {
		return "", ErrKeyGeneratorParamsNamespaceRequired
	}
	return fmt.Sprintf("%s:%s:lock", prefix, namespace), nil
}
