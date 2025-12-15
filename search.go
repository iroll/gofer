// search module for gofer 0.5
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

func HandleSearch(w http.ResponseWriter, r *http.Request) {
	host := r.URL.Query().Get("host")
	port := r.URL.Query().Get("port")
	selector := r.URL.Query().Get("selector")

	if host == "" || port == "" || selector == "" {
		http.Error(w, "Missing host, port, or selector", http.StatusBadRequest)
		return
	}

	returnURL := r.URL.Query().Get("return")
	if returnURL == "" {
		returnURL = "/"
	}

	switch r.Method {

	case http.MethodGet:
		// GET = search landing page (no TCP, no menu)
		html := renderSearchFrame("", host, port, selector, returnURL)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(html))
		return

	case http.MethodPost:
		//POST = execute search
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Invalid form data", http.StatusBadRequest)
			return
		}

		query := strings.TrimSpace(r.FormValue("query"))
		if query == "" {
			http.Error(w, "Empty query", http.StatusBadRequest)
			return
		}

		rawMenu, err := SearchQuery(host, port, selector, query)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}

		menuHTML := formatMenuHTML(rawMenu, host, port, selector, true)
		html := renderSearchFrame(menuHTML, host, port, selector, returnURL)

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(html))
		return

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
}

func SearchQuery(host, port, selector, query string) (string, error) {
	address := net.JoinHostPort(host, port)

	conn, err := net.DialTimeout("tcp", address, TCP_TIMEOUT)
	if err != nil {
		return "", err
	}
	defer conn.Close()

	conn.SetDeadline(time.Now().Add(TCP_TIMEOUT))
	reader := bufio.NewReader(conn)

	// Send query
	fmt.Fprintf(conn, "%s\t%s\r\n", selector, query)

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
func renderSearchFrame(innerHTML, host, port, selector, returnURL string) string {
	var html strings.Builder

	_ = selector

	html.WriteString(fmt.Sprintf(`
		<!DOCTYPE html>
		<html>
		<head>
			<title>gofer search - %s:%s</title>
			<style>

				:root { color-scheme: light dark; }

				body {
					font-family: monospace;
					line-height: 1.4;
					width: 100ch;
					margin: 0 auto;
					padding-bottom: 1ch;
				}
				
				.gopher-link { 
					margin: 0;
				 	white-space: pre;
				} 

				.gopher-link:last-child {
					margin-bottom: 1ch;
				}
				
				.return { 
					margin-top: 1ch;
				} 				
				
				.results { 
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
				
			</style>
		</head>
		<body>

		<div class="query-bar">
			<form method="POST">
				<span class="query-label">query</span>
				<input type="text" name="query" autofocus>
			</form>
		</div>

		<div class="results">
			%s
		</div>

		<div class="return">
			<a href="%s">Exit Search</a>
		</div>

		</body>
		</html>
`, host, port, innerHTML, returnURL))

	return html.String()
}
