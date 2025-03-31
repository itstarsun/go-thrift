// Package thriftbinary implements the Thrift Binary protocol encoding.
// See https://github.com/apache/thrift/blob/master/doc/specs/thrift-binary-protocol.md
// for details.
package thriftbinary

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"math"

	"github.com/itstarsun/go-thrift/encoding/thriftwire"
)

const (
	versionMask = 0xffff0000
	version1    = 0x80010000
)

// Protocol is the [thriftwire.Protocol] that implements the Thrift Binary protocol encoding.
var Protocol protocol

// ProtocolNonStrict is the [thriftwire.Protocol] that implements the Thrift Binary protocol encoding
// using the older encoding (aka non-strict).
var ProtocolNonStrict protocolNonStrict

type protocol struct{}

func (protocol) NewReader(r io.Reader) thriftwire.Reader {
	return newReader(r, true)
}

func (protocol) NewWriter(w io.Writer) thriftwire.Writer {
	return newWriter(w, true)
}

func (protocol) String() string {
	return "thriftbinary.Protocol"
}

type protocolNonStrict struct{}

func (protocolNonStrict) NewReader(r io.Reader) thriftwire.Reader {
	return newReader(r, false)
}

func (protocolNonStrict) NewWriter(w io.Writer) thriftwire.Writer {
	return newWriter(w, false)
}

func (protocolNonStrict) String() string {
	return "thriftbinary.ProtocolNonStrict"
}

var (
	_ thriftwire.Protocol = (*protocol)(nil)
	_ thriftwire.Protocol = (*protocolNonStrict)(nil)
)

type reader struct {
	*bufio.Reader
	strict bool
}

func newReader(r io.Reader, strict bool) *reader {
	return &reader{Reader: bufio.NewReader(r), strict: strict}
}

func (x *reader) ReadMessageBegin() (h thriftwire.MessageHeader, err error) {
	n, err := x.readSize()
	if err != nil {
		return h, err
	}
	if n < 0 {
		version := int64(n) & versionMask
		if version != version1 {
			return h, fmt.Errorf("thriftbinary: bad version %x", version)
		}
		h.Type = thriftwire.MessageType(n)
		h.Name, err = x.ReadString()
		if err != nil {
			return h, err
		}
		h.ID, err = x.ReadI32()
		return h, err
	} else if x.strict {
		return h, fmt.Errorf("thriftbinary: missing version")
	}
	h.Name, err = thriftwire.ReadString(x.Reader, n)
	if err != nil {
		return h, err
	}
	h.Type, err = x.readMessageType()
	if err != nil {
		return h, err
	}
	h.ID, err = x.ReadI32()
	return h, err
}

func (x *reader) ReadMessageEnd() error {
	return nil
}

func (x *reader) ReadStructBegin() (h thriftwire.StructHeader, err error) {
	return h, nil
}

func (x *reader) ReadStructEnd() error {
	return nil
}

func (x *reader) ReadFieldBegin() (h thriftwire.FieldHeader, err error) {
	h.Type, err = x.readType()
	if err != nil || h.Type == thriftwire.Stop {
		return h, err
	}
	h.ID, err = x.ReadI16()
	return h, err
}

func (x *reader) ReadFieldEnd() error {
	return nil
}

func (x *reader) ReadMapBegin() (h thriftwire.MapHeader, err error) {
	h.Key, err = x.readType()
	if err != nil {
		return h, err
	}
	h.Value, err = x.readType()
	if err != nil {
		return h, err
	}
	h.Size, err = x.readSize()
	return h, err
}

func (x *reader) ReadMapEnd() error {
	return nil
}

func (x *reader) ReadSetBegin() (h thriftwire.SetHeader, err error) {
	h.Element, err = x.readType()
	if err != nil {
		return h, err
	}
	h.Size, err = x.readSize()
	return h, err
}

func (x *reader) ReadSetEnd() error {
	return nil
}

func (x *reader) ReadListBegin() (thriftwire.ListHeader, error) {
	sh, err := x.ReadSetBegin()
	return thriftwire.ListHeader(sh), err
}

func (x *reader) ReadListEnd() error {
	return nil
}

func (x *reader) ReadBool() (bool, error) {
	v, err := x.ReadByte()
	return v != 0, err
}

func (x *reader) ReadDouble() (float64, error) {
	buf, err := thriftwire.Next(x.Reader, 8)
	if err != nil {
		return 0, err
	}
	return math.Float64frombits(binary.BigEndian.Uint64(buf)), nil
}

