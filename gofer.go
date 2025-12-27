// gofer 0.9
// a gopher helper for web browsers
// hewing as close to RFC 1436 (1993) as practical
// (C) 2025 Isaac Roll
// See github.com/iroll/gofer for license

package main

import (
	_ "embed"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

// --- Configuration Constants ---
const (
	LOCAL_SERVER_PORT         = "8000"
	DEFAULT_GOPHER_HOST       = "freeshell.org"
	DEFAULT_GOPHER_PORT       = "70"
	TCP_TIMEOUT               = 5 * time.Second
	GOPHER_REQUEST_TERMINATOR = "\r\n"
	FOCUS_ENDPOINT            = "/focus"
)

type menuRenderRule struct {
	icon  string
	color string
}

var menuRules = map[byte]menuRenderRule{
	'0': {icon: "[TXT]"},
	'1': {icon: "[ 1 ]"},
	'2': {icon: "[PhC]"},
	'3': {icon: "[ERR]", color: "red"},
	'4': {icon: "[HQX]"},
	'5': {icon: "[DOS]"},
	'6': {icon: "[UUE]"},
	'7': {icon: "[ 7 ]"},
	'g': {icon: "[GIF]"},
	'I': {icon: "[IMG]"},
	'i': {icon: "[ i ]", color: "gray"},
}

// --- launchBrowser ---
// opens the default web browser to the given URL

func launchBrowser(targetURL string) {
	// make sure that gopher:// links are translated into internal html
	if strings.HasPrefix(targetURL, "gopher://") {
		targetURL = fmt.Sprintf("http://localhost:%s/?uri=%s",
			LOCAL_SERVER_PORT, url.QueryEscape(targetURL))
	}

	var cmd *exec.Cmd

	// Use runtime.GOOS to get the OS the program is running on
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", targetURL)
	case "darwin": // macOS
		cmd = exec.Command("open", targetURL)
	default: // Linux (and others)
		cmd = exec.Command("xdg-open", targetURL)
	}

	// use Start() to avoid blocking the main goroutine
	err := cmd.Start()
	if err != nil {
		fmt.Printf("Warning: Could not launch browser: %v\n", err)
	}
}

// --- serveLanding ---
// provides the default home page

//go:embed ui/landingGopher
var landingGopher []byte

func serveLanding(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	html := formatMenuHTML(
		string(landingGopher),
		"local",
		"",
		"/",
		false,
	)
	w.Write([]byte(html))
}

// --- formatMenuHTML ---
// formatMenuHTML takes raw Gopher data and turns it into minimal HTML.
// It requires the current host, port, and selector for form pre-filling and links.

