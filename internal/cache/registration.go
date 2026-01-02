package cache

import "github.com/adityakw90/go-cache/internal/key"

// RegisterCacheKey registers a cache key for tracking.
// It stores the mapping: keyUsage[keyName] = []prefix
// This allows a single key to be registered with multiple prefixes.
func (c *Cache) RegisterCacheKey(keyName string, prefix string) {
	c.keyMutex.Lock()
	defer c.keyMutex.Unlock()

	// Check if the key is already registered
	if _, exists := c.keyUsage[keyName]; !exists {
		c.keyUsage[keyName] = []string{}
	}

	// Check if prefix already exists in the list
	for _, p := range c.keyUsage[keyName] {
		if p == prefix {
			return // Prefix is already registered
		}
	}

	// Register the new prefix
	c.keyUsage[keyName] = append(c.keyUsage[keyName], prefix)
}

// RegisterCustomKey registers a custom key function.
func (c *Cache) RegisterCustomKey(keyName string, customKeyFunc key.CustomKeyFunction) {
	if customKeyFunc == nil {
		return
	}

	c.keyMutex.Lock()
	defer c.keyMutex.Unlock()

	if _, exists := c.customKeys[keyName]; !exists {
		c.customKeys[keyName] = make(map[string]key.CustomKeyFunction)
	}

	c.customKeys[keyName][customKeyFunc.Name()] = customKeyFunc
}

// GetCacheKeyUsage returns the list of prefixes registered for a given key name.
// This is useful for debugging and monitoring cache key usage.
// Returns an empty slice if the key is not registered.
func (c *Cache) GetCacheKeyUsage(keyName string) []string {
	c.keyMutex.Lock()
	defer c.keyMutex.Unlock()

	if prefixes, exists := c.keyUsage[keyName]; exists {
		// Return a copy to prevent external modification
		result := make([]string, len(prefixes))
		copy(result, prefixes)
		return result
	}

	return []string{}
}
