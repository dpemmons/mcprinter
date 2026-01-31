package main_test

import (
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

func TestIntegration_TextPrint(t *testing.T) {
	// Build binary
	dir := t.TempDir()
	bin := filepath.Join(dir, "mcprinter")
	build := exec.Command("go", "build", "-o", bin, ".")
	build.Dir = getProjectRoot(t)
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build failed: %s\n%s", err, out)
	}

	// Start fake printer
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

	host, port, _ := net.SplitHostPort(ln.Addr().String())

	cmd := exec.Command(bin, "--host", host, "--port", port, "Hello, printer!")
	cmd.Env = append(os.Environ(), "PRINTER_HOST=", "PRINTER_PORT=")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("mcprinter failed: %s\n%s", err, out)
	}

	select {
	case data := <-received:
		if len(data) == 0 {
			t.Fatal("received empty data")
		}
		// Should contain the text somewhere in the ESC/POS stream
		if !containsBytes(data, []byte("Hello, printer!")) {
			t.Error("printed data does not contain expected text")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timeout waiting for print data")
	}
}

func containsBytes(haystack, needle []byte) bool {
	for i := 0; i <= len(haystack)-len(needle); i++ {
		match := true
		for j := 0; j < len(needle); j++ {
			if haystack[i+j] != needle[j] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

func getProjectRoot(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	return wd
}
