package thriftwire

// MessageType is the type of a Thrift message.
//
//go:generate stringer -type MessageType
type MessageType byte

const (
	Call      MessageType = 1
	Reply     MessageType = 2
	Exception MessageType = 3
	OneWay    MessageType = 4
)

// InvalidMessageTypeError values describe errors resulting from an invalid [MessageType].
type InvalidMessageTypeError MessageType

func (e InvalidMessageTypeError) Error() string {
	return "thriftwire: invalid message type: " + MessageType(e).String()
}

// Type is the type of a Thrift value.
//
//go:generate stringer -type Type
type Type byte

const (
	Stop   Type = 0
	Bool   Type = 2
	Byte   Type = 3
	Double Type = 4
	I16    Type = 6
	I32    Type = 8
	I64    Type = 10
	String Type = 11
	Struct Type = 12
	Map    Type = 13
	Set    Type = 14
	List   Type = 15
	UUID   Type = 16
)

// InvalidTypeError values describe errors resulting from an invalid [Type].
type InvalidTypeError Type

func (e InvalidTypeError) Error() string {
	return "thriftwire: invalid type: " + Type(e).String()
}

// A MessageHeader represents the header of a Thrift message.
type MessageHeader struct {
	Name string // may be empty
	Type MessageType
	ID   int32
}

// A StructHeader represents the header of a Thrift struct.
type StructHeader struct {
	Name string // may be empty
}

// A FieldHeader represents the header of a Thrift field.
type FieldHeader struct {
	Name string // may be empty
	Type Type
	ID   int16
}

// A MapHeader represents the header of a Thrift map.
type MapHeader struct {
	Key, Value Type // may be Stop if Size <= 0
	Size       int
}

// A SetHeader represents the header of a Thrift set.
type SetHeader struct {
	Element Type // may be Stop if Size <= 0
	Size    int
}

// A ListHeader represents the header of a Thrift list.
type ListHeader struct {
	Element Type // may be Stop if Size <= 0
	Size    int
}
