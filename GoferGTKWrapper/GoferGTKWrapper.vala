using Gtk;
using Adw;

public class GopherApp : Adw.Application {
    private Adw.ApplicationWindow window;
    private Gtk.Label status_label;
    private Subprocess gofer_process;

    public GopherApp() {
        Object(
            application_id: "com.iroll.gofer",
            flags: ApplicationFlags.HANDLES_OPEN
        );
    }

    protected override void activate() {
        // First launch without URL
        if (window == null) {
            launch_gofer(null);
            create_window();
        }
        window.present();
    }

    protected override void open(File[] files, string hint) {
        // Handle gopher:// URLs
        if (files.length > 0) {
            string uri = files[0].get_uri();
            
            if (window == null) {
                // First launch with URL
                launch_gofer(uri);
                create_window();
            } else {
                // Already running - pass URL to gofer
                forward_url_to_gofer(uri);
            }
            window.present();
        }
    }

    private void launch_gofer(string? gopher_url) {
        try {
            string[] argv;
            
            // Get path to gofer binary (assumes it's in PATH or same directory)
            if (gopher_url != null) {
                argv = {"gofer", gopher_url};
            } else {
                argv = {"gofer"};
            }

            gofer_process = new Subprocess.newv(
                argv,
                SubprocessFlags.STDOUT_PIPE | SubprocessFlags.STDERR_PIPE
            );

            // Monitor gofer process - quit wrapper when it dies
            gofer_process.wait_async.begin(null, (obj, res) => {
                try {
                    gofer_process.wait_async.end(res);
                } catch (Error e) {
                    warning("Error waiting for gofer: %s", e.message);
                }
                // Gofer died, quit the wrapper
                quit();
            });

        } catch (Error e) {
            var dialog = new Adw.AlertDialog(
                "Failed to launch gofer",
                e.message
            );
            dialog.add_response("ok", "OK");
            if (window != null) {
                dialog.present(window);
            } else {
                dialog.present(null);
            }
        }
    }

    private void forward_url_to_gofer(string uri) {
        // Pass URL to already-running gofer via its API using GIO
        string escaped_uri = Uri.escape_string(uri, null, false);
        string url = @"http://localhost:8000/focus?uri=$(escaped_uri)";
        
        try {
            var file = File.new_for_uri(url);
            // Fire and forget - just trigger the request
            file.load_contents_async.begin(null, (obj, res) => {
                try {
                    uint8[] contents;
                    string etag_out;
                    file.load_contents_async.end(res, out contents, out etag_out);
                } catch (Error e) {
                    // Ignore errors - gofer might not be running yet
                }
            });
        } catch (Error e) {
            warning("Failed to forward URL to gofer: %s", e.message);
        }
    }

    private void create_window() {
        window = new Adw.ApplicationWindow(this);
        window.set_title("gofer");
        window.set_default_size(300, 120);
        window.set_resizable(false);

        // Create a ToolbarView to hold the HeaderBar and content
        var toolbar_view = new Adw.ToolbarView();
        
        var header = new Adw.HeaderBar();
        header.set_show_title(true);
        toolbar_view.add_top_bar(header);

        // Main container
        var box = new Gtk.Box(Gtk.Orientation.VERTICAL, 0);
        box.set_margin_top(20);
        box.set_margin_bottom(20);
        box.set_margin_start(20);
        box.set_margin_end(20);

        // Status label
        status_label = new Gtk.Label("gofer is digging");
        status_label.add_css_class("title-2");
        status_label.set_margin_bottom(20);

        // Button box
        var button_box = new Gtk.Box(Gtk.Orientation.HORIZONTAL, 10);
        button_box.set_homogeneous(true);

        // Home button
        var home_button = new Gtk.Button.with_label("Home");
        home_button.clicked.connect(on_home_clicked);

        // Quit button
        var quit_button = new Gtk.Button.with_label("Quit");
        quit_button.clicked.connect(on_quit_clicked);

        button_box.append(home_button);
        button_box.append(quit_button);

        box.append(status_label);
        box.append(button_box);

        toolbar_view.set_content(box);
        window.set_content(toolbar_view);
        window.close_request.connect(on_window_close);
        
        // Check handler AFTER window is fully set up
        // Use idle to ensure window is realized
        Idle.add(() => {
            check_gopher_handler();
            return false;
        });
    }

    private void on_home_clicked() {
        // Open landing page in default browser
        try {
            AppInfo.launch_default_for_uri("http://localhost:8000/landing", null);
        } catch (Error e) {
            warning("Failed to open landing page: %s", e.message);
        }
    }

    private void on_quit_clicked() {
        quit_app();
    }

    private bool on_window_close() {
        quit_app();
        return false;
    }

    private void quit_app() {
        // Kill gofer process if running
        if (gofer_process != null) {
            try {
                gofer_process.force_exit();
            } catch (Error e) {
                warning("Error terminating gofer: %s", e.message);
            }
        }
        quit();
    }

    private void check_gopher_handler() {
        try {
            var proc = new Subprocess(
                SubprocessFlags.STDOUT_PIPE | SubprocessFlags.STDERR_PIPE,
                "xdg-mime", "query", "default", "x-scheme-handler/gopher"
            );

            string stdout_str;
            string stderr_str;
            
            // Wait for the process to complete
            if (!proc.communicate_utf8(null, null, out stdout_str, out stderr_str)) {
                warning("Failed to query gopher handler");
                return;
            }

            var current = stdout_str.strip();
            
            // Check if we're already the default handler
            if (current != "com.iroll.gofer.desktop" && current != "") {
                ask_to_set_default();
            } else if (current == "") {
                // No handler set at all - offer to set ourselves
                ask_to_set_default();
            }

        } catch (Error e) {
            warning("Error checking gopher handler: %s", e.message);
            // Don't block startup on this error
        }
    }

    private void ask_to_set_default() {
    // 1. Initialize without the parent window (we pass that later)
    var dialog = new Adw.AlertDialog(
        "Make gofer the default handler?",
        "Gopher links are currently handled by another application.\n\n" +
        "Would you like gofer to open gopher:// links?"
    );

    // 2. Add responses (buttons)
    // Note: add_response is now used for simple labeled buttons
    dialog.add_response("no", "Not now");
    dialog.add_response("yes", "Make default");

    // 3. Set appearances and defaults
    dialog.set_response_appearance("yes", Adw.ResponseAppearance.SUGGESTED);
    dialog.set_default_response("yes");
    dialog.set_close_response("no");

    // 4. Use the 'choose' method (Asynchronous)
    // 'window' is the parent passed here instead of in the constructor
    dialog.choose.begin(window, null, (obj, res) => {
        string response = dialog.choose.end(res);
        
        if (response == "yes") {
            set_gopher_default();
        }
        // No need for dialog.close(), choose.end handles the cleanup
    });
}

    private void set_gopher_default() {
        try {
            var proc = new Subprocess.newv(
                { "xdg-mime", "default", "com.iroll.gofer.desktop", "x-scheme-handler/gopher" },
                SubprocessFlags.STDOUT_PIPE | SubprocessFlags.STDERR_PIPE
            );
            
            // Wait for it to complete
            proc.wait(null);
            
            if (proc.get_exit_status() == 0) {
                // Success - update status
                status_label.set_text("gofer is digging!");
            } else {
                warning("xdg-mime failed with status: %d", proc.get_exit_status());
            }
            
        } catch (Error e) {
            warning("Failed to set gopher handler: %s", e.message);
        }
    }
}

public static int main(string[] args) {
    var app = new GopherApp();
    return app.run(args);
}
