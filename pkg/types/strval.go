package types

import (
	"strconv"
	"strings"
)

// StrValue 可转换为其他类型的字符串
type StrValue string

// IsEmpty 是否为空
func (v StrValue) IsEmpty() bool {
	return v == ""
}

// Default 设置默认值
func (v StrValue) Default(value string) StrValue {
	if v == "" {
		return StrValue(value)
	}
	return v
}

// String 转为 string 类型
func (v StrValue) String() string {
	return string(v)
}

// Int 转为 int 类型
func (v StrValue) Int() (int, error) {
	return strconv.Atoi(string(v))
}

// Uint 转为 uint 类型
func (v StrValue) Uint() (uint, error) {
	num, err := strconv.ParseUint(string(v), 10, 0)

	return uint(num), err
}

// Int8 转为 int8 类型
func (v StrValue) Int8() (int8, error) {
	num, err := strconv.ParseInt(string(v), 10, 8)

	return int8(num), err
}

// Uint8 转为 int8 类型
func (v StrValue) Uint8() (uint8, error) {
	num, err := strconv.ParseUint(string(v), 10, 8)

	return uint8(num), err
}

// Int16 转为 int16 类型
func (v StrValue) Int16() (int16, error) {
	num, err := strconv.ParseInt(string(v), 10, 16)

	return int16(num), err
}

// Uint16 转为 uint16 类型
func (v StrValue) Uint16() (uint16, error) {
	num, err := strconv.ParseUint(string(v), 10, 16)

	return uint16(num), err
}

// Int32 转为 int32 类型
func (v StrValue) Int32() (int32, error) {
	num, err := strconv.ParseInt(string(v), 10, 32)

	return int32(num), err
}

// Uint32 转为 uint32 类型
func (v StrValue) Uint32() (uint32, error) {
	num, err := strconv.ParseUint(string(v), 10, 32)

	return uint32(num), err
}

// Int64 转为 int64 类型
func (v StrValue) Int64() (int64, error) {
	return strconv.ParseInt(string(v), 10, 64)
}

// Uint64 转为 int64 类型
func (v StrValue) Uint64() (uint64, error) {
	return strconv.ParseUint(string(v), 10, 64)
}

// Bool 转为 bool 类型
func (v StrValue) Bool() (bool, error) {
	return strconv.ParseBool(string(v))
}

// Float32 转为 float32 类型
func (v StrValue) Float32() (float32, error) {
	num, err := strconv.ParseFloat(string(v), 32)

	return float32(num), err
}

// Float64 转为 float64 类型
func (v StrValue) Float64() (float64, error) {
	return strconv.ParseFloat(string(v), 64)
}

// StringArray 转为 string 数组类型
func (v StrValue) StringArray() []string {
	arr := strings.Split(string(v), ",")
	resultArr := make([]string, 0, len(arr))
	for _, item := range arr {
		if strings.TrimSpace(item) != "" {
			resultArr = append(resultArr, item)
		}
	}
	return resultArr
}

// IntArray 转为 int 数组类型
func (v StrValue) IntArray() []int {
	arr := strings.Split(string(v), ",")
	resultArr := make([]int, 0, len(arr))
	for _, item := range arr {
		if strings.TrimSpace(item) != "" {
			val, _ := strconv.Atoi(item)
			resultArr = append(resultArr, val)
		}
	}
	return resultArr
}

// UintArray 转为 uint 数组类型
func (v StrValue) UintArray() []uint {
	arr := strings.Split(string(v), ",")
	resultArr := make([]uint, 0, len(arr))
	for _, item := range arr {
		if strings.TrimSpace(item) != "" {
			val, _ := strconv.ParseUint(item, 10, 0)
			resultArr = append(resultArr, uint(val))
		}
	}
	return resultArr
}

// Int8Array 转为 int8 数组类型
func (v StrValue) Int8Array() []int8 {
	arr := strings.Split(string(v), ",")
	resultArr := make([]int8, 0, len(arr))
	for _, item := range arr {
		if strings.TrimSpace(item) != "" {
			val, _ := strconv.ParseInt(item, 10, 8)
			resultArr = append(resultArr, int8(val))
		}
	}
	return resultArr
}

