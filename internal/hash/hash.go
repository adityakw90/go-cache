package hash

import (
	"crypto/md5"
	"encoding/hex"
)

// CacheKey generates an MD5 hash from function name and arguments.
func CacheKey(funcName string, args []interface{}) string {
	buffer := getBuffer()
	defer putBuffer(buffer)

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
