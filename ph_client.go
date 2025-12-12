package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"time"
)

const PH_DEFAULT_PORT = "105"
const PH_TIMEOUT = 5 * time.Second

// -----------------------------------------------------------
// ParsePHRoute("/ph:hostname:port") -> host, port
// -----------------------------------------------------------
func ParsePHRoute(path string) (string, string, error) {
	// Expecting: "/ph:hostname:port"
	trimmed := strings.TrimPrefix(path, "/ph:")
	parts := strings.Split(trimmed, ":")

	if len(parts) < 1 {
		return "", "", fmt.Errorf("invalid PH route: %s", path)
	}

	host := parts[0]
	port := PH_DEFAULT_PORT

	if len(parts) > 1 && parts[1] != "" {
		port = parts[1]
	}

	return host, port, nil
}

// -----------------------------------------------------------
// PHInitialGreeting(host, port) -> string
//
// Connects to PH server, reads greeting line, closes socket.
// -----------------------------------------------------------
func PHInitialGreeting(host, port string) (string, error) {
	address := net.JoinHostPort(host, port)

	conn, err := net.DialTimeout("tcp", address, PH_TIMEOUT)
	if err != nil {
		return "", fmt.Errorf("PH connect failed: %w", err)
	}
	defer conn.Close()

	conn.SetDeadline(time.Now().Add(PH_TIMEOUT))

	reader := bufio.NewReader(conn)

	greeting, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("PH read failed: %w", err)
	}

	// PH greeting lines may have trailing CRLF → trim it
	return strings.TrimSpace(greeting), nil
}
