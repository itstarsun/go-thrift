package thrifttest

import (
	"bytes"
	"math"
	"reflect"
	"testing"

	"github.com/itstarsun/go-thrift/encoding/thriftwire"
	"github.com/itstarsun/go-thrift/thrift"
)

// ProtocolOptions is the options passed to [TestProtocol].
type ProtocolOptions struct {
	UUID bool // whether the protocol support UUID
}

type boolStruct struct {
	Bool bool `thrift:"1,required"`
}

type floatStruct struct {
	Float32 float32 `thrift:"1"`
	Float64 float64 `thrift:"2"`
}

type intStruct struct {
	Int16 int16 `thrift:"1"`
	Int32 int32 `thrift:"2"`
	Int64 int64 `thrift:"3"`
}

type testStruct struct {
	Struct *testStruct `thrift:"1"`
	String string      `thrift:"2"`
	Bytes  []byte      `thrift:"3"`
}

func isMessageNameEqual(x, y string) bool {
	return x == y || x == "" || y == ""
}

func isMessageHeaderEqual(x, y thriftwire.MessageHeader) bool {
	return x.Type == y.Type && x.ID == y.ID && isMessageNameEqual(x.Name, y.Name)
}

// TestProtocol tests the [thriftwire.Protocol] implementation.
func TestProtocol(t *testing.T, p thriftwire.Protocol, opts ProtocolOptions) {
	var msg thriftwire.MessageHeader

	var b bytes.Buffer
	w := p.NewWriter(&b)
	r := p.NewReader(&b)

	for _, tt := range []struct {
		name string
		in   any

		hasUUID bool
		hasMap  bool
	}{{
		name: "False",
		in:   boolStruct{Bool: false},
	}, {
		name: "True",
		in:   boolStruct{Bool: true},
	}, {
		name: "SmallestNonzeroFloats",
		in: floatStruct{
			Float32: math.SmallestNonzeroFloat32,
			Float64: math.SmallestNonzeroFloat64,
		},
	}, {
		name: "MaxFloats",
		in: floatStruct{
			Float32: math.MaxFloat32,
			Float64: math.MaxFloat64,
		},
	}, {
		name: "MinInts",
		in: intStruct{
			Int16: math.MinInt16,
			Int32: math.MinInt32,
			Int64: math.MinInt64,
		},
	}, {
		name: "MaxInts",
		in: intStruct{
			Int16: math.MaxInt16,
			Int32: math.MaxInt32,
			Int64: math.MaxInt64,
		},
	}, {
		name: "ZeroStringMap",
		in: struct {
			Value map[string]string `thrift:"1,required"`
		}{nil},
	}, {
		name: "StringMap",
		in: struct {
			Value map[string]string `thrift:"1,required"`
		}{map[string]string{"hello": "world"}},
	}, {
		name: "ZeroStringSet",
		in: struct {
			Value thrift.Set[string] `thrift:"1,required"`
		}{nil},
	}, {
		name: "StringSet",
		in: struct {
			Value thrift.Set[string] `thrift:"1,required"`
		}{[]string{"hello", "world"}},
	}, {
		name: "ZeroStringList",
		in: struct {
			Value thrift.List[string] `thrift:"1,required"`
		}{nil},
	}, {
		name: "StringList",
		in: struct {
			Value thrift.List[string] `thrift:"1,required"`
		}{[]string{"hello", "world"}},
	}, {
		name: "UUID",
		in: struct {
			UUID [16]byte `thrift:"1,required"`
		}{},
		hasUUID: true,
	}, {
		name: "NestedStruct",
		in: testStruct{
			String: "1",
			Bytes:  []byte("1"),
			Struct: &testStruct{
				String: "2",
				Bytes:  []byte("2"),
				Struct: &testStruct{
					String: "3",
					Bytes:  []byte("3"),
				},
			},
		},
	}, {
		name: "LargeSetListOfBools",
		in: struct {
			Set  thrift.Set[bool]  `thrift:"32"`
			List thrift.List[bool] `thrift:"64"`
		}{
			Set:  make(thrift.Set[bool], 64),
			List: make(thrift.List[bool], 64),
		},
	}, {
		name: "LargeMapOfBools",
		in: struct {
			Map map[string]bool `thrift:"0"`
		}{
			Map: func() map[string]bool {
				m := make(map[string]bool, 'Z'-'A'+1)
				for c := 'A'; c <= 'Z'; c++ {
					m[string(c)] = true
				}
				return m
			}(),
		},
		hasMap: true,
	}} {
		t.Run(tt.name, func(t *testing.T) {
			if tt.hasUUID && !opts.UUID {
				t.Skipf("UUID is not supported")
			}

			msg.Name = tt.name
			msg.Type %= thriftwire.OneWay
			msg.Type++
			msg.ID++

			w.Reset(&b)
			marshal(t, w, msg, tt.in)
			want := append([]byte(nil), b.Bytes()...)

			r.Reset(&b)
			if h, err := r.ReadMessageBegin(); err != nil {
				t.Fatal(err)
			} else if !isMessageHeaderEqual(h, msg) {
				t.Errorf("\ngot  %#v\nwant %#v", h, msg)
			}
			out := reflect.New(reflect.TypeOf(tt.in)).Interface()
			if err := thrift.Unmarshal(r, out); err != nil {
				t.Fatal(err)
			}
			if err := r.ReadMessageEnd(); err != nil {
				t.Fatal(err)
			}
			if b.Len() != 0 {
				t.Fatalf("message is not fully consumed")
			}

			if !tt.hasMap {
				for i := 0; i < 3; i++ {
					b.Reset()
					marshal(t, w, msg, out)
					if got := b.Bytes(); !bytes.Equal(got, want) {
						t.Errorf("\ngot  %q\nwant %q", got, want)
					}
				}
			}

			r.Reset(&b)
			for i := 0; i < 3; i++ {
				b.Write(want)
				if _, err := r.ReadMessageBegin(); err != nil {
					t.Fatal(err)
				}
				if err := thriftwire.Skip(r, thriftwire.Struct); err != nil {
					t.Fatal(err)
				}
				if err := r.ReadMessageEnd(); err != nil {
					t.Fatal(err)
				}
			}
			b.Reset()
		})
	}
}

func marshal(t testing.TB, w thriftwire.Writer, msg thriftwire.MessageHeader, in any) {
	t.Helper()
	if err := w.WriteMessageBegin(msg); err != nil {
		t.Fatal(err)
	}
	if err := thrift.Marshal(w, in); err != nil {
		t.Fatal(err)
	}
	if err := w.WriteMessageEnd(); err != nil {
		t.Fatal(err)
	}
	if err := w.Flush(); err != nil {
		t.Fatal(err)
	}
}
