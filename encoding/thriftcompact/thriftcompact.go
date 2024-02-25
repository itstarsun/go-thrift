// Package thriftcompact implements the Thrift Compact protocol encoding.
// See https://github.com/apache/thrift/blob/master/doc/specs/thrift-compact-protocol.md
// for details.
package thriftcompact

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"math"

	"github.com/itstarsun/go-thrift/encoding/thriftwire"
)

const (
	protocolID = 0x082
	version1   = 1
)

// Protocol is the [thriftwire.Protocol] that implements the Thrift Compact protocol encoding.
var Protocol protocol

type protocol struct{}

func (protocol) NewReader(r io.Reader) thriftwire.Reader {
	return &reader{Reader: bufio.NewReader(r)}
}

func (protocol) NewWriter(w io.Writer) thriftwire.Writer {
	return &writer{Writer: bufio.NewWriter(w), uw: w}
}

type reader struct {
	*bufio.Reader
	lastFieldIDs []int16
	lastFieldID  int16
	boolField    boolField
}

type boolField byte

const (
	boolNone boolField = iota
	boolTrue
	boolFalse
)

func (f *boolField) set(t compactType) {
	switch t {
	case _BOOLEAN_TRUE:
		*f = boolTrue
	case _BOOLEAN_FALSE:
		*f = boolFalse
	}
}

func (x *reader) ReadMessageBegin() (h thriftwire.MessageHeader, err error) {
	p, err := x.ReadByte()
	if err != nil {
		return h, err
	}
	if p != protocolID {
		return h, fmt.Errorf("thriftcompact: bad protocol ID %x", p)
	}
	vt, err := x.ReadByte()
	if err != nil {
		return h, err
	}
	version := vt & 0x1f
	if version != version1 {
		return h, fmt.Errorf("thriftcompact: bad version %d", version)
	}
	h.Type = thriftwire.MessageType(vt >> 5)
	id, err := x.readUvarint()
	h.ID = int32(id)
	if err != nil {
		return h, err
	}
	if h.Name, err = x.ReadString(); err != nil {
		return h, err
	}
	return h, nil
}

func (x *reader) ReadMessageEnd() error {
	return nil
}

func (x *reader) ReadStructBegin() (h thriftwire.StructHeader, err error) {
	x.lastFieldIDs = append(x.lastFieldIDs, x.lastFieldID)
	x.lastFieldID = 0
	return h, nil
}

func (x *reader) ReadStructEnd() error {
	if len(x.lastFieldIDs) == 0 {
		return fmt.Errorf("thriftcompact: ReadStructEnd called without matching ReadStructBegin")
	}
	x.lastFieldID = x.lastFieldIDs[len(x.lastFieldIDs)-1]
	x.lastFieldIDs = x.lastFieldIDs[:len(x.lastFieldIDs)-1]
	return nil
}

func (x *reader) ReadFieldBegin() (h thriftwire.FieldHeader, err error) {
	dt, err := x.ReadByte()
	if err != nil {
		return h, err
	}
	t := compactType(dt & 15)
	h.Type, err = t.wire()
	if err != nil || h.Type == thriftwire.Stop {
		return h, err
	}
	modifier := int16(dt >> 4)
	if modifier == 0 {
		h.ID, err = x.ReadI16()
		if err != nil {
			return h, err
		}
	} else {
		h.ID = x.lastFieldID + modifier
	}
	x.lastFieldID = h.ID
	x.boolField.set(t)
	return h, nil
}

func (x *reader) ReadFieldEnd() error {
	return nil
}

func readSize(b *bufio.Reader) (int, error) {
	v, err := binary.ReadUvarint(b)
	return int(v), err
}

// ReadMapHeader reads a map header from b.
func ReadMapHeader(b *bufio.Reader) (h thriftwire.MapHeader, err error) {
	h.Size, err = readSize(b)
	if err != nil || h.Size == 0 {
		return h, err
	}
	kv, err := b.ReadByte()
	if err != nil {
		return h, err
	}
	h.Key, err = compactType(kv >> 4).wire()
	if err != nil {
		return h, err
	}
	h.Value, err = compactType(kv & 15).wire()
	if err != nil {
		return h, err
	}
	return h, nil
}

