package thriftwire

import (
	"fmt"
	"io"
)

// Protocol is the interface that defines methods for creating [Reader] and [Writer].
//
// NewReader returns a new [Reader] that reads from the given [io.Reader].
//
// NewWriter returns a new [Writer] that writes to the given [io.Writer].
type Protocol interface {
	NewReader(io.Reader) Reader
	NewWriter(io.Writer) Writer
}

// JoinProtocol returns a new [Protocol] that uses r to create a new [Reader]
// and uses w to create a new [Writer].
func JoinProtocol(r, w Protocol) Protocol {
	return joinedProtocol{r, w}
}

type joinedProtocol struct{ r, w Protocol }

func (p joinedProtocol) NewReader(r io.Reader) Reader { return p.r.NewReader(r) }
func (p joinedProtocol) NewWriter(w io.Writer) Writer { return p.w.NewWriter(w) }
func (p joinedProtocol) String() string {
	return fmt.Sprintf("thriftwire.JoinProtocol(%v, %v)", p.r, p.w)
}

// Reader is the interface that defines methods for reading Thrift values.
//
// ReadBytes reads the next string, appends it to the given buffer,
// and returns the extended buffer.
//
// Reset resets the Reader's state and allows it to be reused again
// as a new Reader, but instead reads from the given [io.Reader].
type Reader interface {
	ReadMessageBegin() (MessageHeader, error)
	ReadMessageEnd() error
	ReadStructBegin() (StructHeader, error)
	ReadStructEnd() error
	ReadFieldBegin() (FieldHeader, error)
	ReadFieldEnd() error
	ReadMapBegin() (MapHeader, error)
	ReadMapEnd() error
	ReadSetBegin() (SetHeader, error)
	ReadSetEnd() error
	ReadListBegin() (ListHeader, error)
	ReadListEnd() error
	ReadBool() (bool, error)
	ReadByte() (byte, error)
	ReadDouble() (float64, error)
	ReadI16() (int16, error)
	ReadI32() (int32, error)
	ReadI64() (int64, error)
	ReadString() (string, error)
	ReadBytes([]byte) ([]byte, error)
	ReadUUID(*[16]byte) error
	SkipString() error
	SkipUUID() error
	Reset(io.Reader)
}

// Writer is the interface that defines methods for writing Thrift values.
//
// Flush writes any buffered data to the underlying [io.Writer].
// Flush calls the Flush method on the underlying [io.Writer]
// if it implements the [Flusher] interface.
//
// Reset resets the Writer's state and allows it to be reused again
// as a new Writer, but instead writes to the given [io.Writer].
type Writer interface {
	WriteMessageBegin(MessageHeader) error
	WriteMessageEnd() error
	WriteStructBegin(StructHeader) error
	WriteStructEnd() error
	WriteFieldBegin(FieldHeader) error
	WriteFieldEnd() error
	WriteMapBegin(MapHeader) error
	WriteMapEnd() error
	WriteSetBegin(SetHeader) error
	WriteSetEnd() error
	WriteListBegin(ListHeader) error
	WriteListEnd() error
	WriteBool(bool) error
	WriteByte(byte) error
	WriteDouble(float64) error
	WriteI16(int16) error
	WriteI32(int32) error
	WriteI64(int64) error
	WriteString(string) error
	WriteBytes([]byte) error
	WriteUUID(*[16]byte) error
	Flush() error
	Reset(io.Writer)
}

// Flusher is the interface that wraps the Flush method.
//
// Flush writes any buffered data to the underlying [io.Writer].
type Flusher interface {
	Flush() error
}

// Flush is a convenience function that calls the Flush method on w
// if it implements the [Flusher] interface.
func Flush(w io.Writer) error {
	if f, ok := w.(Flusher); ok {
		return f.Flush()
	}
	return nil
}
