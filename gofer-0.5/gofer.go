// gofer 0.5
// a gopher helper for web browsers
// hewing as close to RFC 1436 (1993) as practical
// (C) 2025 Isaac Roll
// See github.com/iroll/gofer for license

package main

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"
)

// --- Configuration Constants ---
const (
	LOCAL_SERVER_PORT         = "8000"
	DEFAULT_GOPHER_HOST       = "freeshell.org"
	DEFAULT_GOPHER_PORT       = "70"
	SHUTDOWN_TIMEOUT_SECONDS  = 60
	TCP_TIMEOUT               = 5 * time.Second
	GOPHER_REQUEST_TERMINATOR = "\r\n"
	FOCUS_ENDPOINT            = "/focus"
)

// --- Inactivity Monitor ---
// a js heartbeat on 55 second intervals from a formatted HTML page keeps gofer alive
// otherwise, after 60 seconds of inactivity it terminates

var lastRequestTime = time.Now()
var shutdownMux sync.Mutex

func updateActivity() { // resets the inactivity timer. Called by all HTTP handlers.
	shutdownMux.Lock()
	lastRequestTime = time.Now()
	shutdownMux.Unlock()
}

func monitorInactivity() { // checks the time since the last request and shuts down if timed out.
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		shutdownMux.Lock()
		idleDuration := time.Since(lastRequestTime)
		shutdownMux.Unlock()

		if idleDuration > SHUTDOWN_TIMEOUT_SECONDS*time.Second {
			fmt.Println("No activity for", SHUTDOWN_TIMEOUT_SECONDS, "seconds. Shutting down...")
			os.Exit(0)
		}
	}
}

// --- Utility Functions ---

// launchBrowser opens the default web browser to the given URL.
func launchBrowser(url string) {
	var cmd *exec.Cmd

	// Use runtime.GOOS to get the OS the program is running on
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	case "darwin": // macOS
		cmd = exec.Command("open", url)
	default: // Linux (and others)
		cmd = exec.Command("xdg-open", url)
	}

	// use Start() to avoid blocking the main goroutine
	err := cmd.Start()
	if err != nil {
		fmt.Printf("Warning: Could not launch browser: %v\n", err)
	}
}

// gopherRequestBytes connects to a remote Gopher server, sends the selector, and returns raw bytes.
// Use this for binary types: 9, g, I, etc.
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

