# Overview

**Disclaimer**: I unfortunately only had about 10 hours to commit to this challenge due to a lot of work related pressure, but I am still very happy with the result. Most of the functionality I wanted to implement is here, but with some additional time I would implement additional features like permission handling as well as improve the existing once by adding additional tests and performing some minor refactoring.

This project implements an in-memory virtual filesystem in Go that supports basic Unix-like file operations. It provides both a programmatic API and a command-line interface (CLI) for interacting with the filesystem. Operations include file and directory creation/deletion, content read/write/execution, directory listing, and path navigation (supporting both relative and absolute paths).

# Demo
A video demonstrating the virtual file system and its functionalities is available in `demo/demo.webm`.
Same is also available below as a gif but in low quality.
<a href="https://ibb.co/YBX99SpK"><img src="https://i.ibb.co/608hhSgp/demo-demo-1.gif" alt="demo-demo-1" border="0"></a>

# Features

- File and Directory Operations: Create, delete, execute, and list files/directories.
- Executing of files from the virtual file system on the host. **Please make sure you know what you are doing if using this feature**
- Content Reading/Writing: Supports reading from and writing to files.
- Path Navigation: Handles both absolute and relative paths.
- Metadata Tracking: Basic metadata (creation and modification times) is maintained.
- CLI Interface: Unix-like commands such as ls, cd, mkdir, touch, cat, write, rm, exec, and pwd are supported.
- File system is persisted between sessions using file system events and a remote redis server
- Due remote persistance, the state of the system can also be remotely controlled and/or validated
- Autocomplete for commands and files (tab)

# Usage

## Prerequisites

- Working Go installation (only tested on 1.24 but >1.15 should be fine)
- Only tested on Ubuntu 24.04 LTS other Linux distros should work also.
  - Non-Linux environments should also be supported except for the exec command but I have not been able to test this

## Running

### 1. Build the Application:

```
go build -o vfs vfs.go
```

### 2. Run the application

To run the virtual file system simply run the executable built in the previous step.

```
./vfs
```

To persist the file system over sessions you can provide the `--secret` flag with the correct secret for decrypting the redis connection configuration.

```
./vfs --secret=SECRET
```

I opted for secret decryption instead of environment variables or similar mainly for three reasons:

- Initializing variables from a env file is an I/O operation
- Did not want to store configuration for the file system in a system-wide env variable
- I wanted for the file system to be run and deployed using only a single executable

## Commands

The CLI uses the virtual file system API to implement the following unix-like commands:

- `ls <path>` - List directory contents
- `cd <path>` - Change directory
- `mkdir <dir>` - Create a new directory
- `touch <file>` - Create an empty file
- `cat <file>` - Display file contents
- `write <file> <content>` - Write content to a file (creates file if not exists)
- `rm <file|dir>` - Remove a file or an empty directory
- `exec <file>` - Execute a file from the virtual file system on the host file system using [memfd_create](https://www.man7.org/linux/man-pages/man2/memfd_create.2.html)
- `pwd` - Print current working directory
- `help` - Show this help message
- `exit` - Exit the CLI

# Design Decisions

- **VirtualFS Structure**: A VirtualFS struct holds the root directory and the current working directory, enabling stateful management of the filesystem.
- **FSNode Structure**: Every file or directory is represented by an FSNode that contains properties like name, type (file or directory), content (for files), children (for directories), and basic metadata.
- **Error Handling**: CLI errors are gracefully handled ensuring a robust user experience. The exception to this is any secret decryption related commands which result in an unrecoverable error by design

## Data Structure

### FSNode

- **Name**: Name of the file/directory
- **IsDir**: Flag to indicate whether the node is a directory
- **Content**: Holds file data (empty for directories)
- **Children**: A map storing all children of the node for easy and efficient lookup
- **Parent**: A pointer to the parent node for upward navigation
- **Created**/**Modified**: Timestamps for tracking metadata

### VirtualFS

- **Root**: Pointer to the root node (`/`)
- **Cwd**: Pointer to the current working directory
- **EventLog**: Pointer to the event log for storing file system operations - used for restoring the file system
- **IsRestoring**: Flag to indicate that file system restoration is in progress, mainly so that events are not duplicated

## Persistance

The system is persisted using a remote Redis (actully Dragonfly) server. This was done mainly for convenience and easy of use. This also allows for the file system to be remotely controlled which can be a nice feature depending on the use-case. Since it also does not rely on a system specific call, this same approach will also work on any host system.

An alternative, platform specific implementation would have been to persist the file system in volatile memory using `memfd_create` in Linux. I used the same approach for supporting execution of files. This would be a great backup in case the remote alternative is not available, but due to time constraints I did not have time to explore this.

# Known Limitations & Improvements

- Large files may cause unexpected behaviors due to in-memory design
- Advanced metadata such as permissions, ownership etc. is not implemented due to time constraints
- The CLI uses basic token splitting in command parsing that does not support more complex use-cases
- File system (event) state should support snapshots for better performance when restoring large file systems
- Masking redis traffic could be useful for defense evasion
- File metadata (created/modified) is not persisted between sessions, mainly due to time limitations
- Better testing of behaviors related to `exec` commands since I added this feature very last minute
