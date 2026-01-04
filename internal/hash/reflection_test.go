package hash

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestHash_HandleReflection_Struct tests handleReflection with struct types.
func TestHash_HandleReflection_Struct(t *testing.T) {
	type TestStruct struct {
		Name string
		Age  int
	}

	tests := []struct {
		name      string
		value     interface{}
		checkFunc func(t *testing.T, result string)
	}{
		{
			name:  "struct with fields",
			value: TestStruct{Name: "John", Age: 30},
			checkFunc: func(t *testing.T, result string) {
				assert.Equal(t, "{Name:John,Age:30}", result)
			},
		},
		{
			name:  "empty struct",
			value: TestStruct{},
			checkFunc: func(t *testing.T, result string) {
				assert.Equal(t, "{Name:,Age:0}", result)
			},
		},
		{
			name:  "different struct values produce different results",
			value: TestStruct{Name: "John", Age: 30},
			checkFunc: func(t *testing.T, result string) {
				assert.Equal(t, "{Name:John,Age:30}", result)
				buffer2 := &bytes.Buffer{}
				handleReflection(buffer2, TestStruct{Name: "Jane", Age: 30})
				assert.NotEqual(t, result, buffer2.String())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buffer := &bytes.Buffer{}
			handleReflection(buffer, tt.value)
			tt.checkFunc(t, buffer.String())
		})
	}
}

// TestHash_HandleReflection_StructWithUnexportedFields tests structs with unexported fields.
func TestHash_HandleReflection_StructWithUnexportedFields(t *testing.T) {
	type TestStruct struct {
		Name     string
		age      int // unexported
		Exported string
	}
	type TestLastUnexportedStruct struct {
		Name     string
		age      int // unexported
		Exported string
		password string
	}

	tests := []struct {
		name      string
		value     interface{}
		checkFunc func(t *testing.T, result string)
	}{
		{
			name:  "struct with unexported fields",
			value: TestStruct{Name: "John", age: 30, Exported: "public"},
			checkFunc: func(t *testing.T, result string) {
				assert.Equal(t, "{Name:John,Exported:public}", result)
			},
		},
		{
			name:  "should produce consistent results",
			value: TestStruct{Name: "John", age: 30, Exported: "public"},
			checkFunc: func(t *testing.T, result string) {
				assert.Equal(t, "{Name:John,Exported:public}", result)
				buffer2 := &bytes.Buffer{}
				handleReflection(buffer2, TestStruct{Name: "Jane", age: 30, Exported: "public"})
				assert.NotEqual(t, result, buffer2.String())
			},
		},
		{
			name:  "last unexported field",
			value: TestLastUnexportedStruct{Name: "John", age: 30, Exported: "public", password: "password"},
			checkFunc: func(t *testing.T, result string) {
				assert.Equal(t, "{Name:John,Exported:public}", result)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buffer := &bytes.Buffer{}
			handleReflection(buffer, tt.value)
			tt.checkFunc(t, buffer.String())
		})
	}
}

// TestHash_HandleReflection_Pointer tests handleReflection with pointer types.
func TestHash_HandleReflection_Pointer(t *testing.T) {
	tests := []struct {
		name      string
		setupFunc func() interface{}
		checkFunc func(t *testing.T, result string)
	}{
		{
			name: "pointer to string",
			setupFunc: func() interface{} {
				value := "test"
				return &value
			},
			checkFunc: func(t *testing.T, result string) {
				assert.Equal(t, "test", result)
				// Should produce consistent results
				value := "test"
				buffer2 := &bytes.Buffer{}
				handleReflection(buffer2, &value)
				assert.Equal(t, result, buffer2.String())
			},
		},
		{
			name: "pointer to int",
			setupFunc: func() interface{} {
				value := 42
				return &value
			},
			checkFunc: func(t *testing.T, result string) {
				assert.Equal(t, "42", result)
				// Should produce consistent results
				value := 42
				buffer2 := &bytes.Buffer{}
				handleReflection(buffer2, &value)
				assert.Equal(t, result, buffer2.String())
			},
		},
		{
			name: "pointer to struct",
			setupFunc: func() interface{} {
				type S struct{ Name string }
				value := S{Name: "test"}
				return &value
			},
			checkFunc: func(t *testing.T, result string) {
				assert.Equal(t, "{Name:test}", result)
				// Should produce consistent results
				type S struct{ Name string }
				value := S{Name: "test"}
				buffer2 := &bytes.Buffer{}
				handleReflection(buffer2, &value)
				assert.Equal(t, result, buffer2.String())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value := tt.setupFunc()
			buffer := &bytes.Buffer{}
			handleReflection(buffer, value)
			tt.checkFunc(t, buffer.String())
		})
	}
}