func (x *reader) ReadMapBegin() (thriftwire.MapHeader, error) {
	return ReadMapHeader(x.Reader)
}

func (x *reader) ReadMapEnd() error {
	return nil
}

// ReadSetHeader reads a set header from b.
func ReadSetHeader(b *bufio.Reader) (h thriftwire.SetHeader, err error) {
	st, err := b.ReadByte()
	if err != nil {
		return h, err
	}
	h.Element, err = compactType(st & 15).wire()
	if err != nil {
		return h, err
	}
	h.Size = int(st >> 4)
	if h.Size == 15 {
		h.Size, err = readSize(b)
		if err != nil {
			return h, err
		}
	}
	return h, nil
}

func (x *reader) ReadSetBegin() (thriftwire.SetHeader, error) {
	return ReadSetHeader(x.Reader)
}

func (x *reader) ReadSetEnd() error {
	return nil
}

// ReadListHeader reads a list header from b.
func ReadListHeader(b *bufio.Reader) (h thriftwire.ListHeader, err error) {
	sh, err := ReadSetHeader(b)
	return thriftwire.ListHeader(sh), err
}

func (x *reader) ReadListBegin() (thriftwire.ListHeader, error) {
	return ReadListHeader(x.Reader)
}

func (x *reader) ReadListEnd() error {
	return nil
}

func (x *reader) ReadBool() (bool, error) {
	if x.boolField != boolNone {
		v := x.boolField == boolTrue
		x.boolField = boolNone
		return v, nil
	}
	v, err := x.ReadByte()
	return v != 0, err
}

func (x *reader) ReadDouble() (float64, error) {
	buf, err := thriftwire.Next(x.Reader, 8)
	if err != nil {
		return 0, err
	}
	return math.Float64frombits(binary.LittleEndian.Uint64(buf)), nil
}

func (x *reader) ReadI16() (int16, error) {
	v, err := x.ReadI64()
	return int16(v), err
}

func (x *reader) ReadI32() (int32, error) {
	v, err := x.ReadI64()
	return int32(v), err
}

func (x *reader) ReadI64() (int64, error) {
	return binary.ReadVarint(x.Reader)
}

func (x *reader) ReadString() (string, error) {
	n, err := x.readSize()
	if err != nil {
		return "", err
	}
	return thriftwire.ReadString(x.Reader, n)
}

func (x *reader) ReadBytes(buf []byte) ([]byte, error) {
	n, err := x.readSize()
	if err != nil {
		return buf, err
	}
	return thriftwire.ReadBytes(x.Reader, n, buf)
}

func (x *reader) ReadUUID(v *[16]byte) error {
	_, err := io.ReadFull(x.Reader, v[:])
	return err
}

func (x *reader) SkipString() error {
	n, err := x.readSize()
	if err != nil {
		return err
	}
	_, err = x.Discard(n)
	return err
}

func (x *reader) SkipUUID() error {
	_, err := x.Discard(16)
	return err
}

func (x *reader) readUvarint() (uint64, error) {
	return thriftwire.ReadUvarint(x.Reader)
}

func (x *reader) readSize() (int, error) {
	v, err := x.readUvarint()
	return int(v), err
}

func (x *reader) Reset(r io.Reader) {
	x.Reader.Reset(r)
	*x = reader{
		Reader:       bufio.NewReader(r),
		lastFieldIDs: x.lastFieldIDs[:0],
	}
}

type writer struct {
	*bufio.Writer
	uw           io.Writer
	buf          [binary.MaxVarintLen64]byte
	lastFieldIDs []int16
	lastFieldID  int16
	boolField    thriftwire.FieldHeader
}

func (x *writer) WriteMessageBegin(h thriftwire.MessageHeader) (err error) {
	err = x.WriteByte(protocolID)
	if err != nil {
		return err
	}
	err = x.WriteByte(version1 | byte(h.Type)<<5)
	if err != nil {
		return err
	}
	err = x.writeUvarint(uint64(h.ID))
	if err != nil {
		return err
	}
	return x.WriteString(h.Name)
}