// gopherRequest returns text for menu/text handling.
// This is NOT safe for binary. Use gopherRequestBytes for that.
func gopherRequest(host string, port string, selector string) (string, error) {
	b, err := gopherRequestBytes(host, port, selector)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// helper to determine whether a link goes to the text pipeline or the byte pipeline
func isTransparentType(t byte) bool {
	switch t {
	case '0', '1':
		return true
	default:
		return false
	}
}

// --- HTML Formatting Component ---

// formatMenuHTML takes raw Gopher data and turns it into minimal HTML.
// It requires the current host, port, and selector for form pre-filling and links.
func formatMenuHTML(rawGopherData, currentHost, currentPort, currentSelector string, embedded bool) string {

	// Start with the HTML boilerplate, including the input form at the top
	var html strings.Builder

	// 1. Helper for writing the common item types
	writeLink := func(icon string, itemType byte, host, port, selector, display string) {
		link := fmt.Sprintf(
			"<a href=\"/?type=%c&host=%s&port=%s&selector=%s\">%s</a>",
			itemType, host, port, url.QueryEscape(selector), display,
		)
		html.WriteString(fmt.Sprintf(
			"<p class=\"gopher-link\">%s%s</p>\n",
			icon,
			link,
		))
	}

	// 1. Construct the current Gopher URI for the input field's value
	currentGopherURI := fmt.Sprintf("%s:%s%s", currentHost, currentPort, currentSelector)

	if !embedded {
		html.WriteString(fmt.Sprintf(`
		
	
		<!DOCTYPE html>
		<html>
		<head>
			<title>gofer - %s:%s%s</title>
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
					padding: 0 0 0 0;
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
			<form action="/" method="GET">
				<span class="query-label">gopher://</span>
				<input type="text" id="uri" name="uri" value="%s" placeholder="freeshell.org:70/">			
			</form>
		</div>

	`,
			// Arguments 1, 2, 3, 4: For the title and the URI input value
			currentHost, currentPort, currentSelector, currentGopherURI))
	}

	// --- End of the argument list ---

	// Process the lines from the Gopher response
	lines := strings.Split(rawGopherData, "\n")

	for _, line := range lines {
		// Check for empty lines
		if strings.TrimSpace(line) == "" {
			continue
		}

		// Check for Gopher EOF
		if strings.TrimSpace(line) == "." {
			break
		}

		// Gopher line format: TypeDisplayString\tSelector\tHost\tPort
		fields := strings.Split(line, "\t")

		var itemType byte
		var displayString, selector, host, port string

		// If the line is malformed, assume itemType 3

		if len(fields) < 4 {
			itemType = '3'
			displayString = "Malformed Line (Type 3 Error): " + strings.TrimSpace(line)
			selector = "/"
			host = currentHost
			port = currentPort
		} else {

			// 1. Extract Item Type and Display String
			itemType = fields[0][0]
			displayString = fields[0][1:]

			// 2. Extract Selector, Host, and Port
			selector = fields[1]
			host = fields[2]
			port = fields[3]
		}

		displayString = strings.TrimRight(displayString, " \t\r")

		if displayString == "" {
			continue
		}

		typeIcon := ""

		// 3. Determine HTML output based on the MINIMAL set of Item Types
		switch itemType {

		case '0': // Linkable item: Text file (Type 0)
			typeIcon = "[TXT]"
			// Build the link back to the gofer html engine
			link := fmt.Sprintf("<a href=\"/?type=%c&host=%s&port=%s&selector=%s\">%s</a>", itemType, host, port, url.QueryEscape(selector), displayString)
			// future gopher version link := fmt.Sprintf("<a href=\"gopher://%s:%s/%c%s\">%s</a>", host, port, itemType, selector, displayString)
			html.WriteString(fmt.Sprintf("<p class=\"gopher-link\">%s%s</p>\n", typeIcon, link))

		case '1': // Linkable items: Menu (Type 1)
			typeIcon = "[ 1 ]"
			// Build the link back to the gofer html engine
			link := fmt.Sprintf("<a href=\"/?type=%c&host=%s&port=%s&selector=%s\">%s</a>", itemType, host, port, url.QueryEscape(selector), displayString)
			// future gopher version link := fmt.Sprintf("<a href=\"gopher://%s:%s/%c%s\">%s</a>", host, port, itemType, selector, displayString)
			html.WriteString(fmt.Sprintf("<p class=\"gopher-link\">%s%s</p>\n", typeIcon, link))

		case '2': // PH/CSO directory server entry
			typeIcon = "[PhC]"

			// Host/port from the Gopher line
			phHost := host
			phPort := port
			if phPort == "" {
				phPort = "105" // PH default
			}

			// Build base PH URL: /ph:host:port, also create the return link
			returnTo := fmt.Sprintf("/?host=%s&port=%s&selector=%s",
				currentHost,
				currentPort,
				url.QueryEscape(currentSelector),
			)

			phURL := fmt.Sprintf("/ph/%s:%s?return=%s",
				phHost,
				phPort,
				url.QueryEscape(returnTo),
			)

			// Only attach selector parameter if the gopher entry actually had one
			if selector != "" {
				phURL = fmt.Sprintf("%s?selector=%s",
					phURL,
					url.QueryEscape(selector),
				)
			}

			link := fmt.Sprintf("<a href=\"%s\">%s</a>", phURL, displayString)
			html.WriteString(fmt.Sprintf("<p class=\"gopher-link\">%s%s</p>\n", typeIcon, link))
			continue

		case '3': // Error (transparent)
			typeIcon = "[ERR]"
			html.WriteString(fmt.Sprintf("<p class=\"gopher-link\"><span style=\"color: red;\">%s</span>%s</p>\n", typeIcon, displayString))

		case '4': // Macintosh BinHex File (opaque)
			writeLink("[HQX]", itemType, host, port, selector, displayString)

		case '5': // MS DOS Binary File (opaque)
			writeLink("[DOS]", itemType, host, port, selector, displayString)

		case '6': // Unix UUEncoded File (opaque)
			writeLink("[UUE]", itemType, host, port, selector, displayString)

		case '7': // Searchable Index (Type 7)
			typeIcon = "[ 7 ]"

			// Route to search handler (to be implemented)
			link := fmt.Sprintf(
				"<a href=\"/search?host=%s&port=%s&selector=%s\">%s</a>",
				host,
				port,
				url.QueryEscape(selector),
				displayString,
			)

			html.WriteString(fmt.Sprintf(
				"<p class=\"gopher-link\">%s%s</p>\n",
				typeIcon,
				link,
			))

		case 'g': // GIF image (opaque)
			writeLink("[GIF]", itemType, host, port, selector, displayString)

		case 'I': // Generic image (opaque)
			writeLink("[IMG]", itemType, host, port, selector, displayString)

		case 'i': // Informational text (transparent)
			typeIcon = "[ i ]"
			html.WriteString(fmt.Sprintf("<p class=\"gopher-link\"><span style=\"color: gray;\">%s</span>%s</p>\n", typeIcon, displayString))

		default: // Unknown type: treated as opaque.
			typeIcon = fmt.Sprintf("[!%c!]", itemType)
			link := fmt.Sprintf("<a href=\"/?type=%c&host=%s&port=%s&selector=%s\">%s</a>", itemType, host, port, url.QueryEscape(selector), displayString)
			html.WriteString(fmt.Sprintf("<p class=\"gopher-link\"><span style=\"color: red;\">%s</span>%s</p>\n", typeIcon, link))
		}
	}

	// js heartbeat for the inactivity monitor
	html.WriteString(fmt.Sprintf(`
		<script>
		  setInterval(function() {
		    fetch('http://localhost:%s/heartbeat')
		    .catch(error => {
		      console.log('Error - gofer has closed unexpectedly');
		    });
		  }, 55000);
		</script>
	`, LOCAL_SERVER_PORT))

	if !embedded {
		html.WriteString(`</body></html>`)
	}
	return html.String()
}

// --- HTTP Server Handlers ---

// handleHeartbeat updates the activity timer without loading content.
func handleHeartbeat(w http.ResponseWriter, r *http.Request) {
	updateActivity()
	w.WriteHeader(http.StatusOK)
	// No body needed. A successful status code is enough to reset the timer.
}

// serveGopher handles the primary Gopher requests (e.g., /?host=... or just /).
func serveGopher(w http.ResponseWriter, r *http.Request) {
	updateActivity() // Reset the inactivity timer

	var (
		err         error
		u           *url.URL
		rawResponse string
		rawBytes    []byte
	)

	query := r.URL.Query()

	// 1. Check for submission from the new single-field URI bar
	gopherURI := query.Get("uri")

	// 2. Initialize variables (or use values from old menu link clicks)
	gopherTypeQuery := query.Get("type")
	host := query.Get("host")
	port := query.Get("port")
	selector := query.Get("selector")

	if gopherURI != "" {

		raw := gopherURI
		if !strings.Contains(raw, "://") {
			raw = "gopher://" + raw
		}

		u, err = url.Parse(raw)

		if err == nil && (u.Scheme == "gopher" || u.Scheme == "") {
			// Overwrite host, port, and selector from the parsed URI

			host = u.Hostname()

			port = u.Port()
			if port == "" {
				port = DEFAULT_GOPHER_PORT
			}

			selector = strings.TrimPrefix(u.Path, "/")
		}
	} else {
		// Only apply defaults if navigating via an old-style link or direct "/" load
		if host == "" {
			host = DEFAULT_GOPHER_HOST
		}
		if port == "" {
			port = DEFAULT_GOPHER_PORT
		}
		if selector == "" {
			selector = "/"
		}
	}

	rawResponse, err = gopherRequest(host, port, selector)

	if err != nil {
		// Connection Error - a synthetic type-3 line for the formatter
		synthetic := fmt.Sprintf("3Connection failed: %s\t/\t%s\t%s\n.\n",
			err.Error(), host, port)

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		htmlContent := formatMenuHTML(synthetic, host, port, selector, false)
		w.Write([]byte(htmlContent))
		return
	}

	// Determine the Gopher type requested.
	var gopherType byte
	if len(gopherTypeQuery) > 0 {
		gopherType = gopherTypeQuery[0]
	} else {
		// Fallback for direct browser entry or older links (assume menu)
		gopherType = '1'
	}

	isTransparent := isTransparentType(gopherType)

	// Fetch content based on pipeline
	if isTransparent {
		rawResponse, err = gopherRequest(host, port, selector)
	} else {
		rawBytes, err = gopherRequestBytes(host, port, selector)
	}

	if err != nil {
		synthetic := fmt.Sprintf("3Connection failed: %s\t/\t%s\t%s\n.\n",
			err.Error(), host, port)

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		htmlContent := formatMenuHTML(synthetic, host, port, selector, false)
		w.Write([]byte(htmlContent))
		return
	}

	// Handle content based on Gopher Type
	switch gopherType {

	case '0', 'i': // Text File (Type 0) or Informational Text (Type i)
		// Type 0 is sent to the browser as raw text with the correct HTTP header.
		// Type 'i' is only used in a menu and should not be requested directly, but treat it as text/plain if it is.
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Write([]byte(rawResponse))

	case '1': // Menu (Type 1)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		htmlContent := formatMenuHTML(rawResponse, host, port, selector, false)
		w.Write([]byte(htmlContent))

	default:
		// Unknown types are treated as opaque bytes.
		// We provide no strong opinion; the browser decides.
		w.Header().Set("Content-Type", http.DetectContentType(rawBytes))
		w.Write(rawBytes)
	}
}

// handleFocus is called by a newly launched 'gofer' process (PID 2) to signal
// the running process (PID 1) to load a new gopher URI and refresh the browser.
func handleFocus(w http.ResponseWriter, r *http.Request) {
	updateActivity() // Reset the inactivity timer

	// 1. Get the gopher URI passed from the second instance
	gopherURI := r.URL.Query().Get("uri")

	if gopherURI == "" {
		http.Error(w, "Missing 'uri' parameter.", http.StatusBadRequest)
		return
	}

	// 2. Convert the gopher URI into the local HTTP link
	u, err := url.Parse(gopherURI)
	if err != nil || u.Scheme != "gopher" {
		http.Error(w, "Invalid gopher URI.", http.StatusBadRequest)
		return
	}

	// Reconstruct the URL for our local server
	// Example: gopher://freeshell.org:70/1/users becomes /?host=freeshell.org&port=70&selector=1/users

	// u.Path contains the item type and selector (e.g., /1/users)
	// u.Host contains host:port (e.g., freeshell.org:70)

	// We need to pass the raw path, without the leading slash for the selector
	// But since our serveGopher handler handles the parsing of the selector from the path correctly,
	// we just need to reconstruct the full local URL.

	// Use our existing serveGopher logic (which uses the query params)
	// We must separate host and port from u.Host

	host := u.Hostname()
	port := u.Port()
	if port == "" {
		port = DEFAULT_GOPHER_PORT
	}

	// The selector is the path without the leading slash
	selector := strings.TrimPrefix(u.Path, "/")

	// Construct the local URL to load
	localURL := fmt.Sprintf("http://localhost:%s/?host=%s&port=%s&selector=%s", LOCAL_SERVER_PORT, host, port, selector)

	// 3. Launch the browser to the new URL
	// The browser will typically focus on the existing tab or open a new one.
	launchBrowser(localURL)

	// 4. Respond to the second instance
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Redirecting session to: %s", localURL)
}

// handlePHEntry catches requests for Type 2 cso-ph directory requests
func handlePHEntry(w http.ResponseWriter, r *http.Request) {
	updateActivity()
	HandlePH(w, r)
}

// --- Main Function ---

func main() {

	// --- STEP 1: Parse Command-Line Arguments (Gopher URI) ---

	// Determine the initial Gopher URL to load.
	// This will be used in the first instance (PID 1) to open the browser.
	initialGopherURL := fmt.Sprintf("http://localhost:%s/?host=%s&port=%s&selector=/", LOCAL_SERVER_PORT, DEFAULT_GOPHER_HOST, DEFAULT_GOPHER_PORT)

	// If a command-line argument is passed (likely a gopher:// URI from the OS handler)
	if len(os.Args) > 1 {
		// The argument is the gopher URI
		gopherURI := os.Args[1]
		// Convert it to our local HTTP URL for the browser
		// We use the Focus endpoint logic to convert the gopher URI to local URL
		u, err := url.Parse(gopherURI)
		if err == nil && u.Scheme == "gopher" {
			host := u.Hostname()
			port := u.Port()
			if port == "" {
				port = DEFAULT_GOPHER_PORT
			}
			selector := strings.TrimPrefix(u.Path, "/")
			initialGopherURL = fmt.Sprintf("http://localhost:%s/?host=%s&port=%s&selector=%s", LOCAL_SERVER_PORT, host, port, selector)
		} else {
			fmt.Printf("Warning: Invalid URI received: %s. Loading default page.\n", gopherURI)
		}
	}

	// --- STEP 2: Singleton Check (Attempt to bind to the port) ---

	listener, err := net.Listen("tcp", ":"+LOCAL_SERVER_PORT)
	if err != nil {
		// Port is already in use (PID 1 is running) -> This is PID 2
		fmt.Printf("gofer (PID %d) is already running on port %s. Sending Re-Focus signal.\n", os.Getpid(), LOCAL_SERVER_PORT)

		// Send a request to PID 1 to handle the new Gopher URI
		// If we were launched with a Gopher URL, forward it. Otherwise request a generic focus.
		var targetURL string
		if len(os.Args) > 1 {
			targetURL = fmt.Sprintf(
				"http://localhost:%s%s?uri=%s",
				LOCAL_SERVER_PORT,
				FOCUS_ENDPOINT,
				url.QueryEscape(os.Args[1]),
			)
		} else {
			// No arg provided; request focus without a URI.
			targetURL = fmt.Sprintf(
				"http://localhost:%s%s",
				LOCAL_SERVER_PORT,
				FOCUS_ENDPOINT,
			)
		}

		resp, err := http.Get(targetURL)
		if err != nil {
			fmt.Printf("Error sending re-focus signal: %v\n", err)
			os.Exit(1)
		}
		defer resp.Body.Close()

		// PID 2 exits gracefully after sending the signal.
		os.Exit(0)
	}
	defer listener.Close()

	// --- STEP 3: Primary Instance (PID 1) Initialization ---

	fmt.Printf("gofer (PID %d) starting server on port %s...\n", os.Getpid(), LOCAL_SERVER_PORT)

	// 1. Start the inactivity monitor in a separate goroutine
	go monitorInactivity()

	// 2. Set up the HTTP handlers
	http.HandleFunc("/", serveGopher)
	http.HandleFunc(FOCUS_ENDPOINT, handleFocus)   // handler for PID 2 signals
	http.HandleFunc("/heartbeat", handleHeartbeat) // handler for keep-alive ping
	http.HandleFunc("/ph/", handlePHEntry)         // handler for type 2 ph_client and cso directorys
	http.HandleFunc("/search", HandleSearch)       // handler for type 7 searches

	// 3. Launch the browser to the initial URL (parsed from CLI or default)
	launchBrowser(initialGopherURL)

	// 4. Start the server using the listener we successfully created
	// This blocks the main goroutine until termination (by the monitor or Ctrl+C)
	server := &http.Server{Handler: nil}
	err = server.Serve(listener)

	if err != nil && err != http.ErrServerClosed {
		fmt.Printf("Error serving HTTP: %v\n", err)
		os.Exit(1)
	}
}
