package hash

import (
	"bytes"
	"sort"
)

// handleMapString converts a map[string]interface{} to a hashable form.
// Keys are sorted to ensure consistent hashing.
func handleMapString(buffer *bytes.Buffer, m map[string]interface{}) {
	// Sort keys to ensure consistent hashable result
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	buffer.WriteString("{")

	for i, key := range keys {
		buffer.WriteString(key)
		buffer.WriteString(":")
		makeHashable(buffer, m[key])

		if i < len(keys)-1 {
			buffer.WriteString(",")
		}
	}
	buffer.WriteString("}")
}

// handleSlice converts a []interface{} to a hashable form.
func handleSlice(buffer *bytes.Buffer, s []interface{}) {
	buffer.WriteString("[")
	for i, item := range s {
		makeHashable(buffer, item)
		if i < len(s)-1 {
			buffer.WriteString(",")
		}
	}
	buffer.WriteString("]")
}

