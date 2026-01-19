//  main.swift 0.9.5
//  gofer wrapper for macOS written in Swift
//  (C) 2025 Isaac Roll

import Cocoa

// Application Delegate
class AppDelegate: NSObject, NSApplicationDelegate {
    var window: NSWindow!
    var statusLabel: NSTextField!
    var goferProcess: Process?

    //  MARK: - App Lifecycle
    func applicationDidFinishLaunching(_ notification: Notification) {
        // Register URL handler immediately
        NSAppleEventManager.shared().setEventHandler(
            self,
            andSelector: #selector(handleURLEvent(_:withReplyEvent:)),
            forEventClass: AEEventClass(kInternetEventClass),
            andEventID: AEEventID(kAEGetURL)
        )

        setupUI()
        setupMenu()
        launchGofer()
    }

    // MARK: - Menu & UI

    func setupMenu() {
        let mainMenu = NSMenu()

        let appMenuItem = NSMenuItem()
        mainMenu.addItem(appMenuItem)

        let appMenu = NSMenu()
        appMenuItem.submenu = appMenu

        let aboutItem = NSMenuItem(
            title: "About gofer",
            action: #selector(showAbout(_:)),
            keyEquivalent: ""
        )
        aboutItem.target = self
        appMenu.addItem(aboutItem)

        appMenu.addItem(NSMenuItem.separator())

        let homeItem = NSMenuItem(
            title: "Home",
            action: #selector(objc_goHome(_:)),
            keyEquivalent: "h"
        )
        homeItem.target = self
        appMenu.addItem(homeItem)

        let quitItem = NSMenuItem(
            title: "Quit gofer",
            action: #selector(quitApp(_:)),
            keyEquivalent: "q"
        )
        quitItem.target = self
        appMenu.addItem(quitItem)

        NSApp.mainMenu = mainMenu
    }

    func setupUI() {
        let mask: NSWindow.StyleMask = [.titled, .closable, .miniaturizable]
        window = NSWindow(
            contentRect: NSRect(x: 0, y: 0, width: 300, height: 120),
            styleMask: mask,
            backing: .buffered,
            defer: false
        )
        window.title = "gofer"
        window.center()

        let container = NSView(frame: window.contentView!.frame)

        statusLabel = NSTextField(labelWithString: "gofer is digging")
        statusLabel.frame = NSRect(x: 20, y: 70, width: 260, height: 24)
        statusLabel.alignment = .center
        statusLabel.font = .systemFont(ofSize: 16)

        let homeBtn = NSButton(title: "Home", target: nil, action: nil)
        homeBtn.frame = NSRect(x: 60, y: 20, width: 80, height: 32)
        homeBtn.bezelStyle = .rounded
        homeBtn.target = self
        homeBtn.action = #selector(objc_goHome(_:))

        let quitBtn = NSButton(title: "Quit", target: nil, action: nil)
        quitBtn.frame = NSRect(x: 160, y: 20, width: 80, height: 32)
        quitBtn.bezelStyle = .rounded
        quitBtn.target = self
        quitBtn.action = #selector(quitApp)

        container.addSubview(statusLabel)
        container.addSubview(homeBtn)
        container.addSubview(quitBtn)
        window.contentView = container

        window.makeKeyAndOrderFront(nil)
        NSApp.activate(ignoringOtherApps: true)
    }

    // MARK: - gofer launcher

    // Manual app launch path (no URL)
    func launchGofer() {
        let binPath = Bundle.main.bundleURL
            .appendingPathComponent("Contents/MacOS/gofer").path

        let process = Process()
        process.executableURL = URL(fileURLWithPath: binPath)
        process.arguments = []
        process.terminationHandler = { _ in self.goferProcess = nil }

        try? process.run()
        self.goferProcess = process
    }

    // URL trampoline: always exec gofer with the URL
    // (even if already running)
    func execGofer(_ urlString: String) {
        let binPath = Bundle.main.bundleURL
            .appendingPathComponent("Contents/MacOS/gofer").path

        let process = Process()
        process.executableURL = URL(fileURLWithPath: binPath)
        process.arguments = [urlString]

        try? process.run()
    }

    // MARK: - Handle URL Events from the system
    
    @objc func handleURLEvent(_ event: NSAppleEventDescriptor,
                              withReplyEvent replyEvent: NSAppleEventDescriptor) {
        guard let urlString =
            event.paramDescriptor(forKeyword: keyDirectObject)?.stringValue
        else { return }

        execGofer(urlString)
    }

    func application(_ app: NSApplication, open urls: [URL]) {
        guard let url = urls.first else { return }
        execGofer(url.absoluteString)
    }

    // MARK: - Actions

    @objc(goHome:)
    func objc_goHome(_ sender: Any?) {
        NSWorkspace.shared.open(URL(string: "http://localhost:8000/landing")!)
    }

    @objc func quitApp(_ sender: Any?) {
        let task = Process()
        task.executableURL = URL(fileURLWithPath: "/usr/bin/killall")
        task.arguments = ["gofer"]
        try? task.run()

        NSApplication.shared.terminate(nil)
    }

    @objc func showAbout(_ sender: Any?) {
        let alert = NSAlert()
        alert.messageText = "About gofer"
        alert.informativeText = "internet gopher getter\nversion 0.9.5"
        alert.addButton(withTitle: "OK")
        alert.runModal()
    }

}

// Entry Point
let app = NSApplication.shared
let delegate = AppDelegate()
app.delegate = delegate
app.run()