// Uint8Array 转为 uint8 数组类型
func (v StrValue) Uint8Array() []uint8 {
	arr := strings.Split(string(v), ",")
	resultArr := make([]uint8, 0, len(arr))
	for _, item := range arr {
		if strings.TrimSpace(item) != "" {
			val, _ := strconv.ParseUint(item, 10, 8)
			resultArr = append(resultArr, uint8(val))
		}
	}
	return resultArr
}

// Int16Array 转为 int16 数组类型
func (v StrValue) Int16Array() []int16 {
	arr := strings.Split(string(v), ",")
	resultArr := make([]int16, 0, len(arr))
	for _, item := range arr {
		if strings.TrimSpace(item) != "" {
			val, _ := strconv.ParseInt(item, 10, 16)
			resultArr = append(resultArr, int16(val))
		}
	}
	return resultArr
}

// Uint16Array 转为 uint16 数组类型
func (v StrValue) Uint16Array() []uint16 {
	arr := strings.Split(string(v), ",")
	resultArr := make([]uint16, 0, len(arr))
	for _, item := range arr {
		if strings.TrimSpace(item) != "" {
			val, _ := strconv.ParseUint(item, 10, 16)
			resultArr = append(resultArr, uint16(val))
		}
	}
	return resultArr
}

// Int32Array 转为 int32 数组类型
func (v StrValue) Int32Array() []int32 {
	arr := strings.Split(string(v), ",")
	resultArr := make([]int32, 0, len(arr))
	for _, item := range arr {
		if strings.TrimSpace(item) != "" {
			val, _ := strconv.ParseInt(item, 10, 32)
			resultArr = append(resultArr, int32(val))
		}
	}
	return resultArr
}

// Uint32Array 转为 uint32 数组类型
func (v StrValue) Uint32Array() []uint32 {
	arr := strings.Split(string(v), ",")
	resultArr := make([]uint32, 0, len(arr))
	for _, item := range arr {
		if strings.TrimSpace(item) != "" {
			val, _ := strconv.ParseUint(item, 10, 32)
			resultArr = append(resultArr, uint32(val))
		}
	}
	return resultArr
}

// Int64Array 转为 int64 数组类型
func (v StrValue) Int64Array() []int64 {
	arr := strings.Split(string(v), ",")
	resultArr := make([]int64, 0, len(arr))
	for _, item := range arr {
		if strings.TrimSpace(item) != "" {
			val, _ := strconv.ParseInt(item, 10, 64)
			resultArr = append(resultArr, val)
		}
	}
	return resultArr
}

// Uint64Array 转为 uint64 数组类型
func (v StrValue) Uint64Array() []uint64 {
	arr := strings.Split(string(v), ",")
	resultArr := make([]uint64, 0, len(arr))
	for _, item := range arr {
		if strings.TrimSpace(item) != "" {
			val, _ := strconv.ParseUint(item, 10, 64)
			resultArr = append(resultArr, uint64(val))
		}
	}
	return resultArr
}

// BoolArray 转为 bool 数组类型
func (v StrValue) BoolArray() []bool {
	arr := strings.Split(string(v), ",")
	resultArr := make([]bool, 0, len(arr))
	for _, item := range arr {
		if strings.TrimSpace(item) != "" {
			val, _ := strconv.ParseBool(item)
			resultArr = append(resultArr, val)
		}
	}
	return resultArr
}

// Float32Array 转为 float32 数组类型
func (v StrValue) Float32Array() []float32 {
	arr := strings.Split(string(v), ",")
	resultArr := make([]float32, 0, len(arr))
	for _, item := range arr {
		if strings.TrimSpace(item) != "" {
			val, _ := strconv.ParseFloat(item, 32)
			resultArr = append(resultArr, float32(val))
		}
	}
	return resultArr
}

// Float64Array 转为 float64 数组类型
func (v StrValue) Float64Array() []float64 {
	arr := strings.Split(string(v), ",")
	resultArr := make([]float64, 0, len(arr))
	for _, item := range arr {
		if strings.TrimSpace(item) != "" {
			val, _ := strconv.ParseFloat(item, 64)
			resultArr = append(resultArr, float64(val))
		}
	}
	return resultArr
}
