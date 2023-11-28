package thriftmemo

import (
	"fmt"
	"math"
	"testing"

	"github.com/itstarsun/go-thrift/encoding/thriftwire"
)

func must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

func eq[T comparable](t *testing.T, x, y T) {
	t.Helper()
	if x != y {
		t.Fatalf("got %v, want %v", x, y)
	}
}

func TestMemo(t *testing.T) {
	for _, tt := range []struct {
		values []any
	}{{
		values: []any{"", "hello, world!"},
	}, {
		values: []any{false, true},
	}, {
		values: []any{
			true,
			float32(math.MaxFloat32),
			float64(math.MaxFloat64),
			int8(math.MaxInt8),
			int16(math.MaxInt16),
			int32(math.MaxInt32),
			int64(math.MaxInt64),
		},
	}, {
		values: []any{
			[16]byte{},
		},
	}} {
		var m Memo

		w := m.Writer()
		for _, want := range tt.values {
			switch want := want.(type) {
			case bool:
				must(t, w.WriteBool(want))
			case float32:
				must(t, w.WriteDouble(float64(want)))
			case float64:
				must(t, w.WriteDouble(want))
			case int8:
				must(t, w.WriteByte(byte(want)))
			case int16:
				must(t, w.WriteI16(want))
			case int32:
				must(t, w.WriteI32(want))
			case int64:
				must(t, w.WriteI64(want))
			case string:
				must(t, w.WriteString(want))
			case [16]byte:
				must(t, w.WriteUUID(&want))
			default:
				panic(fmt.Sprintf("invalid type: %T", want))
			}
		}

		r := m.Reader()
		for _, want := range tt.values {
			switch want := want.(type) {
			case bool:
				got, err := r.ReadBool()
				if err != nil {
					t.Fatal(err)
				}
				eq(t, got, want)
			case float32:
				got, err := r.ReadDouble()
				if err != nil {
					t.Fatal(err)
				}
				eq(t, float32(got), want)
			case float64:
				got, err := r.ReadDouble()
				if err != nil {
					t.Fatal(err)
				}
				eq(t, got, want)
			case int8:
				got, err := r.ReadByte()
				if err != nil {
					t.Fatal(err)
				}
				eq(t, int8(got), want)
			case int16:
				got, err := r.ReadI16()
				if err != nil {
					t.Fatal(err)
				}
				eq(t, got, want)
			case int32:
				got, err := r.ReadI32()
				if err != nil {
					t.Fatal(err)
				}
				eq(t, got, want)
			case int64:
				got, err := r.ReadI64()
				if err != nil {
					t.Fatal(err)
				}
				eq(t, got, want)
			case string:
				got, err := r.ReadString()
				if err != nil {
					t.Fatal(err)
				}
				eq(t, got, want)
			case [16]byte:
				var got [16]byte
				must(t, r.ReadUUID(&got))
				eq(t, got, want)
			default:
				panic(fmt.Sprintf("invalid type: %T", want))
			}
		}
	}
}

func TestMemoStruct(t *testing.T) {
	var m Memo

	w := m.Writer()
	must(t, w.WriteStructBegin(thriftwire.StructHeader{}))
	must(t, w.WriteFieldBegin(thriftwire.FieldHeader{Type: thriftwire.Bool}))
	must(t, w.WriteBool(true))
	must(t, w.WriteFieldEnd())
	must(t, w.WriteStructEnd())

	r := m.Reader()
	_, err := r.ReadStructBegin()
	must(t, err)
	h, err := r.ReadFieldBegin()
	must(t, err)
	if h.Type != thriftwire.Bool {
		t.Fatalf("invalid field: %v", h)
	}
	v, err := r.ReadBool()
	must(t, err)
	if v != true {
		t.Fatalf("invalid bool: %t", v)
	}
	must(t, r.ReadFieldEnd())
	h, err = r.ReadFieldBegin()
	must(t, err)
	if h.Type != thriftwire.Stop {
		t.Fatalf("invalid field: %v", h)
	}
	must(t, r.ReadStructEnd())
}
