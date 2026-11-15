using System;
using System.Windows;
using MessageBox = System.Windows.MessageBox;
using NotifyIcon = System.Windows.Forms.NotifyIcon;
using ContextMenuStrip = System.Windows.Forms.ContextMenuStrip;
using ToolStripSeparator = System.Windows.Forms.ToolStripSeparator;
using System.Diagnostics;
using System.IO;
using System.Net.Http;
using System.Threading.Tasks;

namespace GoferWrapper
{
    public partial class App : System.Windows.Application
    {
        private static readonly HttpClient http = new HttpClient();
        private NotifyIcon? _tray;
        private Process? goferProcess;
        private string GoferIconPath => Path.Combine(AppContext.BaseDirectory, "gofer.ico");
        private string GoferBackendPath => Path.Combine(AppContext.BaseDirectory, "gofer-backend.exe");

        private string ResolveGoferBackendPath()
        {
            if (File.Exists(GoferBackendPath))
            {
                return GoferBackendPath;
            }

            string legacyWrapperPath = Path.Combine(AppContext.BaseDirectory, "gofer-wrapper.exe");
            if (File.Exists(legacyWrapperPath))
            {
                return legacyWrapperPath;
            }

            string legacyGoferPath = Path.Combine(AppContext.BaseDirectory, "gofer.exe");
            if (File.Exists(legacyGoferPath))
            {
                return legacyGoferPath;
            }

            return GoferBackendPath;
        }

        protected override void OnStartup(StartupEventArgs e)
        {
            ShutdownMode = ShutdownMode.OnExplicitShutdown;
            base.OnStartup(e);

            try
            {
                _tray = new NotifyIcon
                {
                    Text = "gofer",
                    Icon = File.Exists(GoferIconPath) ? new System.Drawing.Icon(GoferIconPath) : null,
                    Visible = true,
                    ContextMenuStrip = new ContextMenuStrip()
                };

                _tray.ContextMenuStrip.Items.Add("Home", null, OnHome);
                _tray.ContextMenuStrip.Items.Add("About", null, OnAbout);
                _tray.ContextMenuStrip.Items.Add(new ToolStripSeparator());
                _tray.ContextMenuStrip.Items.Add("Quit", null, OnQuit);

                string? gopherUrl = e.Args.Length > 0 ? e.Args[0] : null;

                EnsureGoferRunning(gopherUrl);

            }
            catch (Exception ex)
            {
                MessageBox.Show(
                    ex.ToString(),
                    "gofer-wrapper crash",
                    MessageBoxButton.OK,
                    MessageBoxImage.Error
                );
                Shutdown();
            }

        }
        private void EnsureGoferRunning(string? gopherUrl)
        {
            if (goferProcess == null || goferProcess.HasExited)
            {
                var psi = new ProcessStartInfo
                {
                    FileName = ResolveGoferBackendPath(),
                    UseShellExecute = false,
                    CreateNoWindow = true,
                    WindowStyle = ProcessWindowStyle.Hidden,
                    WorkingDirectory = AppContext.BaseDirectory
                };

                if (!string.IsNullOrEmpty(gopherUrl))
                {
                    psi.ArgumentList.Add(gopherUrl);
                }

                goferProcess = Process.Start(psi);

                if (goferProcess == null)
                {
                    MessageBox.Show(
                        "Failed to launch gofer backend.",
                        "gofer",
                        MessageBoxButton.OK,
                        MessageBoxImage.Error
                    );
                    return;
                }

                var thisProcess = goferProcess;
                thisProcess.EnableRaisingEvents = true;
                thisProcess.Exited += (_, __) =>
                {
                    // Only shut down if this is still the active gofer
                    if (ReferenceEquals(goferProcess, thisProcess))
                    {
                        Dispatcher.Invoke(() => Shutdown());
                    }
                };
            }
            else if (!string.IsNullOrEmpty(gopherUrl))
            {
                ForwardUrlToGofer(gopherUrl);
            }
        }

        private async void ForwardUrlToGofer(string uri)
        {
            try
            {
                string escaped = Uri.EscapeDataString(uri);
                string url = $"http://localhost:8000/focus?uri={escaped}";

                await http.GetAsync(url);
            }
            catch
            {
                // Fire-and-forget, ignore failures
            }
        }
    
        private void OnHome(object? sender, EventArgs e)
        {
            EnsureGoferRunning(null);

            Process.Start(new ProcessStartInfo
            {
                FileName = "http://localhost:8000/landing",
                UseShellExecute = true
            });
        }

        private void OnAbout(object? sender, EventArgs e)
        {
            MessageBox.Show("gofer\nthe internet gopher getter\nversion 0.9.6", "About");
        }


        private void OnQuit(object? sender, EventArgs e)
        {
            try
            {
                if (goferProcess != null && !goferProcess.HasExited)
                {
                    goferProcess.Kill(true);
                }
            }
            catch { }

            if (_tray != null)
            {
                _tray.Visible = false;
                _tray.Dispose();
            }

            Shutdown();
        }
    }
}
