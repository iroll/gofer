// gofer_engine 0.9.5
// network handler for gofer
// (C) 2025 Isaac Roll
// See github.com/iroll/gofer for license

package main

import (
	"fmt"
	"io"
	"net"
	"net/url"
	"strings"
	"time"
)

type GopherTarget struct {
	Host     string
	Port     string
	Selector string
	Type     byte
}

type GopherResponse struct {
	Target   GopherTarget
	RawText  string // populated for text/menu
	RawBytes []byte // populated for binary
	Opaque   bool
}

// normalizeGopherInput: converts the gopher URI structure to an
// internal http structure to bridge with the system browser

func normalizeGopherInput(raw string) (GopherTarget, error) {
	var gt GopherTarget

	input := strings.TrimSpace(raw)
	if input == "" {
		return gt, fmt.Errorf("empty gopher input")
	}

	// 1. Force scheme
	if !strings.Contains(input, "://") {
		input = "gopher://" + input
	}

	u, err := url.Parse(input)
	if err != nil {
		return gt, fmt.Errorf("invalid gopher URI: %w", err)
	}

	// 2. Host recovery
	host := u.Hostname()
	if host == "" {
		// Try path-as-host fallback (e.g. gopher://sdf.org)
		if !strings.Contains(u.Path, "/") {
			host = u.Path
			u.Path = ""
		}
	}

	if host == "" {
		return gt, fmt.Errorf("missing gopher host")
	}

	// 3. Port default
	port := u.Port()
	if port == "" {
		port = DEFAULT_GOPHER_PORT
	}

	// 4. Selector normalization (Go-version stable)
	rawPath := u.EscapedPath()

	selector := ""
	switch rawPath {
	case "", "/":
		selector = ""
	default:
		selector = strings.TrimPrefix(rawPath, "/")
	}

	// 5. Type inference
	var t byte = '1'
	if selector != "" && selector[0] >= '0' && selector[0] <= '9' {
		t = selector[0]
	}

	gt = GopherTarget{
		Host:     host,
		Port:     port,
		Selector: selector,
		Type:     t,
	}

	return gt, nil
}

// gopherRequestBytes connects to a remote Gopher server, sends the selector, and returns raw bytes.
// This is the default case ("byte pipline"); for menu/text handling use gopherRequest ("text pipeline").

func gopherRequestBytes(host string, port string, selector string) ([]byte, error) {
	address := net.JoinHostPort(host, port)

	conn, err := net.DialTimeout("tcp", address, TCP_TIMEOUT)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Gopher server %s: %w", address, err)
	}
	defer conn.Close()

	conn.SetDeadline(time.Now().Add(TCP_TIMEOUT))

	request := selector + GOPHER_REQUEST_TERMINATOR

	if _, err := conn.Write([]byte(request)); err != nil {
		return nil, fmt.Errorf("failed to write selector to socket: %w", err)
	}

	// Read everything until EOF / timeout
	b, err := io.ReadAll(conn)
	if err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			return nil, fmt.Errorf("socket timeout while reading from %s", address)
		}
		return nil, fmt.Errorf("error reading from socket: %w", err)
	}

	return b, nil
}

// gopherRequest ("text pipeline") returns text for menu/text handling.
// This is NOT safe for binary. Use gopherRequestBytes for that.

func gopherRequest(host string, port string, selector string) (string, error) {
	b, err := gopherRequestBytes(host, port, selector)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// helper to determine whether a link goes to the text pipeline or the byte pipeline
func usesTextPipeline(t byte) bool {
	switch t {
	case '0', '1':
		return true
	default:
		return false
	}
}

// Fetch: This function becomes the sole authority for:
// - deciding opaque vs transparent
// - choosing text vs bytes
// - returning a coherent result

func Fetch(target GopherTarget) (GopherResponse, error) {
	var resp GopherResponse
	resp.Target = target

	opaque := !usesTextPipeline(target.Type)
	resp.Opaque = opaque

	if opaque {
		b, err := gopherRequestBytes(target.Host, target.Port, target.Selector)
		if err != nil {
			return resp, err
		}
		resp.RawBytes = b
		return resp, nil
	}

	// transparent (text / menu)
	txt, err := gopherRequest(target.Host, target.Port, target.Selector)
	if err != nil {
		return resp, err
	}
	resp.RawText = txt
	return resp, nil
}
