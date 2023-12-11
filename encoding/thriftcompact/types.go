package thriftcompact

import (
	"github.com/itstarsun/go-thrift/encoding/thriftwire"
)

//go:generate stringer -type compactType -trimprefix _
type compactType byte

const (
	_STOP          compactType = 0
	_BOOLEAN_TRUE  compactType = 1
	_BOOLEAN_FALSE compactType = 2
	_I8            compactType = 3
	_I16           compactType = 4
	_I32           compactType = 5
	_I64           compactType = 6
	_DOUBLE        compactType = 7
	_BINARY        compactType = 8
	_LIST          compactType = 9
	_SET           compactType = 10
	_MAP           compactType = 11
	_STRUCT        compactType = 12
	_UUID          compactType = 13
)

type invalidCompactTypeError compactType

func (e invalidCompactTypeError) Error() string {
	return "thriftcompact: invalid type: " + compactType(e).String()
}

func typeCompact(t thriftwire.Type) (compactType, error) {
	switch t {
	case thriftwire.Byte:
		return _I8, nil
	case thriftwire.I16:
		return _I16, nil
	case thriftwire.I32:
		return _I32, nil
	case thriftwire.I64:
		return _I64, nil
	case thriftwire.Double:
		return _DOUBLE, nil
	case thriftwire.String:
		return _BINARY, nil
	case thriftwire.List:
		return _LIST, nil
	case thriftwire.Set:
		return _SET, nil
	case thriftwire.Map:
		return _MAP, nil
	case thriftwire.Struct:
		return _STRUCT, nil
	case thriftwire.UUID:
		return _UUID, nil
	}
	return _STOP, thriftwire.InvalidTypeError(t)
}

func (t compactType) wire() (thriftwire.Type, error) {
	switch t {
	case _STOP:
		return thriftwire.Stop, nil
	case _BOOLEAN_TRUE, _BOOLEAN_FALSE:
		return thriftwire.Bool, nil
	case _I8:
		return thriftwire.Byte, nil
	case _I16:
		return thriftwire.I16, nil
	case _I32:
		return thriftwire.I32, nil
	case _I64:
		return thriftwire.I64, nil
	case _DOUBLE:
		return thriftwire.Double, nil
	case _BINARY:
		return thriftwire.String, nil
	case _LIST:
		return thriftwire.List, nil
	case _SET:
		return thriftwire.Set, nil
	case _MAP:
		return thriftwire.Map, nil
	case _STRUCT:
		return thriftwire.Struct, nil
	case _UUID:
		return thriftwire.UUID, nil
	}
	return thriftwire.Stop, invalidCompactTypeError(t)
}
