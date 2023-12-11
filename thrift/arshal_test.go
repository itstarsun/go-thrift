package thrift

import (
	"errors"
	"math"
	"reflect"
	"testing"

	"github.com/itstarsun/go-thrift/internal/thriftmemo"
)

func ptr[T any](v T) *T {
	return &v
}

func slicesEqual[S ~[]E, E comparable](s1, s2 S) bool {
	// TODO(go1.21): Use slices.Equal.
	if len(s1) != len(s2) {
		return false
	}
	for i := range s1 {
		if s1[i] != s2[i] {
			return false
		}
	}
	return true
}

type aStruct struct {
	String string     `thrift:"1"`
	List   []*aStruct `thrift:"2"`
}

type aRequiredField struct {
	RequiredField string `thrift:"1,required"`
}

func TestArshaler(t *testing.T) {
	for _, tt := range []struct {
		in    any
		steps []string
	}{{
		in:    true,
		steps: []string{"Bool"},
	}, {
		in:    int8(math.MaxInt8),
		steps: []string{"Byte"},
	}, {
		in:    int16(math.MaxInt16),
		steps: []string{"I16"},
	}, {
		in:    int32(math.MaxInt32),
		steps: []string{"I32"},
	}, {
		in:    int64(math.MaxInt64),
		steps: []string{"I64"},
	}, {
		in:    uint8(math.MaxUint8),
		steps: []string{"Byte"},
	}, {
		in:    uint16(math.MaxUint16),
		steps: []string{"I16"},
	}, {
		in:    uint32(math.MaxUint32),
		steps: []string{"I32"},
	}, {
		in:    uint64(math.MaxUint64),
		steps: []string{"I64"},
	}, {
		in:    "hello, world!",
		steps: []string{"String"},
	}, {
		in:    []byte("hello, world!"),
		steps: []string{"Bytes"},
	}, {
		in:    [16]byte{},
		steps: []string{"UUID"},
	}, {
		in:    map[bool]bool{false: false, true: true},
		steps: []string{"MapBegin", "Bool", "Bool", "Bool", "Bool", "MapEnd"},
	}, {
		in:    Set[float32]{0, 1, 0},
		steps: []string{"SetBegin", "Double", "Double", "Double", "SetEnd"},
	}, {
		in:    List[bool]{false, true, false},
		steps: []string{"ListBegin", "Bool", "Bool", "Bool", "ListEnd"},
	}, {
		in:    []*string{ptr("hello"), ptr("world"), (*string)(nil)},
		steps: []string{"ListBegin", "String", "String", "String", "ListEnd"},
	}, {
		in:    map[string]string(nil),
		steps: []string{"MapBegin", "MapEnd"},
	}, {
		in:    []string(nil),
		steps: []string{"ListBegin", "ListEnd"},
	}, {
		in: aStruct{
			String: "hello, world!",
			List:   []*aStruct{{}},
		},
		steps: []string{
			"StructBegin",

			"FieldBegin",
			"String",
			"FieldEnd",

			"FieldBegin",
			"ListBegin",

			"StructBegin",
			"FieldStop",
			"StructEnd",

			"ListEnd",
			"FieldEnd",

			"FieldStop",
			"StructEnd",
		},
	}, {
		in: aRequiredField{},
		steps: []string{
			"StructBegin",
			"FieldBegin",
			"String",
			"FieldEnd",
			"FieldStop",
			"StructEnd",
		},
	}} {
		for i, step := range tt.steps {
			if step == "FieldStop" {
				tt.steps[i] = "FieldBegin"
			}
		}

		out := reflect.New(reflect.TypeOf(tt.in))
		for n := len(tt.steps); n >= 0; n-- {
			expectError := n != len(tt.steps)

			var m thriftmemo.Memo
			if expectError {
				m.Breakpoint = n + 1
			}

			if err := Marshal(m.Writer(), tt.in); err != nil && !expectError {
				t.Fatalf("unexpected error: %v", err)
			} else if expectError {
				var se *SemanticError
				if !errors.As(err, &se) {
					t.Fatalf("got %T, want %T", err, se)
				}
				if se.action != "marshal" {
					t.Fatalf("got %q, want %q", se.action, "marshal")
				}
				var we *wireError
				if !errors.As(err, &we) {
					t.Fatalf("expect %T", we)
				}
				if we.action != "Write"+tt.steps[n] && we.action != "WriteStructEnd" {
					t.Fatalf("got %q, want %q", we.action, "Write"+tt.steps[n])
				}
			}

			write := m.Steps()
			if !slicesEqual(write, tt.steps[:n]) {
				t.Fatalf("\ngot  %v\nwant %v", write, tt.steps[:n])
			}

			if err := Unmarshal(m.Reader(), out.Interface()); err != nil && !expectError {
				t.Fatal(err)
			} else if expectError {
				var se *SemanticError
				if !errors.As(err, &se) {
					t.Fatalf("got %T, want %T", err, se)
				}
				if se.action != "unmarshal" {
					t.Fatalf("got %q, want %q", se.action, "unmarshal")
				}
				var we *wireError
				if !errors.As(err, &we) {
					t.Fatalf("expect %T", we)
				}
				if we.action != "Read"+tt.steps[n] {
					t.Fatalf("got %q, want %q", we.action, "Read"+tt.steps[n])
				}
			}

			read := m.Steps()[len(write):]
			if !slicesEqual(read, write) {
				t.Errorf("\ngot  %v\nwant %v", read, write)
			}
		}
	}
}
