package thriftwire_test

import (
	"math"
	"testing"

	"github.com/itstarsun/go-thrift/encoding/thriftwire"
	"github.com/itstarsun/go-thrift/internal/thriftmemo"
)

func must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

func TestSkip(t *testing.T) {
	for _, tt := range []struct {
		t     thriftwire.Type
		steps []string
	}{{
		t:     thriftwire.Bool,
		steps: []string{"Bool"},
	}, {
		t:     thriftwire.Byte,
		steps: []string{"Byte"},
	}, {
		t:     thriftwire.Double,
		steps: []string{"Double"},
	}, {
		t:     thriftwire.I16,
		steps: []string{"I16"},
	}, {
		t:     thriftwire.I32,
		steps: []string{"I32"},
	}, {
		t:     thriftwire.I64,
		steps: []string{"I64"},
	}, {
		t:     thriftwire.String,
		steps: []string{"String"},
	}, {
		t: thriftwire.Struct,
		steps: []string{
			"StructBegin",
			"FieldBegin",
			"Bool",
			"FieldEnd",
			"FieldBegin",
			"String",
			"FieldEnd",
			"FieldBegin",
			"StructEnd",
		},
	}, {
		t:     thriftwire.Map,
		steps: []string{"MapBegin", "Bool", "Bool", "Bool", "Bool", "MapEnd"},
	}, {
		t:     thriftwire.Set,
		steps: []string{"SetBegin", "Bool", "Bool", "SetEnd"},
	}, {
		t:     thriftwire.List,
		steps: []string{"ListBegin", "Bool", "Bool", "ListEnd"},
	}, {
		t:     thriftwire.UUID,
		steps: []string{"UUID"},
	}} {
		t.Run(tt.t.String(), func(t *testing.T) {
			var m thriftmemo.Memo

			w := m.Writer()
			switch tt.t {
			case thriftwire.Bool:
				must(t, w.WriteBool(true))
			case thriftwire.Byte:
				must(t, w.WriteByte(math.MaxUint8))
			case thriftwire.Double:
				must(t, w.WriteDouble(math.MaxFloat64))
			case thriftwire.I16:
				must(t, w.WriteI16(math.MaxInt16))
			case thriftwire.I32:
				must(t, w.WriteI32(math.MaxInt32))
			case thriftwire.I64:
				must(t, w.WriteI64(math.MaxInt64))
			case thriftwire.String:
				must(t, w.WriteString(""))
			case thriftwire.Struct:
				must(t, w.WriteStructBegin(thriftwire.StructHeader{}))
				must(t, w.WriteFieldBegin(thriftwire.FieldHeader{Type: thriftwire.Bool}))
				must(t, w.WriteBool(true))
				must(t, w.WriteFieldEnd())
				must(t, w.WriteFieldBegin(thriftwire.FieldHeader{Type: thriftwire.String}))
				must(t, w.WriteString(""))
				must(t, w.WriteFieldEnd())
				must(t, w.WriteStructEnd())
			case thriftwire.Map:
				must(t, w.WriteMapBegin(thriftwire.MapHeader{
					Key:   thriftwire.Bool,
					Value: thriftwire.Bool,
					Size:  2,
				}))
				must(t, w.WriteBool(true))
				must(t, w.WriteBool(false))
				must(t, w.WriteBool(false))
				must(t, w.WriteBool(true))
				must(t, w.WriteMapEnd())
			case thriftwire.Set:
				must(t, w.WriteSetBegin(thriftwire.SetHeader{
					Element: thriftwire.Bool,
					Size:    2,
				}))
				must(t, w.WriteBool(true))
				must(t, w.WriteBool(false))
				must(t, w.WriteSetEnd())
			case thriftwire.List:
				must(t, w.WriteListBegin(thriftwire.ListHeader{
					Element: thriftwire.Bool,
					Size:    2,
				}))
				must(t, w.WriteBool(true))
				must(t, w.WriteBool(false))
				must(t, w.WriteListEnd())
			case thriftwire.UUID:
				var v [16]byte
				must(t, w.WriteUUID(&v))
			default:
				panic("invalid type: " + tt.t.String())
			}

			r := m.Reader()
			if err := thriftwire.Skip(r, tt.t); err != nil {
				t.Fatal(err)
			}
		})
	}
}
