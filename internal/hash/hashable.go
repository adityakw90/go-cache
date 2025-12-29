package hash

import (
	"bytes"
	"strconv"
)

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

