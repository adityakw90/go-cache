package cache

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"sync"
)

// bufferPool is a pool for reusing bytes.Buffer to reduce allocations.
var bufferPool = sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}

// getCacheHash generates an MD5 hash from function name and arguments.
func (c *Cache) getCacheHash(funcName string, args []interface{}) string {
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

// makeHashable recursively converts objects to a hashable form.
// It handles maps, slices, structs, pointers, and simple types.
func makeHashable(buffer *bytes.Buffer, obj interface{}) {
	switch v := obj.(type) {
	case map[string]interface{}:
		handleMapString(buffer, v)
	case []interface{}:
		handleSlice(buffer, v)
	case string:
		buffer.WriteString(v)
	case int:
		buffer.Write(strconv.AppendInt(nil, int64(v), 10))
	case int64:
		buffer.Write(strconv.AppendInt(nil, v, 10))
	case int32:
		buffer.Write(strconv.AppendInt(nil, int64(v), 10))
	case int16:
		buffer.Write(strconv.AppendInt(nil, int64(v), 10))
	case int8:
		buffer.Write(strconv.AppendInt(nil, int64(v), 10))
	case uint:
		buffer.Write(strconv.AppendUint(nil, uint64(v), 10))
	case uint64:
		buffer.Write(strconv.AppendUint(nil, v, 10))
	case uint32:
		buffer.Write(strconv.AppendUint(nil, uint64(v), 10))
	case uint16:
		buffer.Write(strconv.AppendUint(nil, uint64(v), 10))
	case uint8:
		buffer.Write(strconv.AppendUint(nil, uint64(v), 10))
	case float64:
		buffer.Write(strconv.AppendFloat(nil, v, 'f', -1, 64))
	case float32:
		buffer.Write(strconv.AppendFloat(nil, float64(v), 'f', -1, 32))
	case bool:
		buffer.Write(strconv.AppendBool(nil, v))
	case nil:
		buffer.WriteString("<nil>")
	default:
		// Handle reflection for pointers, structs, etc.
		handleReflection(buffer, obj)
	}
}

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

// handleReflection handles complex types using reflection.
func handleReflection(buffer *bytes.Buffer, obj interface{}) {
	val := reflect.ValueOf(obj)

	switch val.Kind() {
	case reflect.Ptr:
		handlePointer(buffer, val)
	case reflect.Struct:
		handleStruct(buffer, val)
	case reflect.Slice, reflect.Array:
		handleReflectionSlice(buffer, val)
	case reflect.Map:
		handleReflectionMap(buffer, val)
	default:
		buffer.WriteString(fmt.Sprintf("%v", val.Interface())) // Fallback for unsupported types
	}
}

// handlePointer handles pointer types.
func handlePointer(buffer *bytes.Buffer, obj reflect.Value) {
	if obj.IsNil() {
		buffer.WriteString("<nil>")
		return
	}
	makeHashable(buffer, obj.Elem().Interface())
}

// handleStruct handles struct types.
func handleStruct(buffer *bytes.Buffer, obj reflect.Value) {
	objType := obj.Type()
	numFields := obj.NumField()

	buffer.WriteString("{")

	for i := 0; i < numFields; i++ {
		field := obj.Field(i)

		// Skip unexported fields
		if !field.CanInterface() {
			continue
		}

		// Write field name and value to the buffer
		buffer.WriteString(objType.Field(i).Name)
		buffer.WriteString(":")
		makeHashable(buffer, field.Interface())

		// Add a comma if it's not the last field
		if i < numFields-1 {
			buffer.WriteString(",")
		}
	}
	buffer.WriteString("}")
}

// handleReflectionSlice handles slice and array types via reflection.
func handleReflectionSlice(buffer *bytes.Buffer, val reflect.Value) {
	buffer.WriteString("[")
	for i := 0; i < val.Len(); i++ {
		makeHashable(buffer, val.Index(i).Interface())
		if i < val.Len()-1 {
			buffer.WriteString(",")
		}
	}
	buffer.WriteString("]")
}

// handleReflectionMap handles map types via reflection.
func handleReflectionMap(buffer *bytes.Buffer, val reflect.Value) {
	keys := val.MapKeys()
	// Sort keys for consistent hashing
	sort.Slice(keys, func(i, j int) bool {
		return fmt.Sprintf("%v", keys[i].Interface()) < fmt.Sprintf("%v", keys[j].Interface())
	})

	buffer.WriteString("{")
	for i, key := range keys {
		buffer.WriteString(fmt.Sprintf("%v", key.Interface()))
		buffer.WriteString(":")
		makeHashable(buffer, val.MapIndex(key).Interface())
		if i < len(keys)-1 {
			buffer.WriteString(",")
		}
	}
	buffer.WriteString("}")
}
