package thriftcompact

import (
	"testing"

	"github.com/itstarsun/go-thrift/encoding/thriftwire"
)

func TestCompactType(t *testing.T) {
	for _, tt := range []struct {
		ct compactType
		wt thriftwire.Type
	}{
		{_STOP, thriftwire.Stop},
		{_BOOLEAN_TRUE, thriftwire.Bool},
		{_BOOLEAN_FALSE, thriftwire.Bool},
		{_I8, thriftwire.Byte},
		{_I16, thriftwire.I16},
		{_I32, thriftwire.I32},
		{_I64, thriftwire.I64},
		{_DOUBLE, thriftwire.Double},
		{_BINARY, thriftwire.String},
		{_LIST, thriftwire.List},
		{_SET, thriftwire.Set},
		{_MAP, thriftwire.Map},
		{_STRUCT, thriftwire.Struct},
		{_UUID, thriftwire.UUID},
	} {
		t.Run(tt.ct.String(), func(t *testing.T) {
			if wt, err := tt.ct.wire(); err != nil || wt != tt.wt {
				t.Errorf("%s.wire() = (%s, %v), want %s", tt.ct, wt, err, tt.wt)
			}

			if tt.wt != thriftwire.Stop && tt.wt != thriftwire.Bool {
				if ct, err := typeCompact(tt.wt); err != nil || ct != tt.ct {
					t.Errorf("typeCompact(%s) = (%s, %v), want %s", tt.wt, ct, err, tt.ct)
				}
			}
		})
	}
}