// TestHash_HandleReflection_NilPointer tests nil pointer handling.
func TestHash_HandleReflection_NilPointer(t *testing.T) {
	tests := []struct {
		name      string
		value     interface{}
		checkFunc func(t *testing.T, result string)
	}{
		{
			name:  "nil pointer to string",
			value: (*string)(nil),
			checkFunc: func(t *testing.T, result string) {
				assert.Equal(t, "<nil>", result)
				// Should produce consistent results
				buffer2 := &bytes.Buffer{}
				handleReflection(buffer2, (*string)(nil))
				assert.Equal(t, result, buffer2.String())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buffer := &bytes.Buffer{}
			handleReflection(buffer, tt.value)
			tt.checkFunc(t, buffer.String())
		})
	}
}

// TestHash_HandleReflection_ReflectionSlice tests slice types handled via reflection.
func TestHash_HandleReflection_ReflectionSlice(t *testing.T) {
	tests := []struct {
		name      string
		value     interface{} // concrete slice type, not []interface{}
		checkFunc func(t *testing.T, result string)
	}{
		{
			name:  "string slice",
			value: []string{"a", "b", "c"},
			checkFunc: func(t *testing.T, result string) {
				assert.Equal(t, "[a,b,c]", result)
				// Should produce consistent results
				buffer2 := &bytes.Buffer{}
				handleReflection(buffer2, []string{"a", "b", "c"})
				assert.Equal(t, result, buffer2.String())
			},
		},
		{
			name:  "int slice",
			value: []int{1, 2, 3},
			checkFunc: func(t *testing.T, result string) {
				assert.Equal(t, "[1,2,3]", result)
				// Should produce consistent results
				buffer2 := &bytes.Buffer{}
				handleReflection(buffer2, []int{1, 2, 3})
				assert.Equal(t, result, buffer2.String())
			},
		},
		{
			name:  "empty slice",
			value: []string{},
			checkFunc: func(t *testing.T, result string) {
				assert.Equal(t, "[]", result)
			},
		},
		{
			name:  "slice with single element",
			value: []int{42},
			checkFunc: func(t *testing.T, result string) {
				assert.Equal(t, "[42]", result)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buffer := &bytes.Buffer{}
			handleReflection(buffer, tt.value)
			tt.checkFunc(t, buffer.String())
		})
	}
}

// TestHash_HandleReflection_ReflectionMap tests map types handled via reflection.
func TestHash_HandleReflection_ReflectionMap(t *testing.T) {
	tests := []struct {
		name      string
		value     interface{} // concrete map type, not map[string]interface{}
		checkFunc func(t *testing.T, result string)
	}{
		{
			name:  "int to string map",
			value: map[int]string{1: "a", 2: "b"},
			checkFunc: func(t *testing.T, result string) {
				// Keys should be sorted for consistency
				assert.Contains(t, result, "1:")
				assert.Contains(t, result, "2:")
				assert.Contains(t, result, "a")
				assert.Contains(t, result, "b")
				// Should produce consistent results regardless of map key order
				buffer2 := &bytes.Buffer{}
				handleReflection(buffer2, map[int]string{2: "b", 1: "a"})
				assert.Equal(t, result, buffer2.String())
			},
		},
		{
			name:  "empty map",
			value: map[int]string{},
			checkFunc: func(t *testing.T, result string) {
				assert.Equal(t, "{}", result)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buffer := &bytes.Buffer{}
			handleReflection(buffer, tt.value)
			tt.checkFunc(t, buffer.String())
		})
	}
}

