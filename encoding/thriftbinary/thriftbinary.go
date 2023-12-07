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

type protocolNonStrict struct{}

func (protocolNonStrict) NewReader(r io.Reader) thriftwire.Reader {
	return newReader(r, false)
}

func (protocolNonStrict) NewWriter(w io.Writer) thriftwire.Writer {
	return newWriter(w, false)
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

func (r *reader) ReadMessageBegin() (h thriftwire.MessageHeader, err error) {
	n, err := r.readSize()
	if err != nil {
		return h, err
	}
	if n < 0 {
		version := int64(n) & versionMask
		if version != version1 {
			return h, fmt.Errorf("thriftbinary: bad version %x", version)
		}
		h.Type = thriftwire.MessageType(n)
		h.Name, err = r.ReadString()
		if err != nil {
			return h, err
		}
		h.ID, err = r.ReadI32()
		return h, err
	} else if r.strict {
		return h, fmt.Errorf("thriftbinary: missing version")
	}
	h.Name, err = thriftwire.ReadString(r.Reader, n)
	if err != nil {
		return h, err
	}
	t, err := r.ReadByte()
	h.Type = thriftwire.MessageType(t)
	if err != nil {
		return h, err
	}
	h.ID, err = r.ReadI32()
	return h, err
}

func (r *reader) ReadMessageEnd() error {
	return nil
}

func (r *reader) ReadStructBegin() (h thriftwire.StructHeader, err error) {
	return h, nil
}

func (r *reader) ReadStructEnd() error {
	return nil
}

func (r *reader) ReadFieldBegin() (h thriftwire.FieldHeader, err error) {
	t, err := r.ReadByte()
	h.Type = thriftwire.Type(t)
	if err != nil || h.Type == thriftwire.Stop {
		return h, err
	}
	h.ID, err = r.ReadI16()
	return h, err
}

func (r *reader) ReadFieldEnd() error {
	return nil
}

func (r *reader) ReadMapBegin() (h thriftwire.MapHeader, err error) {
	k, err := r.ReadByte()
	h.Key = thriftwire.Type(k)
	if err != nil {
		return h, err
	}
	v, err := r.ReadByte()
	h.Value = thriftwire.Type(v)
	if err != nil {
		return h, err
	}
	h.Size, err = r.readSize()
	return h, err
}

func (r *reader) ReadMapEnd() error {
	return nil
}

func (r *reader) ReadSetBegin() (h thriftwire.SetHeader, err error) {
	e, err := r.ReadByte()
	h.Element = thriftwire.Type(e)
	if err != nil {
		return h, err
	}
	h.Size, err = r.readSize()
	return h, err
}

func (r *reader) ReadSetEnd() error {
	return nil
}

func (r *reader) ReadListBegin() (thriftwire.ListHeader, error) {
	sh, err := r.ReadSetBegin()
	return thriftwire.ListHeader(sh), err
}

func (r *reader) ReadListEnd() error {
	return nil
}

func (r *reader) ReadBool() (bool, error) {
	v, err := r.ReadByte()
	return v != 0, err
}

func (r *reader) ReadDouble() (float64, error) {
	buf, err := thriftwire.Next(r.Reader, 8)
	if err != nil {
		return 0, err
	}
	return math.Float64frombits(binary.BigEndian.Uint64(buf)), nil
}

func (r *reader) ReadI16() (int16, error) {
	buf, err := thriftwire.Next(r.Reader, 2)
	if err != nil {
		return 0, err
	}
	return int16(binary.BigEndian.Uint16(buf)), nil
}

func (r *reader) ReadI32() (int32, error) {
	buf, err := thriftwire.Next(r.Reader, 4)
	if err != nil {
		return 0, err
	}
	return int32(binary.BigEndian.Uint32(buf)), nil
}

func (r *reader) ReadI64() (int64, error) {
	buf, err := thriftwire.Next(r.Reader, 8)
	if err != nil {
		return 0, err
	}
	return int64(binary.BigEndian.Uint64(buf)), nil
}

func (r *reader) ReadString() (string, error) {
	n, err := r.readSize()
	if err != nil {
		return "", err
	}
	return thriftwire.ReadString(r.Reader, n)
}

func (r *reader) ReadBytes(buf []byte) ([]byte, error) {
	n, err := r.readSize()
	if err != nil {
		return buf, err
	}
	return thriftwire.ReadBytes(r.Reader, n, buf)
}

func (r *reader) ReadUUID(v *[16]byte) error {
	_, err := io.ReadFull(r.Reader, v[:])
	return err
}

func (r *reader) SkipString() error {
	n, err := r.readSize()
	if err != nil {
		return err
	}
	_, err = r.Discard(n)
	return err
}

func (r *reader) SkipUUID() error {
	_, err := r.Discard(16)
	return err
}

func (r *reader) readSize() (int, error) {
	v, err := r.ReadI32()
	return int(v), err
}

type writer struct {
	*bufio.Writer
	uw     io.Writer
	buf    [8]byte
	strict bool
}

func newWriter(w io.Writer, strict bool) thriftwire.Writer {
	return &writer{Writer: bufio.NewWriter(w), uw: w, strict: strict}
}

func (w *writer) WriteMessageBegin(h thriftwire.MessageHeader) error {
	if w.strict {
		if err := w.WriteI32(int32(uint32(h.Type) | version1)); err != nil {
			return err
		}
		if err := w.WriteString(h.Name); err != nil {
			return err
		}
		return w.WriteI32(h.ID)
	}
	if err := w.WriteString(h.Name); err != nil {
		return err
	}
	if err := w.WriteByte(byte(h.Type)); err != nil {
		return err
	}
	return w.WriteI32(h.ID)
}

func (w *writer) WriteMessageEnd() error {
	return nil
}

func (w *writer) WriteStructBegin(h thriftwire.StructHeader) error {
	return nil
}

func (w *writer) WriteStructEnd() error {
	return w.WriteByte(byte(thriftwire.Stop))
}

func (w *writer) WriteFieldBegin(h thriftwire.FieldHeader) error {
	if err := w.WriteByte(byte(h.Type)); err != nil {
		return err
	}
	return w.WriteI16(h.ID)
}

func (w *writer) WriteFieldEnd() error {
	return nil
}

func (w *writer) WriteMapBegin(h thriftwire.MapHeader) error {
	if err := w.WriteByte(byte(h.Key)); err != nil {
		return err
	}
	if err := w.WriteByte(byte(h.Value)); err != nil {
		return err
	}
	return w.WriteByte(byte(h.Size))
}

func (w *writer) WriteMapEnd() error {
	return nil
}

func (w *writer) WriteSetBegin(h thriftwire.SetHeader) error {
	if err := w.WriteByte(byte(h.Element)); err != nil {
		return err
	}
	return w.WriteByte(byte(h.Size))
}

func (w *writer) WriteSetEnd() error {
	return nil
}

func (w *writer) WriteListBegin(h thriftwire.ListHeader) error {
	return w.WriteSetBegin(thriftwire.SetHeader(h))
}

func (w *writer) WriteListEnd() error {
	return nil
}

func (w *writer) WriteBool(v bool) error {
	if v {
		return w.WriteByte(1)
	}
	return w.WriteByte(0)
}

func (w *writer) WriteDouble(v float64) error {
	buf := w.buf[:8]
	binary.BigEndian.PutUint64(buf, math.Float64bits(v))
	_, err := w.Write(buf)
	return err
}

func (w *writer) WriteI16(v int16) error {
	buf := w.buf[:2]
	binary.BigEndian.PutUint16(buf, uint16(v))
	_, err := w.Write(buf)
	return err
}

func (w *writer) WriteI32(v int32) error {
	buf := w.buf[:4]
	binary.BigEndian.PutUint32(buf, uint32(v))
	_, err := w.Write(buf)
	return err
}

func (w *writer) WriteI64(v int64) error {
	buf := w.buf[:8]
	binary.BigEndian.PutUint64(buf, uint64(v))
	_, err := w.Write(buf)
	return err
}

func (w *writer) WriteString(v string) error {
	if err := w.writeSize(len(v)); err != nil {
		return err
	}
	_, err := w.Writer.WriteString(v)
	return err
}

func (w *writer) WriteBytes(v []byte) error {
	if err := w.writeSize(len(v)); err != nil {
		return err
	}
	_, err := w.Write(v)
	return err
}

func (w *writer) WriteUUID(v *[16]byte) error {
	_, err := w.Write(v[:])
	return err
}

func (w *writer) writeSize(v int) error {
	return w.WriteI32(int32(v))
}

func (w *writer) Flush() error {
	if err := w.Writer.Flush(); err != nil {
		return err
	}
	return thriftwire.Flush(w.uw)
}

func (w *writer) Reset(uw io.Writer) {
	w.Writer.Reset(uw)
	w.uw = uw
}
