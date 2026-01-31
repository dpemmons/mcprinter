package transport_test

import (
	"io"
	"net"
	"testing"
	"time"

	"github.com/dpemmons/mcprinter/internal/transport"
)

func TestSend_WritesDataToServer(t *testing.T) {
	// Start a TCP listener to simulate a printer
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	received := make(chan []byte, 1)
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		data, _ := io.ReadAll(conn)
		received <- data
	}()

	addr := ln.Addr().String()
	host, port, _ := net.SplitHostPort(addr)

	err = transport.Send(host, port, []byte("hello printer"))
	if err != nil {
		t.Fatalf("Send failed: %v", err)
	}

	select {
	case data := <-received:
		if string(data) != "hello printer" {
			t.Errorf("got %q, want %q", string(data), "hello printer")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for data")
	}
}

func TestSend_ErrorOnUnreachableHost(t *testing.T) {
	// Use a port that nothing is listening on
	err := transport.Send("127.0.0.1", "1", []byte("data"))
	if err == nil {
		t.Fatal("expected error for unreachable host")
	}
}