func formatMenuHTML(rawGopherData, currentHost, currentPort, currentSelector string, embedded bool) string {

	// Start with the HTML boilerplate, including the input form at the top
	var html strings.Builder

	// 1a. unified link builder
	buildInternalLink := func(t byte, host, port, selector, display string) string {
		u := url.URL{Path: "/"}
		q := u.Query()
		q.Set("type", string([]byte{t}))
		q.Set("host", host)
		q.Set("port", port)
		q.Set("selector", selector)
		u.RawQuery = q.Encode()
		return fmt.Sprintf("<a href=\"%s\">%s</a>", u.String(), display)
	}

	// 1b. unified link writer
	writeLink := func(t byte, icon string, host, port, selector, display string) {
		link := buildInternalLink(t, host, port, selector, display)
		html.WriteString(fmt.Sprintf("<p class=\"gopher-link\">%s%s</p>\n", icon, link))
	}

	// 2. Construct the current Gopher URI for the input field's value
	var currentGopherURI string
	if currentSelector == "/" {
		currentGopherURI = fmt.Sprintf("%s:%s/", currentHost, currentPort)
	} else {
		// currentSelector already has leading slash from u.Path
		currentGopherURI = fmt.Sprintf("%s:%s%s", currentHost, currentPort, currentSelector)
	}

	// 3. The HTML/CSS framework
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
        			display: flex;
        			width: 100%%;
        			align-items: center;
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

	// 4. Process the lines from the Gopher response
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

		// If a malformed line (less than four fields) is detected, assume itemType 3

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
			selector = strings.TrimSpace(fields[1])
			host = strings.TrimSpace(fields[2])
			port = strings.TrimSpace(fields[3])
		}

		displayString = strings.TrimRight(displayString, " \t\r")

		if displayString == "" {
			continue
		}

		// External HTTP(S) link embedded in selector — hand off to browser verbatim
		lower := strings.ToLower(selector)

		idx := strings.Index(lower, "http://")
		if idx == -1 {
			idx = strings.Index(lower, "https://")
		}
		if idx != -1 {
			ext := selector[idx:] // slice ORIGINAL string

			html.WriteString(fmt.Sprintf(
				"<p class=\"gopher-link\"><span style=\"color: red;\">[ ! ]</span><a href=\"%s\">%s</a></p>\n",
				ext,
				displayString,
			))
			continue
		}

		// 3. Determine HTML output based on the itemType
		switch itemType {

		case '0': // Linkable item: Text file (text pipeline)
			rule := menuRules[itemType]
			writeLink(itemType, rule.icon, host, port, selector, displayString)

		case '1': // Linkable items: Menu (text pipeline)
			rule := menuRules[itemType]
			writeLink(itemType, rule.icon, host, port, selector, displayString)

		case '2': // PH/CSO directory server entry

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
			html.WriteString(fmt.Sprintf(
				"<p class=\"gopher-link\">[PhC]%s</p>\n",
				link,
			))
			continue

		case '3': // Error (text pipeline)
			rule := menuRules[itemType]
			html.WriteString(fmt.Sprintf(
				"<p class=\"gopher-link\"><span style=\"color: %s;\">%s</span>%s</p>\n",
				rule.color,
				rule.icon,
				displayString,
			))

		case '4', '5', '6', 'g', 'I': // BinHex, DosBin, UUE, GIF, Generic Image (byte pipeline)
			rule := menuRules[itemType]
			writeLink(itemType, rule.icon, host, port, selector, displayString)

		case '7': // Searchable Index (Type 7)

			// Route to search handler (to be implemented)
			link := fmt.Sprintf(
				"<a href=\"/search?host=%s&port=%s&selector=%s\">%s</a>",
				host,
				port,
				url.QueryEscape(selector),
				displayString,
			)

			html.WriteString(fmt.Sprintf(
				"<p class=\"gopher-link\">[ 7 ]%s</p>\n",
				link,
			))

		case 'i': // Informational text (text pipeline)
			rule := menuRules[itemType]
			html.WriteString(fmt.Sprintf(
				"<p class=\"gopher-link\"><span style=\"color: %s;\">%s</span>%s</p>\n",
				rule.color,
				rule.icon,
				displayString,
			))

		default: // Unknown type: treated as byte pipeline.
			link := buildInternalLink(itemType, host, port, selector, displayString)
			html.WriteString(fmt.Sprintf(
				"<p class=\"gopher-link\"><span style=\"color: red;\">[!%c!]</span>%s</p>\n",
				itemType,
				link,
			))

		}
	}

	if !embedded {
		html.WriteString(`</body></html>`)
	}
	return html.String()
}

// --- serveGopher: the HTTP Server Handlers ---
// serveGopher handles the primary Gopher requests (e.g., /?host=... or just /).
// NOTE: itemType is authoritative. Do not infer from selector or content.
func serveGopher(w http.ResponseWriter, r *http.Request) {

	var (
		err        error
		engineResp GopherResponse
	)

	query := r.URL.Query()

	// 1. Check for normalized gopher input
	gopherURI := query.Get("uri")

	var (
		host     string
		port     string
		selector string
		gType    byte
	)

	// URI path takes precedence
	if gopherURI != "" {
		target, err := normalizeGopherInput(gopherURI)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		host = target.Host
		port = target.Port
		selector = target.Selector
		gType = target.Type

	} else {
		// Legacy navigation via internal links
		host = query.Get("host")
		port = query.Get("port")
		selector = query.Get("selector")

		if host == "" {
			host = DEFAULT_GOPHER_HOST
		}
		if port == "" {
			port = DEFAULT_GOPHER_PORT
		}
		if selector == "" {
			selector = "/"
		}

		// Fallback
		typeParam := query.Get("type")
		if typeParam != "" {
			gType = typeParam[0]
		} else if selector == "/" {
			// root menu has no originating item type
			gType = '1'
		} else {
			http.Error(w, "missing gopher item type", http.StatusBadRequest)
			return
		}

	}

	// Fetch content based on pipeline
	engineResp, err = Fetch(GopherTarget{
		Host:     host,
		Port:     port,
		Selector: selector,
		Type:     gType,
	})

	if err != nil {
		synthetic := fmt.Sprintf(
			"3Connection failed: %s\t/\t%s\t%s\n.\n",
			err.Error(),
			host,
			port,
		)

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		htmlContent := formatMenuHTML(synthetic, host, port, selector, false)
		w.Write([]byte(htmlContent))
		return
	}

	// Render based on engine response

	// opaque pipeline if bytes are present
	if len(engineResp.RawBytes) > 0 {
		w.Header().Set("Content-Type", http.DetectContentType(engineResp.RawBytes))
		w.Write(engineResp.RawBytes)
		return
	}

	// text pipeline content
	// menu vs text is determined by type
	if gType == '1' {
		// menu
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		htmlContent := formatMenuHTML(engineResp.RawText, host, port, selector, false)
		w.Write([]byte(htmlContent))
		return
	}

	// text file
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte(engineResp.RawText))

}

