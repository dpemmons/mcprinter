package transport

import (
	"fmt"
	"net"
	"time"
)

// Send connects to host:port via TCP and writes data.
func Send(host, port string, data []byte) error {
	addr := net.JoinHostPort(host, port)
	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		return fmt.Errorf("cannot reach printer at %s: %w", addr, err)
	}
	defer conn.Close()

	conn.SetWriteDeadline(time.Now().Add(30 * time.Second))
	_, err = conn.Write(data)
	if err != nil {
		return fmt.Errorf("failed writing to printer: %w", err)
	}
	return nil
}