func (x *reader) ReadI16() (int16, error) {
	buf, err := thriftwire.Next(x.Reader, 2)
	if err != nil {
		return 0, err
	}
	return int16(binary.BigEndian.Uint16(buf)), nil
}

func (x *reader) ReadI32() (int32, error) {
	buf, err := thriftwire.Next(x.Reader, 4)
	if err != nil {
		return 0, err
	}
	return int32(binary.BigEndian.Uint32(buf)), nil
}

func (x *reader) ReadI64() (int64, error) {
	buf, err := thriftwire.Next(x.Reader, 8)
	if err != nil {
		return 0, err
	}
	return int64(binary.BigEndian.Uint64(buf)), nil
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

func (x *reader) readMessageType() (thriftwire.MessageType, error) {
	v, err := x.ReadByte()
	return thriftwire.MessageType(v), err
}

func (x *reader) readType() (thriftwire.Type, error) {
	v, err := x.ReadByte()
	return thriftwire.Type(v), err
}

func (x *reader) readSize() (int, error) {
	v, err := x.ReadI32()
	return int(v), err
}

type writer struct {
	*bufio.Writer
	w      io.Writer
	buf    [8]byte
	strict bool
}

func newWriter(w io.Writer, strict bool) thriftwire.Writer {
	return &writer{Writer: bufio.NewWriter(w), w: w, strict: strict}
}

func (x *writer) WriteMessageBegin(h thriftwire.MessageHeader) error {
	if x.strict {
		if err := x.WriteI32(int32(uint32(h.Type) | version1)); err != nil {
			return err
		}
		if err := x.WriteString(h.Name); err != nil {
			return err
		}
		return x.WriteI32(h.ID)
	}
	if err := x.WriteString(h.Name); err != nil {
		return err
	}
	if err := x.writeMessageType(h.Type); err != nil {
		return err
	}
	return x.WriteI32(h.ID)
}

func (x *writer) WriteMessageEnd() error {
	return nil
}

func (x *writer) WriteStructBegin(h thriftwire.StructHeader) error {
	return nil
}

func (x *writer) WriteStructEnd() error {
	return x.writeType(thriftwire.Stop)
}

func (x *writer) WriteFieldBegin(h thriftwire.FieldHeader) error {
	if err := x.writeType(h.Type); err != nil {
		return err
	}
	return x.WriteI16(h.ID)
}

func (x *writer) WriteFieldEnd() error {
	return nil
}

func (x *writer) WriteMapBegin(h thriftwire.MapHeader) error {
	if err := x.writeType(h.Key); err != nil {
		return err
	}
	if err := x.writeType(h.Value); err != nil {
		return err
	}
	return x.writeSize(h.Size)
}

func (x *writer) WriteMapEnd() error {
	return nil
}

func (x *writer) WriteSetBegin(h thriftwire.SetHeader) error {
	if err := x.writeType(h.Element); err != nil {
		return err
	}
	return x.writeSize(h.Size)
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
	if v {
		return x.WriteByte(1)
	}
	return x.WriteByte(0)
}

func (x *writer) WriteDouble(v float64) error {
	buf := x.buf[:8]
	binary.BigEndian.PutUint64(buf, math.Float64bits(v))
	_, err := x.Write(buf)
	return err
}

func (x *writer) WriteI16(v int16) error {
	buf := x.buf[:2]
	binary.BigEndian.PutUint16(buf, uint16(v))
	_, err := x.Write(buf)
	return err
}

func (x *writer) WriteI32(v int32) error {
	buf := x.buf[:4]
	binary.BigEndian.PutUint32(buf, uint32(v))
	_, err := x.Write(buf)
	return err
}

func (x *writer) WriteI64(v int64) error {
	buf := x.buf[:8]
	binary.BigEndian.PutUint64(buf, uint64(v))
	_, err := x.Write(buf)
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
	return thriftwire.Flush(x.w)
}

func (x *writer) Reset(w io.Writer) {
	x.Writer.Reset(w)
	x.w = w
}

func (x *writer) writeMessageType(t thriftwire.MessageType) error {
	return x.WriteByte(byte(t))
}

func (x *writer) writeType(t thriftwire.Type) error {
	return x.WriteByte(byte(t))
}

func (x *writer) writeSize(v int) error {
	return x.WriteI32(int32(v))
}
