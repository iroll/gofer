# gofer

Gofer is a helper agent that allows any web browser to access **Gopherspace** without extensions.

As of 2025, mainstream browsers have long since dumped Gopher compatibility, leaving a handful of browser extensions (mostly incompatible) and specialty clients for Gopher browsing. The goal for Gofer is to provide a basic **1995 - 2000 feature set** as simply as possible, and to be distributed in a way that restores the gopher:// class to the user's machine. The stretch goal for Gofer is to provide an advanced 1995 - 2000 feature set without spilling the spaghetti. That's easier said than done—item type `[2]` already supposes that the user wants to search a Ph server, and yes, Gofer can do that.

Gofer for Gopher is written in **Go** for obvious reasons; namely, to provide a compiled package for easy distribution and avoidance of dependencies (i.e., Python).

Usage: a macos executable - gofer - is currently provided; otherwise, gofer will need to be built from the provided source. Gofer can be called from the command line, e.g.:

gofer gopher://freeshell.org:70

... and a default browser window will open. Currently gofer will close after 60s of inactivity. Alternatively, your OS or browser may let you register gopher:// links to point to gofer, which eliminates the need for command line launches or timeouts. Packaged releases for macos, debian, and windows that will automatically register gopher:// are a future goal.

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
