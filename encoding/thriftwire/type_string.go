// Code generated by "stringer -type Type"; DO NOT EDIT.

package thriftwire

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[Stop-0]
	_ = x[Bool-2]
	_ = x[Byte-3]
	_ = x[Double-4]
	_ = x[I16-6]
	_ = x[I32-8]
	_ = x[I64-10]
	_ = x[String-11]
	_ = x[Struct-12]
	_ = x[Map-13]
	_ = x[Set-14]
	_ = x[List-15]
	_ = x[UUID-16]
}

const (
	_Type_name_0 = "Stop"
	_Type_name_1 = "BoolByteDouble"
	_Type_name_2 = "I16"
	_Type_name_3 = "I32"
	_Type_name_4 = "I64StringStructMapSetListUUID"
)

var (
	_Type_index_1 = [...]uint8{0, 4, 8, 14}
	_Type_index_4 = [...]uint8{0, 3, 9, 15, 18, 21, 25, 29}
)

func (i Type) String() string {
	switch {
	case i == 0:
		return _Type_name_0
	case 2 <= i && i <= 4:
		i -= 2
		return _Type_name_1[_Type_index_1[i]:_Type_index_1[i+1]]
	case i == 6:
		return _Type_name_2
	case i == 8:
		return _Type_name_3
	case 10 <= i && i <= 16:
		i -= 10
		return _Type_name_4[_Type_index_4[i]:_Type_index_4[i+1]]
	default:
		return "Type(" + strconv.FormatInt(int64(i), 10) + ")"
	}
}
