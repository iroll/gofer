# gofer

Gofer is a helper agent that allows any web browser to access **Gopherspace** without extensions.

As of 2025, mainstream browsers have long since dumped Gopher compatibility, leaving a handful of browser extensions (mostly incompatible) and specialty clients for Gopher browsing. The goal for Gofer is to provide a basic **1995 - 2000 feature set** as simply as possible, and to be distributed in a way that restores the gopher:// class to the user's machine. The stretch goal for Gofer is to provide an advanced 1995 - 2000 feature set without spilling the spaghetti. That's easier said than done—item type `[2]` already supposes that the user wants to search a Ph server, and yes, Gofer can do that.

Gofer for Gopher is written in **Go** for obvious reasons; namely, to provide a compiled package for easy distribution and avoidance of dependencies (i.e., Python).

The basic list of Gopher item types is provided below, as well as their current implementation status:

| Type Code | Description | Status |
| :---: | :--- | :--- |
| **0** | File | **[check!]** |
| **1** | Directory | **[check!]** |
| **2** | Ph/CSO Server | **[check! connect and search]** |
| **3** | Error | **[check!]** |
| **4** | BinHexed Macintosh file | |
| **5** | DOS binary file archive | |
| **6** | UNIX uuencoded file | |
| **7** | Index-Search server | |
| **8** | Telnet session | |
| **9** | Binary file (nonspecific) | |
| **+** | Redundant server | |
| **T** | TN3270 session | |
| **g** | A GIF format graphics file | |
| **I** | Image file (nonspecific) | |
