# gofer
Gofer is a helper agent that allows any web browser to access **Gopherspace** without extensions.

Mainstream browsers have long since dumped Gopher compatibility, leaving a handful of incompatible browser extensions and specialty clients for Gopher browsing. The goal for Gofer is to provide a basic **1995 - 2000 feature set** as simply as possible, without over-thinking presentation. It is intended to be distributed in a way that restores the gopher:// class to the user's machine.  

Gofer for Gopher is written in **Go** for obvious reasons; namely, to provide a compiled package for easy distribution and avoidance of dependencies (i.e., Python).

## gofer 0.5
Development of gofer 0.5 focused on the basic gopher client functions, type handling, and UI.
Usage: a macos executable - gofer - is currently provided; otherwise, gofer will need to be built from the provided source. Gofer 0.5 can be called from the command line, e.g.:

gofer gopher://freeshell.org:70

... and a default browser window will open. Currently gofer will close after 60s of inactivity in the main window and will need to be launched from the command line again. 

## gofer 0.9
Development on gofer 0.9 is focused on some architectural improvements, namely the persistence of the gofer helper app and building native packages for macOS, debian, and windows that will automatically register gopher:// handling with the OS. UI beautification is the intended scope to reach 1.0.

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