// TestHash_HandleStruct tests handleStruct function directly.
func TestHash_HandleStruct(t *testing.T) {
	type TestStruct struct {
		Name string
		Age  int
	}

	tests := []struct {
		name      string
		value     interface{}
		checkFunc func(t *testing.T, result string)
	}{
		{
			name:  "struct with fields",
			value: TestStruct{Name: "John", Age: 30},
			checkFunc: func(t *testing.T, result string) {
				assert.Contains(t, result, "Name:")
				assert.Contains(t, result, "Age:")
				assert.Contains(t, result, "John")
				assert.Equal(t, "{Name:John,Age:30}", result)
			},
		},
		{
			name:  "empty struct",
			value: TestStruct{},
			checkFunc: func(t *testing.T, result string) {
				// Empty struct still has fields with zero values
				assert.Contains(t, result, "Name:")
				assert.Contains(t, result, "Age:")
				assert.Equal(t, "{Name:,Age:0}", result)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val := reflect.ValueOf(tt.value)
			buffer := &bytes.Buffer{}
			handleStruct(buffer, val)
			tt.checkFunc(t, buffer.String())
		})
	}
}

// TestHash_HandlePointer tests handlePointer function directly.
func TestHash_HandlePointer(t *testing.T) {
	tests := []struct {
		name      string
		setupFunc func() reflect.Value
		checkFunc func(t *testing.T, result string)
	}{
		{
			name: "pointer to string",
			setupFunc: func() reflect.Value {
				value := "test"
				return reflect.ValueOf(&value)
			},
			checkFunc: func(t *testing.T, result string) {
				assert.Equal(t, "test", result)
			},
		},
		{
			name: "nil pointer",
			setupFunc: func() reflect.Value {
				var value *string
				return reflect.ValueOf(value)
			},
			checkFunc: func(t *testing.T, result string) {
				assert.Equal(t, "<nil>", result)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val := tt.setupFunc()
			buffer := &bytes.Buffer{}
			handlePointer(buffer, val)
			tt.checkFunc(t, buffer.String())
		})
	}
}

// TestHash_HandleReflectionSlice tests handleReflectionSlice function directly.
func TestHash_HandleReflectionSlice(t *testing.T) {
	tests := []struct {
		name      string
		value     interface{}
		checkFunc func(t *testing.T, result string)
	}{
		{
			name:  "string slice",
			value: []string{"a", "b", "c"},
			checkFunc: func(t *testing.T, result string) {
				assert.Equal(t, "[a,b,c]", result)
			},
		},
		{
			name:  "int slice",
			value: []int{1, 2, 3},
			checkFunc: func(t *testing.T, result string) {
				assert.Equal(t, "[1,2,3]", result)
			},
		},
		{
			name:  "empty slice",
			value: []string{},
			checkFunc: func(t *testing.T, result string) {
				assert.Equal(t, "[]", result)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val := reflect.ValueOf(tt.value)
			buffer := &bytes.Buffer{}
			handleReflectionSlice(buffer, val)
			tt.checkFunc(t, buffer.String())
		})
	}
}

