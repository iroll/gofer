// heartmon 0.9
// the keep-alive system for gofer
// (C) 2025 Isaac Roll
// See github.com/iroll/gofer for license

// /heartbeat is a machine liveness endpoint.
// /heartmon is the human-visible lifecycle handle that emits heartbeats.

package main

import (
	"fmt"
	"net/http"
)

// handleHeartbeat updates the activity timer without loading content.
func handleHeartbeat(w http.ResponseWriter, r *http.Request) {
	updateActivity()
	w.WriteHeader(http.StatusOK)
	// No body needed. A successful status code is enough to reset the timer.
}

// serveHeartMon is a little window where the heartbeat lives, closeable by user
func serveHeartMon(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	fmt.Fprintf(w, `
<!DOCTYPE html>
<html>
<head>
	<title>gofer — running</title>
	<style>
		body {
			font-family: monospace;
			text-align: center;
			margin-top: 2em;
		}
		button {
			margin-top: 1em;
			font-family: monospace;
			cursor: pointer;
		}
	</style>
</head>
<body>

	<p>close this tab or window to exit gofer</p>

	<button onclick="popout()">pop out</button>

	<script>
		function ping() {
			fetch('http://localhost:%s/heartbeat')
				.catch(() => {
					window.close();
				});
		}

		function popout() {
			const w = window.open(
				"/heartmon",
				"gofer-heartmon",
				"width=240,height=240,resizable=yes"
			);

			// If popup succeeded, close this tab
			if (w) {
				window.close();
			}
		}

		ping();
		setInterval(ping, 30000);
	</script>

</body>
</html>
`, LOCAL_SERVER_PORT)
}
