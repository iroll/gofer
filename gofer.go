package main

import (
	"bufio"
	"fmt"
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

// --- Inactivity Monitor Logic ---

var lastRequestTime = time.Now()
var shutdownMux sync.Mutex

// updateActivity resets the inactivity timer. Called by all HTTP handlers.
func updateActivity() {
	shutdownMux.Lock()
	lastRequestTime = time.Now()
	shutdownMux.Unlock()
}

// monitorInactivity checks the time since the last request and shuts down if timed out.
func monitorInactivity() {
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

// gopherRequest connects to a remote Gopher server, sends the selector, and returns the raw response.
func gopherRequest(host string, port string, selector string) (string, error) {
	address := net.JoinHostPort(host, port)

	conn, err := net.DialTimeout("tcp", address, TCP_TIMEOUT)
	if err != nil {
		return "", fmt.Errorf("failed to connect to Gopher server %s: %w", address, err)
	}
	defer conn.Close()

	conn.SetDeadline(time.Now().Add(TCP_TIMEOUT))

	request := selector + GOPHER_REQUEST_TERMINATOR

	_, err = conn.Write([]byte(request))
	if err != nil {
		return "", fmt.Errorf("failed to write selector to socket: %w", err)
	}

	// Read the entire response
	reader := bufio.NewReader(conn)
	var responseBuilder strings.Builder

	// Read until EOF or timeout
	for {
		line, err := reader.ReadString('\n')
		if len(line) > 0 {
			responseBuilder.WriteString(line)
		}

		// --- Error Handling Block ---
		if err != nil {
			// 1. Check for EOF (normal termination for Gopher protocol)
			if err.Error() == "EOF" {
				break
			}

			// 2. Check for net.Error Timeout
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				break // Treat timeout as a successful connection termination
			}

			// 3. Any other error is a genuine failure
			return "", fmt.Errorf("error reading from socket: %w", err)
		}
	}

	return responseBuilder.String(), nil
}

// --- HTML Formatting Component ---

// parseAndFormat takes raw Gopher data and turns it into minimal HTML.
// It requires the current host, port, and selector for form pre-filling and links.
func parseAndFormat(rawGopherData, currentHost, currentPort, currentSelector string) string {

	// Start with the HTML boilerplate, including the input form at the top
	var html strings.Builder

	// Inject current values into the form for persistence and debugging
	formHostValue := fmt.Sprintf(`value="%s"`, currentHost)
	formPortValue := fmt.Sprintf(`value="%s"`, currentPort)
	formSelectorValue := fmt.Sprintf(`value="%s"`, currentSelector)

	html.WriteString(fmt.Sprintf(`
		<!DOCTYPE html>
		<html>
		<head>
			<title>gofer - %s:%s%s</title>
			<style>
				<!--
				body { font-family: monospace; max-width: 800px; margin: 0 auto; padding: 20px; line-height: 1.4; }
				/* pre { white-space: pre-wrap; word-break: break-word; font-family: monospace; } */
				.gopher-link { display: block; margin: 4px 0; }
				/* Reduced width of gopher-type for better flow */
				.gopher-type { font-weight: bold; margin-right: 8px; color: #666; width: 25px; display: inline-block; }
				#input-form { margin-bottom: 20px; padding: 15px; border: 1px solid #ccc; background-color: #f9f9f9; }
				#input-form input { margin-right: 10px; padding: 5px; border: 1px solid #ddd; }
				-->
			</style>
		</head>
		<body>
		
		<div id="input-form">
			<form action="/" method="GET">
				<label for="host">Hostname:</label>
				<input type="text" id="host" name="host" placeholder="freeshell.org" %s>
				<label for="port">Port:</label>
				<input type="number" id="port" name="port" placeholder="70" %s style="width: 50px;">
				<label for="selector">Selector:</label>
				<input type="text" id="selector" name="selector" placeholder="/" %s style="width: 250px;">
				<button type="submit">Go!</button>
			</form>
		</div>

		<h1>gopher://%s:%s%s</h1>

	`,
		// --- Start of the argument list ---
		// Arguments 1, 2, 3: For the <title> tag
		currentHost, currentPort, currentSelector,

		// Arguments 4, 5, 6: For the input value attributes (formHostValue, etc.)
		formHostValue, formPortValue, formSelectorValue,

		// Arguments 10, 11, 12: For the new <h1> line
		currentHost, currentPort, currentSelector))

	// --- End of the argument list ---

	// Process the lines from the Gopher response
	lines := strings.Split(rawGopherData, "\n")

	for _, line := range lines {
		// A Gopher menu ends with a single '.' on a line, but typically the connection closes.
		trimmedline := strings.TrimSpace(line)
		if trimmedline == "" || trimmedline == "." {
			continue
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

		displayString = strings.TrimSpace(displayString)

		if displayString == "" {
			continue
		}

		var typeIcon string

		// 3. Determine HTML output based on the MINIMAL set of Item Types
		switch itemType {
		case '0', '1': // Linkable items: Text file (0) or Menu (1)
			typeIcon = fmt.Sprintf("[%c]", itemType)

			// Build the link back to the gofer html engine
			link := fmt.Sprintf("<a href=\"/?host=%s&port=%s&selector=%s\">%s</a>", host, port, selector, displayString)
			// future gopher version link := fmt.Sprintf("<a href=\"gopher://%s:%s/%c%s\">%s</a>", host, port, itemType, selector, displayString)
			html.WriteString(fmt.Sprintf("<div class=\"gopher-link\"><span class=\"gopher-type\">%s</span> %s</div>\n", typeIcon, link))

		case '3': // Error
			typeIcon = "[ERR]"
			html.WriteString(fmt.Sprintf("<div class=\"gopher-link\"><span class=\"gopher-type\" style=\"color: red;\">%s</span> %s</div>\n", typeIcon, displayString))

		case 'i': // Informational text
			typeIcon = "[INF]"
			html.WriteString(fmt.Sprintf("<div class=\"gopher-link\"><span class=\"gopher-type\" style=\"color: gray;\">%s</span> %s</div>\n", typeIcon, displayString))

		default: // All other types (4, 5, 7, 9, I, g, T, etc.) are treated as informational text
			typeIcon = "[?]"
			html.WriteString(fmt.Sprintf("<div class=\"gopher-link\"><span class=\"gopher-type\" style=\"color: gray;\">%s</span> %s</div>\n", typeIcon, displayString))
		}
	}

	// Add a minimal JS heartbeat to keep the server alive while the page is open.
	// SHUTDOWN_TIMEOUT_SECONDS is 60s, so 55s ensures a successful ping.
	// If the ping fails, the server is dead, so the script stops.
	// This maintains the single-tab UX and allows the server to shut down when the user is truly idle.
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

	html.WriteString(`</body></html>`)
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

	query := r.URL.Query()
	host := query.Get("host")
	port := query.Get("port")
	selector := query.Get("selector")

	// Set defaults if missing
	if host == "" {
		host = DEFAULT_GOPHER_HOST
	}
	if port == "" {
		port = DEFAULT_GOPHER_PORT
	}
	if selector == "" {
		selector = "/"
	}

	rawResponse, err := gopherRequest(host, port, selector)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	if err != nil {
		http.Error(w, fmt.Sprintf("<h1>Connection Error</h1><p>Failed to retrieve Gopher resource from %s:%s%s. Details: %s</p>", host, port, selector, err.Error()), http.StatusInternalServerError)
		return
	}

	// Now passing all current params to the formatter
	htmlContent := parseAndFormat(rawResponse, host, port, selector)
	w.Write([]byte(htmlContent))
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
		targetURL := fmt.Sprintf("http://localhost:%s%s?uri=%s", LOCAL_SERVER_PORT, FOCUS_ENDPOINT, url.QueryEscape(os.Args[1]))

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