// TestHash_HandleReflectionMap tests handleReflectionMap function directly.
func TestHash_HandleReflectionMap(t *testing.T) {
	tests := []struct {
		name      string
		value     interface{}
		checkFunc func(t *testing.T, result string)
	}{
		{
			name:  "int to string map",
			value: map[int]string{1: "a", 2: "b"},
			checkFunc: func(t *testing.T, result string) {
				// Keys should be sorted
				assert.Equal(t, "{1:a,2:b}", result)
				// Should produce same result regardless of input order
				val2 := reflect.ValueOf(map[int]string{2: "b", 1: "a"})
				buffer2 := &bytes.Buffer{}
				handleReflectionMap(buffer2, val2)
				assert.Equal(t, result, buffer2.String())
			},
		},
		{
			name:  "empty map",
			value: map[int]string{},
			checkFunc: func(t *testing.T, result string) {
				assert.Equal(t, "{}", result)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val := reflect.ValueOf(tt.value)
			buffer := &bytes.Buffer{}
			handleReflectionMap(buffer, val)
			tt.checkFunc(t, buffer.String())
		})
	}
}

// TestHash_HandleReflection_UnsupportedTypes tests the default fallback case for unsupported types.
func TestHash_HandleReflection_UnsupportedTypes(t *testing.T) {
	tests := []struct {
		name      string
		value     interface{}
		checkFunc func(t *testing.T, result string)
	}{
		{
			name:  "channel type",
			value: make(chan int),
			checkFunc: func(t *testing.T, result string) {
				// The result should be a string like a memory address value
				assert.Regexp(t, `^0x[0-9a-fA-F]+$`, result)
			},
		},
		{
			name:  "function type",
			value: func() {},
			checkFunc: func(t *testing.T, result string) {
				// The result should be a string like a memory address value
				assert.Regexp(t, `^0x[0-9a-fA-F]+$`, result)
			},
		},
		{
			name:  "complex64",
			value: complex64(1 + 2i),
			checkFunc: func(t *testing.T, result string) {
				assert.Equal(t, "(1+2i)", result)
			},
		},
		{
			name:  "complex128",
			value: complex128(1 + 2i),
			checkFunc: func(t *testing.T, result string) {
				assert.Equal(t, "(1+2i)", result)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buffer := &bytes.Buffer{}
			handleReflection(buffer, tt.value)
			tt.checkFunc(t, buffer.String())
		})
	}
}

// TestHash_HandleReflectionMap_ComplexKeyTypes tests maps with various key types.
func TestHash_HandleReflectionMap_ComplexKeyTypes(t *testing.T) {
	tests := []struct {
		name      string
		value     interface{}
		checkFunc func(t *testing.T, result string)
	}{
		{
			name:  "map with float64 keys",
			value: map[float64]string{1.5: "a", 2.5: "b"},
			checkFunc: func(t *testing.T, result string) {
				assert.Equal(t, "{1.5:a,2.5:b}", result)
				// Should produce consistent results regardless of map key order
				buffer2 := &bytes.Buffer{}
				handleReflection(buffer2, map[float64]string{2.5: "b", 1.5: "a"})
				assert.Equal(t, result, buffer2.String())
			},
		},
		{
			name:  "map with bool keys",
			value: map[bool]string{true: "yes", false: "no"},
			checkFunc: func(t *testing.T, result string) {
				assert.Equal(t, "{false:no,true:yes}", result)
				// Keys should be sorted (false < true)
				buffer2 := &bytes.Buffer{}
				handleReflection(buffer2, map[bool]string{true: "yes", false: "no"})
				assert.Equal(t, result, buffer2.String())
			},
		},
		{
			name:  "map with string keys (non-interface)",
			value: map[string]int{"c": 5, "a": 1, "b": 2},
			checkFunc: func(t *testing.T, result string) {
				assert.Equal(t, "{a:1,b:2,c:5}", result)
				// Should produce consistent results
				buffer2 := &bytes.Buffer{}
				handleReflection(buffer2, map[string]int{"c": 5, "b": 2, "a": 1})
				assert.Equal(t, result, buffer2.String())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buffer := &bytes.Buffer{}
			handleReflection(buffer, tt.value)
			tt.checkFunc(t, buffer.String())
		})
	}
}