func (x *writer) WriteMessageEnd() error {
	return nil
}

func (x *writer) WriteStructBegin(h thriftwire.StructHeader) error {
	x.lastFieldIDs = append(x.lastFieldIDs, x.lastFieldID)
	x.lastFieldID = 0
	return nil
}

func (x *writer) WriteStructEnd() error {
	if len(x.lastFieldIDs) == 0 {
		return fmt.Errorf("thriftcompact: WriteStructEnd called without matching WriteStructBegin")
	}
	x.lastFieldID = x.lastFieldIDs[len(x.lastFieldIDs)-1]
	x.lastFieldIDs = x.lastFieldIDs[:len(x.lastFieldIDs)-1]
	return x.WriteByte(byte(_STOP))
}

// MaxFieldHeaderSize is the maximum size of [thriftwire.FieldHeader] in bytes.
const MaxFieldHeaderSize = 1 + binary.MaxVarintLen16

type appendFieldHeaderOptions struct {
	h           thriftwire.FieldHeader
	isTrue      bool
	lastFieldID int16
}

func appendFieldHeader(buf []byte, opts appendFieldHeaderOptions) (_ []byte, err error) {
	var t compactType
	if opts.h.Type == thriftwire.Bool {
		if opts.isTrue {
			t = _BOOLEAN_TRUE
		} else {
			t = _BOOLEAN_FALSE
		}
	} else {
		t, err = typeCompact(opts.h.Type)
		if err != nil {
			return buf, err
		}
	}
	if delta := opts.h.ID - opts.lastFieldID; delta > 0 && delta <= 15 {
		return append(buf, byte(delta<<4)|byte(t)), nil
	}
	buf = append(buf, byte(t))
	buf = appendI16(buf, opts.h.ID)
	return buf, nil
}

// AppendFieldHeader appends h to buf and returns the extended buffer.
func AppendFieldHeader(buf []byte, h thriftwire.FieldHeader, isTrue bool, lastFieldID int16) ([]byte, error) {
	return appendFieldHeader(buf, appendFieldHeaderOptions{h, isTrue, lastFieldID})
}

func (x *writer) writeFieldHeader(h thriftwire.FieldHeader, isTrue bool) error {
	opts := appendFieldHeaderOptions{h, isTrue, x.lastFieldID}
	if err := appendWrite(x, MaxFieldHeaderSize, appendFieldHeader, opts); err != nil {
		return err
	}
	x.lastFieldID = h.ID
	return nil
}

func (x *writer) WriteFieldBegin(h thriftwire.FieldHeader) error {
	if h.Type == thriftwire.Bool {
		x.boolField = h
		return nil
	}
	return x.writeFieldHeader(h, false)
}

func (x *writer) WriteFieldEnd() error {
	return nil
}

// MaxMapHeaderSize is the maximum size of [thriftwire.MapHeader] in bytes.
const MaxMapHeaderSize = binary.MaxVarintLen64 + 1

// AppendMapHeader appends h to buf and returns the extended buffer.
func AppendMapHeader(buf []byte, h thriftwire.MapHeader) ([]byte, error) {
	if h.Size <= 0 {
		return append(buf, 0), nil
	}
	k, err := typeCompact(h.Key)
	if err != nil {
		return buf, err
	}
	v, err := typeCompact(h.Value)
	if err != nil {
		return buf, err
	}
	buf = appendSize(buf, h.Size)
	buf = append(buf, byte(k<<4)|byte(v))
	return buf, nil
}

func (x *writer) WriteMapBegin(h thriftwire.MapHeader) error {
	return appendWrite(x, MaxMapHeaderSize, AppendMapHeader, h)
}

func (x *writer) WriteMapEnd() error {
	return nil
}

// MaxSetHeaderSize is the maximum size of [thriftwire.SetHeader] in bytes.
const MaxSetHeaderSize = 1 + binary.MaxVarintLen64

