package checks

import (
	"net"
	"testing"
)

func TestCanConnect(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	port := ln.Addr().(*net.TCPAddr).Port

	if !CanConnect("127.0.0.1", port) {
		t.Error("expected to connect to a listening port")
	}

	ln.Close()
	if CanConnect("127.0.0.1", port) {
		t.Error("expected failure once the listener is closed")
	}
}
