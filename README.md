# g<img width="32" height="32" alt="gofer_icon" src="https://github.com/user-attachments/assets/ef3af66b-bb5e-41e8-9801-4028fb396100" />f e r

gofer is an internet gopher getter; a helper agent that allows any web browser to access the gopherholes of gopherspace without extensions.

Mainstream browsers have long since dumped Gopher compatibility, leaving a handful of incompatible browser extensions and specialty clients for Gopher browsing. The goal for gofer is to provide a basic late 90s feature set without over-thinking. It should restore the gopher:// class to the user's machine at the lowest learning curve.  

gofer for gopher is written in go for obvious reasons; namely, to avoid dependancies and virtual machines.

## build
The gofer source is cross platform, and the core application (without OS wrapper) can be built using the go compiler from the source files in the root folder
go build gofer

## use
goper can be invoked from the command line or used with an OS-wrapper (preferred). from the CL: ./gofer gopher://url.url:port (typically port 70). Using gopher from the CL *will not* register gopher:// with the OS.

## wrappers
Wrappers hide the CLI and register gopher:// with the OS for smoother browsing

### macos
Copy the app to your Applications folder to ensure that gopher:// is registered. 

MacOS has a quirk where the OS will only send a URL to an application on its first invocation; this doesn't affect normal gopher browsing unless the user pops in and out of www and gopherspace. If a gopher link is clicked and the gofer window comes forward without launching the page, use the relaunch button and try again.

### windows
The windows wrapper lives in the tray. This wrapper is not quite ready for publishing, but is sufficienty far along that a power-user can compile it with .Net from the windows folder.

### linux (debian/gnome)
The linux prebuilds target gnome on debian and are written in GTK. They have a similar look and feel to the macOS version but gnome handles gopher:// URLs more smoothly than macOS so the relaunch button isn't needed.

## list of supported item types
The list of canonical Gopher item types is provided below, as well the exceptional non-standard types and their current implementation status:

| Type Code | Description | Status | gofer typeIcon |
| :---: | :--- | :--- | :--- |
| **0** | Text File | ✅ | [TXT] |
| **1** | Menu or Directory | ✅ | [ 1 ] |
| **2** | Ph/CSO Server | ✅ connect and search | [PhC] |
| **3** | Error | ✅ | [ERR] |
| **4** | BinHexed Macintosh file | ✅ | [HQX] |
| **5** | DOS binary file archive | ✅ | [DOS] |
| **6** | UNIX uuencoded file | ✅ | [UUE] |
| **7** | Index-Search server | ✅ | [SCH] |
| **8** | Telnet session | ❌ | [TEL] |
| **9** | Binary file (nonspecific) | ✅ | [BIN] |
| **g** | A GIF format graphics file | ✅ | [GIF] |
| **I** | Image file (nonspecific) | ✅ | [IMG] |
| **T** | TN3270 session | ❌ | [327] |
| **+** | Redundant server | ❌ | [RDN] |
| **?** | Non-standard Type Codes | ✅ as generic files | [!X!] |
| **h** | Non-standard, HTML URL (hURL) | ✅ html link | [!h!] |
| **i** | Non-standard, Informational text | ✅ | [ i ] |
