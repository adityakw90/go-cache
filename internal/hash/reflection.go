package hash

import (
	"bytes"
	"fmt"
	"reflect"
	"sort"
)

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

	firstWritten := false
	for i := 0; i < numFields; i++ {
		field := obj.Field(i)

		// Skip unexported fields
		if !field.CanInterface() {
			continue
		}

		// Add a comma if we've already written at least one field
		if firstWritten {
			buffer.WriteString(",")
		}

		// Write field name and value to the buffer
		buffer.WriteString(objType.Field(i).Name)
		buffer.WriteString(":")
		makeHashable(buffer, field.Interface())

		// Mark that we've written at least one field
		firstWritten = true
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
