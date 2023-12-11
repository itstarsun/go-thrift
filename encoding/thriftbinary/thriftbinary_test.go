package thriftbinary

import (
	"testing"

	"github.com/itstarsun/go-thrift/testing/thrifttest"
)

var protocolOptions = thrifttest.ProtocolOptions{
	UUID: true,
}

func TestProtocol(t *testing.T) {
	thrifttest.TestProtocol(t, Protocol, protocolOptions)
}

func TestProtocolNonStrict(t *testing.T) {
	thrifttest.TestProtocol(t, ProtocolNonStrict, protocolOptions)
}