// TestHash_HandleReflectionMap_ComplexKeyTypesDirect tests handleReflectionMap directly with complex keys.
func TestHash_HandleReflectionMap_ComplexKeyTypesDirect(t *testing.T) {
	tests := []struct {
		name      string
		value     interface{}
		checkFunc func(t *testing.T, result string)
	}{
		{
			name:  "map with float64 keys",
			value: map[float64]string{1.5: "a", 2.5: "b"},
			checkFunc: func(t *testing.T, result string) {
				// Keys should be sorted
				assert.Equal(t, "{1.5:a,2.5:b}", result)
				// Should produce same result regardless of input order
				val2 := reflect.ValueOf(map[float64]string{2.5: "b", 1.5: "a"})
				buffer2 := &bytes.Buffer{}
				handleReflectionMap(buffer2, val2)
				assert.Equal(t, result, buffer2.String())
			},
		},
		{
			name:  "map with bool keys",
			value: map[bool]string{true: "yes", false: "no"},
			checkFunc: func(t *testing.T, result string) {
				// Keys should be sorted (false < true)
				assert.Equal(t, "{false:no,true:yes}", result)
				// Should produce same result regardless of input order
				val2 := reflect.ValueOf(map[bool]string{true: "yes", false: "no"})
				buffer2 := &bytes.Buffer{}
				handleReflectionMap(buffer2, val2)
				assert.Equal(t, result, buffer2.String())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val := reflect.ValueOf(tt.value)
			buffer := &bytes.Buffer{}
			handleReflectionMap(buffer, val)
			tt.checkFunc(t, buffer.String())
		})
	}
}

// TestHash_DeeplyNestedStructures tests deeply nested combinations.
func TestHash_DeeplyNestedStructures(t *testing.T) {
	type InnerStruct struct {
		Value int
	}

	type MiddleStruct struct {
		Inner InnerStruct
		Slice []InnerStruct
	}

	type OuterStruct struct {
		Middle MiddleStruct
		Map    map[string]MiddleStruct
	}

	tests := []struct {
		name      string
		value     interface{}
		checkFunc func(t *testing.T, result string)
	}{
		{
			name: "deeply nested struct",
			value: OuterStruct{
				Middle: MiddleStruct{
					Inner: InnerStruct{Value: 42},
					Slice: []InnerStruct{{Value: 1}, {Value: 2}},
				},
				Map: map[string]MiddleStruct{
					"key": {
						Inner: InnerStruct{Value: 100},
						Slice: []InnerStruct{{Value: 200}},
					},
				},
			},
			checkFunc: func(t *testing.T, result string) {
				assert.Equal(t, "{Middle:{Inner:{Value:42},Slice:[{Value:1},{Value:2}]},Map:{key:{Inner:{Value:100},Slice:[{Value:200}]}}}", result)
				// Should produce consistent results
				buffer2 := &bytes.Buffer{}
				makeHashable(buffer2, OuterStruct{
					Middle: MiddleStruct{
						Inner: InnerStruct{Value: 42},
						Slice: []InnerStruct{{Value: 1}, {Value: 2}},
					},
					Map: map[string]MiddleStruct{
						"key": {
							Inner: InnerStruct{Value: 100},
							Slice: []InnerStruct{{Value: 200}},
						},
					},
				})
				assert.Equal(t, result, buffer2.String())
			},
		},
		{
			name: "nested maps and slices",
			value: map[string]interface{}{
				"level1": map[string]interface{}{
					"level2": []interface{}{
						map[string]interface{}{
							"level3": "value",
						},
					},
				},
			},
			checkFunc: func(t *testing.T, result string) {
				assert.Equal(t, "{level1:{level2:[{level3:value}]}}", result)
				// Should produce consistent results
				buffer2 := &bytes.Buffer{}
				makeHashable(buffer2, map[string]interface{}{
					"level1": map[string]interface{}{
						"level2": []interface{}{
							map[string]interface{}{
								"level3": "value",
							},
						},
					},
				})
				assert.Equal(t, result, buffer2.String())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buffer := &bytes.Buffer{}
			makeHashable(buffer, tt.value)
			tt.checkFunc(t, buffer.String())
		})
	}
}
