package hash

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
)

// CacheKey generates an MD5 hash from function name and arguments.
func CacheKey(funcName string, args []interface{}) string {
	// Retrieve a buffer from the pool
	buffer := bufferPool.Get().(*bytes.Buffer)
	defer bufferPool.Put(buffer)

	// Reset the buffer to reuse it
	buffer.Reset()

	// Write the function name first
	buffer.WriteString(funcName)
	buffer.WriteString(":") // separator

	// Hash the arguments using makeHashable
	makeHashable(buffer, args)

	// Create the MD5 hash directly from the buffer bytes
	hash := md5.Sum(buffer.Bytes())

	// Return the hash as a hex string
	return hex.EncodeToString(hash[:])
}
