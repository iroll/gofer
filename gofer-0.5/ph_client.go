// Ph Client for gofer 0.5
// hewing as close to RFC 1436 (1998) as practical
// (C) 2025 Isaac Roll
// See github.com/iroll/gofer for license

package main

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
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
	trimmed := strings.TrimPrefix(path, "/ph/")
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

func HandlePH(w http.ResponseWriter, r *http.Request) {
	host, port, err := ParsePHRoute(r.URL.Path)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	returnURL := r.URL.Query().Get("return")
	if returnURL == "" {
		returnURL = "/"
	}

	var content string

	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Invalid form data", http.StatusBadRequest)
			return
		}

		query := strings.TrimSpace(r.FormValue("query"))
		if query == "" {
			http.Error(w, "Empty query", http.StatusBadRequest)
			return
		}

		result, err := PHQuery(host, port, query)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}

		content = result

	} else {
		greeting, err := PHInitialGreeting(host, port)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}

		content = greeting
	}

	html := formatPHPage(host, port, content, returnURL)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html))
}

func PHQuery(host, port, query string) (string, error) {
	address := net.JoinHostPort(host, port)

	conn, err := net.DialTimeout("tcp", address, PH_TIMEOUT)
	if err != nil {
		return "", err
	}
	defer conn.Close()

	conn.SetDeadline(time.Now().Add(PH_TIMEOUT))
	reader := bufio.NewReader(conn)

	// Read greeting (and ignore content)
	_, err = reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	// Send query
	fmt.Fprintf(conn, "query %s\r\n", query)

	// Read response
	var out strings.Builder
	for {
		line, err := reader.ReadString('\n')
		if len(line) > 0 {
			out.WriteString(line)
		}
		if err != nil {
			break
		}
	}

	return strings.TrimSpace(out.String()), nil
}

// HTML UI formatting function
func formatPHPage(host, port, content, returnURL string) string {
	var html strings.Builder

	html.WriteString(fmt.Sprintf(`
		<!DOCTYPE html>
		<html>
		<head>
			<title>gofer PhClient - %s:%s</title>
			<style>

				:root { color-scheme: light dark; }

				body {
					font-family: monospace;
					line-height: 1.4;
					width: 100ch;
					margin: 0 auto;
					padding-bottom: 1ch;
				}
					
				.return { 
					margin-top: 1ch;
				} 

				.query-bar {
					width: 100%%;
					margin: 1ch 0 1ch 0;		
				}
				
				.query-bar form {
        			display: flex; /* Activate Flexbox */
        			width: 100%%; /* Ensure the form uses the full 100ch of .query-bar */
        			align-items: center; /* Vertically center the text and input */
    			}

				.query-label {
					font-size: 1.5em;
					font-weight: bold;
					padding: 0 1ch 0 0;
					flex-shrink: 0;
				}

				input[type="text"] {
					font-family: monospace;
					font-size: 1.5em;
					font-weight: bold;

					flex-grow: 1;
					min-width: 0; 

					outline: 0;
					caret-style: underscore;
				}

				pre {
					width: 100%%;
					padding: 0 0 1ch 0;
					white-space: pre;
				}
				
			</style>
		</head>
		<body>

		<div class="query-bar">
			<form method="POST">
				<span class="query-label">query</span>
				<input type="text" name="query" autofocus>
			</form>
		</div>

		<pre>%s</pre>

		<div class="return">
			<a href="%s">Exit PhClient</a>
		</div>

		</body>
		</html>
`, host, port, content, returnURL))

	return html.String()
}
