package thrift

import (
	"archive/tar"
	"bytes"
	"errors"
	"io"
	"reflect"
	"strings"
	"testing"

	"github.com/itstarsun/go-thrift/encoding/thriftwire"
)

func TestWireError(t *testing.T) {
	tests := []struct {
		err  error
		want string
	}{{
		err:  &wireError{action: "WriteString", err: errors.New("some underlying error")},
		want: "thrift: WriteString: some underlying error",
	}, {
		err:  &wireError{action: "ReadString", err: io.EOF},
		want: "thrift: ReadString: EOF",
	}}

	for _, tt := range tests {
		got := tt.err.Error()
		if got != tt.want {
			t.Errorf("%#v.Error mismatch:\ngot  %v\nwant %v", tt.err, got, tt.want)
		}
	}
}

func TestSemanticError(t *testing.T) {
	tests := []struct {
		err  error
		want string
	}{{
		err:  &SemanticError{},
		want: "thrift: cannot handle",
	}, {
		err:  &SemanticError{ThriftType: thriftwire.Bool},
		want: "thrift: cannot handle Thrift bool",
	}, {
		err:  &SemanticError{action: "unmarshal", ThriftType: thriftwire.Byte},
		want: "thrift: cannot unmarshal Thrift byte",
	}, {
		err:  &SemanticError{action: "unmarshal", ThriftType: thriftwire.Stop},
		want: "thrift: cannot unmarshal", // invalid Thrift types are ignored
	}, {
		err:  &SemanticError{action: "marshal", ThriftType: thriftwire.Double},
		want: "thrift: cannot marshal Thrift double",
	}, {
		err:  &SemanticError{GoType: reflect.TypeOf(bool(false))},
		want: "thrift: cannot handle Go value of type bool",
	}, {
		err:  &SemanticError{action: "marshal", GoType: reflect.TypeOf(int(0))},
		want: "thrift: cannot marshal Go value of type int",
	}, {
		err:  &SemanticError{action: "unmarshal", GoType: reflect.TypeOf(uint(0))},
		want: "thrift: cannot unmarshal Go value of type uint",
	}, {
		err:  &SemanticError{ThriftType: thriftwire.I16, GoType: reflect.TypeOf(tar.Header{})},
		want: "thrift: cannot handle Thrift i16 with Go value of type tar.Header",
	}, {
		err:  &SemanticError{action: "marshal", ThriftType: thriftwire.I32, GoType: reflect.TypeOf(bytes.Buffer{})},
		want: "thrift: cannot marshal Thrift i32 from Go value of type bytes.Buffer",
	}, {
		err:  &SemanticError{action: "unmarshal", ThriftType: thriftwire.I64, GoType: reflect.TypeOf(strings.Reader{})},
		want: "thrift: cannot unmarshal Thrift i64 into Go value of type strings.Reader",
	}, {
		err:  &SemanticError{action: "unmarshal", ThriftType: thriftwire.String, GoType: reflect.TypeOf(float64(0))},
		want: "thrift: cannot unmarshal Thrift string into Go value of type float64",
	}, {
		err:  &SemanticError{action: "marshal", ThriftType: thriftwire.Struct, GoType: reflect.TypeOf(complex128(0))},
		want: "thrift: cannot marshal Thrift struct from Go value of type complex128",
	}, {
		err:  &SemanticError{action: "unmarshal", ThriftType: thriftwire.Map, GoType: reflect.TypeOf((*io.Reader)(nil)).Elem(), Err: errors.New("some underlying error")},
		want: "thrift: cannot unmarshal Thrift map into Go value of type io.Reader: some underlying error",
	}, {
		err:  &SemanticError{Err: errors.New("some underlying error")},
		want: "thrift: cannot handle: some underlying error",
	}, {
		err:  &SemanticError{action: "marshal", ThriftType: thriftwire.Set},
		want: "thrift: cannot marshal Thrift set",
	}, {
		err:  &SemanticError{action: "marshal", ThriftType: thriftwire.List},
		want: "thrift: cannot marshal Thrift list",
	}, {
		err:  &SemanticError{action: "marshal", ThriftType: thriftwire.UUID},
		want: "thrift: cannot marshal Thrift UUID",
	}}

	for _, tt := range tests {
		got := tt.err.Error()
		// Cleanup the error of non-deterministic rendering effects.
		if strings.HasPrefix(got, errorPrefix+"unable to ") {
			got = errorPrefix + "cannot " + strings.TrimPrefix(got, errorPrefix+"unable to ")
		}
		if got != tt.want {
			t.Errorf("%#v.Error mismatch:\ngot  %v\nwant %v", tt.err, got, tt.want)
		}
	}
}
