package thriftbinary

import "github.com/itstarsun/go-thrift/encoding/thriftwire"

func Example_strictWriteNonStrictRead() {
	_ = thriftwire.JoinProtocol(ProtocolNonStrict, Protocol)
}
