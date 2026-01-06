using System;
using System.Windows;
using MessageBox = System.Windows.MessageBox;
using NotifyIcon = System.Windows.Forms.NotifyIcon;
using ContextMenuStrip = System.Windows.Forms.ContextMenuStrip;
using ToolStripSeparator = System.Windows.Forms.ToolStripSeparator;

namespace GoferWrapper
{
    public partial class App : System.Windows.Application
    {
        private NotifyIcon? _tray;

        protected override void OnStartup(StartupEventArgs e)
        {
            base.OnStartup(e);

            _tray = new NotifyIcon
            {
                Text = "gofer",
                Icon = System.Drawing.SystemIcons.Application,
                Visible = true,
                ContextMenuStrip = new ContextMenuStrip()
            };

            _tray.ContextMenuStrip.Items.Add("Home", null, OnHome);
            _tray.ContextMenuStrip.Items.Add("About", null, OnAbout);
            _tray.ContextMenuStrip.Items.Add(new ToolStripSeparator());
            _tray.ContextMenuStrip.Items.Add("Quit", null, OnQuit);
        }

        private void OnHome(object? sender, EventArgs e)
        {
            MessageBox.Show("Home (not wired yet)", "gofer");
        }

        private void OnAbout(object? sender, EventArgs e)
        {
            MessageBox.Show("gofer\nGopher client", "About");
        }

        private void OnQuit(object? sender, EventArgs e)
        {
            if (_tray != null)
            {
                _tray.Visible = false;
                _tray.Dispose();
            }

            Shutdown();
        }
    }
}