// --- handleFocus ---
// called by a newly launched 'gofer' process (PID 2) to signal
// the running process (PID 1) to load a new gopher URI and refresh the browser.

func handleFocus(w http.ResponseWriter, r *http.Request) {

	// 1. Get the gopher URI passed from the second instance
	gopherURI := r.URL.Query().Get("uri")

	if gopherURI == "" {
		http.Error(w, "Missing 'uri' parameter.", http.StatusBadRequest)
		return
	}

	// 2. Convert the gopher URI into the local HTTP link
	target, err := normalizeGopherInput(gopherURI)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	localURL := fmt.Sprintf(
		"http://localhost:%s/?host=%s&port=%s&selector=%s",
		LOCAL_SERVER_PORT,
		target.Host,
		target.Port,
		url.QueryEscape(target.Selector),
	)

	// 3. Launch the browser to the new URL
	// The browser will typically focus on the existing tab or open a new one.
	launchBrowser(localURL)

	// 4. Respond to the second instance
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Redirecting session to: %s", localURL)
}

// handlePHEntry catches requests for Type 2 cso-ph directory requests
func handlePHEntry(w http.ResponseWriter, r *http.Request) {
	HandlePH(w, r)
}

// --- Main Function ---

func main() {

	// --- STEP 1: Parse Command-Line Arguments (Gopher URI) ---

	// Default initial URL (used ONLY by the primary instance)
	initialGopherURL := fmt.Sprintf(
		"http://localhost:%s/landing",
		LOCAL_SERVER_PORT,
	)

	// Scan args for a gopher-ish argument (Safari may insert junk args)
	var rawInput string
	for _, arg := range os.Args[1:] {
		if strings.Contains(arg, "gopher://") || !strings.Contains(arg, "://") {
			rawInput = arg
			break
		}
	}

	if rawInput != "" {
		target, err := normalizeGopherInput(rawInput)
		if err == nil {
			initialGopherURL = fmt.Sprintf(
				"http://localhost:%s/?host=%s&port=%s&selector=%s",
				LOCAL_SERVER_PORT,
				target.Host,
				target.Port,
				url.QueryEscape(target.Selector),
			)
		}
	}

	// --- STEP 2: Singleton Check (Attempt to bind to the port) ---

	listener, err := net.Listen("tcp", ":"+LOCAL_SERVER_PORT)
	if err != nil {
		// --- PID 2 (secondary instance) ---

		// If browser launched us WITHOUT a gopher URI, do nothing.
		// This prevents clobbering the running session.
		if rawInput == "" {
			os.Exit(0)
		}

		// Forward the URI to the primary instance
		targetURL := fmt.Sprintf(
			"http://localhost:%s%s?uri=%s",
			LOCAL_SERVER_PORT,
			FOCUS_ENDPOINT,
			url.QueryEscape(rawInput),
		)

		_, _ = http.Get(targetURL)
		os.Exit(0)
	}

	defer listener.Close()

	// --- STEP 3: Primary Instance (PID 1) Initialization ---

	fmt.Printf("gofer (PID %d) starting server on port %s...\n", os.Getpid(), LOCAL_SERVER_PORT)

	// 1. Start the inactivity monitor in a separate goroutine
	// DEPRECATED

	// 2. Set up the HTTP handlers
	http.HandleFunc("/", serveGopher)
	http.HandleFunc(FOCUS_ENDPOINT, handleFocus) // handler for PID 2 signals
	http.HandleFunc("/ph/", handlePHEntry)       // handler for type 2 ph_client and cso directorys
	http.HandleFunc("/search", HandleSearch)     // handler for type 7 searches
	http.HandleFunc("/landing", serveLanding)    // default internal home page

	// 3. Launch the browser to the initial URL (parsed from CLI or default)
	browserURL := initialGopherURL

	// CRITICAL: If the input is a gopher protocol, we must translate it
	// to our local HTTP proxy link so the OS doesn't call gofer again.
	if strings.HasPrefix(initialGopherURL, "gopher://") {
		browserURL = fmt.Sprintf("http://localhost:%s/?uri=%s",
			LOCAL_SERVER_PORT, url.QueryEscape(initialGopherURL))
	}

	// 4. Launch the browser to our PROXY link, not the raw gopher link
	launchBrowser(browserURL)

	// 4. Start the server using the listener we successfully created
	// This blocks the main goroutine until termination (by the monitor or Ctrl+C)
	server := &http.Server{Handler: nil}
	err = server.Serve(listener)

	if err != nil && err != http.ErrServerClosed {
		fmt.Printf("Error serving HTTP: %v\n", err)
		os.Exit(1)
	}
}