// AppendSetHeader appends h to buf and returns the extended buffer.
func AppendSetHeader(buf []byte, h thriftwire.SetHeader) ([]byte, error) {
	e, err := typeCompact(h.Element)
	if err != nil {
		return buf, err
	}
	if h.Size >= 0 && h.Size < 15 {
		return append(buf, byte(h.Size<<4)|byte(e)), nil
	}
	buf = append(buf, 0xf0|byte(e))
	buf = appendSize(buf, h.Size)
	return buf, nil
}

// MaxListHeaderSize is the maximum size of [thriftwire.ListHeader] in bytes.
const MaxListHeaderSize = MaxSetHeaderSize

// AppendListHeader appends h to buf and returns the extended buffer.
func AppendListHeader(buf []byte, h thriftwire.ListHeader) ([]byte, error) {
	return AppendSetHeader(buf, thriftwire.SetHeader(h))
}

func (x *writer) WriteSetBegin(h thriftwire.SetHeader) error {
	return appendWrite(x, MaxSetHeaderSize, AppendSetHeader, h)
}

func (x *writer) WriteSetEnd() error {
	return nil
}

func (x *writer) WriteListBegin(h thriftwire.ListHeader) error {
	return x.WriteSetBegin(thriftwire.SetHeader(h))
}

func (x *writer) WriteListEnd() error {
	return nil
}

func (x *writer) WriteBool(v bool) error {
	if x.boolField.Type != thriftwire.Stop {
		err := x.writeFieldHeader(x.boolField, v)
		if err != nil {
			return err
		}
		x.boolField.Type = thriftwire.Stop
		return nil
	}
	if v {
		return x.WriteByte(1)
	}
	return x.WriteByte(0)
}

func (x *writer) WriteDouble(v float64) error {
	buf := x.buf[:8]
	binary.LittleEndian.PutUint64(buf, math.Float64bits(v))
	_, err := x.Write(buf)
	return err
}

func (x *writer) WriteI16(v int16) error {
	return x.WriteI64(int64(v))
}

func (x *writer) WriteI32(v int32) error {
	return x.WriteI64(int64(v))
}

func (x *writer) WriteI64(v int64) error {
	buf := x.buf[:]
	n := binary.PutVarint(buf, v)
	_, err := x.Write(buf[:n])
	return err
}

func (x *writer) WriteString(v string) error {
	if err := x.writeSize(len(v)); err != nil {
		return err
	}
	_, err := x.Writer.WriteString(v)
	return err
}

func (x *writer) WriteBytes(v []byte) error {
	if err := x.writeSize(len(v)); err != nil {
		return err
	}
	_, err := x.Write(v)
	return err
}

func (x *writer) WriteUUID(v *[16]byte) error {
	_, err := x.Write(v[:])
	return err
}

func (x *writer) Flush() error {
	if err := x.Writer.Flush(); err != nil {
		return err
	}
	return thriftwire.Flush(x.uw)
}

func (x *writer) Reset(uw io.Writer) {
	x.Writer.Reset(uw)
	*x = writer{
		Writer:       x.Writer,
		lastFieldIDs: x.lastFieldIDs[:0],
	}
}

func appendWrite[IN any](x *writer, max int, f func([]byte, IN) ([]byte, error), in IN) error {
	if err := x.ensureAvailable(max); err != nil {
		return err
	}
	buf := x.AvailableBuffer()
	buf, err := f(buf, in)
	if err != nil {
		return err
	}
	_, err = x.Write(buf)
	return err
}

func (x *writer) ensureAvailable(n int) error {
	if x.Available() >= n {
		return nil
	}
	return x.Writer.Flush()
}

func (x *writer) writeUvarint(v uint64) error {
	buf := x.buf[:]
	n := binary.PutUvarint(buf, v)
	_, err := x.Write(buf[:n])
	return err
}

func (x *writer) writeSize(v int) error {
	return x.writeUvarint(uint64(v))
}

func appendI16(buf []byte, v int16) []byte {
	return binary.AppendVarint(buf, int64(v))
}

func appendSize(buf []byte, v int) []byte {
	return binary.AppendUvarint(buf, uint64(v))
}
