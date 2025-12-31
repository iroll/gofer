# g<img width="32" height="32" alt="gofer_icon" src="https://github.com/user-attachments/assets/ef3af66b-bb5e-41e8-9801-4028fb396100" />f e r

gofer is an internet gopher getter; a helper agent that allows any web browser to access the gopherholes of gopherspace without extensions.

Mainstream browsers have long since dumped Gopher compatibility, leaving a handful of incompatible browser extensions and specialty clients for Gopher browsing. The goal for gofer is to provide a basic late 90s feature set without over-thinking. It should restore the gopher:// class to the user's machine at the lowest learning curve.  

gofer for gopher is written in go for obvious reasons; namely, to avoid dependancies and virtual machines.

## build
go build gofer

## use
goper can be invoked from the command line or used with an OS-wrapper (preferred). from the CL: ./gofer gopher://url.url:port (typically port 70)

### macos
[macOS .app [arm64]](https://github.com/user-attachments/files/24388950/gofer.zip)


a prebuilt macos binary and gofer.app are provided. MacOS has a quirk where the OS will only send a URL to a program on its first invocation; this doesn't affect normal gopher browsing unless the use pops in and out of www and gopherspace. If a gopher link is clicked and the gofer window comes forward without launching the page, use the relaunch button and try again.

### windows
in progress

### linux (debian/gnome)

<a href="https://github.com/iroll/gofer/blob/master/gofer_0.9-1%20debian/gofer_0.9-1_arm64.deb">Debian .deb [arm64]</a>

The linux prebuild targets gnome on debian.


## list of supported item types
The basic list of Gopher item types is provided below, as well as their current implementation status:

| Type Code | Description | Status |
| :---: | :--- | :--- |
| **0** | Text File | **[check!]** |
| **1** | Menu or Directory | **[check!]** |
| **2** | Ph/CSO Server | **[check! connect and search]** |
| **3** | Error | **[check!]** |
| **4** | BinHexed Macintosh file |**[check!]** |
| **5** | DOS binary file archive |**[check!]** |
| **6** | UNIX uuencoded file |**[check!]** |
| **7** | Index-Search server |**[check!]** |
| **8** | Telnet session | |
| **9** | Binary file (nonspecific) |**[check!]** |
| **+** | Redundant server | |
| **T** | TN3270 session | |
| **g** | A GIF format graphics file |**[check!]** |
| **I** | Image file (nonspecific) |**[check!]** |
| **?** | Non-standard Type Codes |**[check! as generic files]** |
