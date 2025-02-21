package cli

import (
	"fmt"
	"fraktal/vfs"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/chzyer/readline"
)

type FSCompleter struct {
	vfs *vfs.VirtualFS
}

var commands = map[string]bool{
	"cd":    true,
	"ls":    true,
	"cat":   true,
	"touch": true,
	"rm":    true,
	"mkdir": true,
	"write": true,
	"help":  true,
	"exit":  true,
	"pwd":   true,
	"exec":  true,
}

// Custom autocompleter for commands and vfs files/directories
func (c *FSCompleter) Do(line []rune, pos int) ([][]rune, int) {
	input := string(line)
	tokens := strings.Fields(input)
	if len(tokens) == 0 {
		return nil, 0
	}

	// If only one token is present, complete command names.
	if len(tokens) == 1 {
		prefix := tokens[0]
		var suggestions [][]rune
		for cmd := range commands {
			if strings.HasPrefix(cmd, prefix) {
				// Return only the missing suffix.
				suffix := cmd[len(prefix):]
				suggestions = append(suggestions, []rune(suffix))
			}
		}
		// Replace from the current cursor position.
		return suggestions, pos
	}

	// The last token is the file/folder prefix.
	prefix := tokens[len(tokens)-1]
	items, err := c.vfs.ListDir(".")
	if err != nil {
		return nil, 0
	}

	var suggestions [][]rune
	for _, item := range items {
		if strings.HasPrefix(item.Name, prefix) {
			// Only return the part that hasn't been typed.
			suffix := item.Name[len(prefix):]
			suggestions = append(suggestions, []rune(suffix))
		}
	}

	return suggestions, pos
}

// Start the CLI for interacting with the virtual filesystem
func StartCLI(vfs *vfs.VirtualFS) {
	fmt.Println("In-Memory Virtual Filesystem. Type 'help' for commands.")

	// Initialize the FS completer with vfs instance
	fsCompleter := &FSCompleter{vfs: vfs}

	// Configure readline with the autocomplete completer.
	rl, err := readline.NewEx(&readline.Config{
		Prompt:       fmt.Sprintf("vfs:%s> ", vfs.PrintWorkingDir()),
		AutoComplete: fsCompleter,
	})
	if err != nil {
		panic(err)
	}
	defer rl.Close()

	for {
		line, err := rl.Readline()
		if err != nil {
			break // Exit on error or EOF
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		tokens := strings.Fields(line)
		if len(tokens) == 0 {
			continue
		}
		cmd := tokens[0]
		args := tokens[1:]
		switch cmd {
		case "help":
			printHelp()
		case "ls":
			path := "."
			if len(args) > 0 {
				path = args[0]
			}
			items, err := vfs.ListDir(path)
			if err != nil {
				fmt.Println("Error:", err)
			} else if len(items) > 0 {
				w := tabwriter.NewWriter(os.Stdout, 1, 1, 1, ' ', 0)
				fmt.Fprintln(w, "-- Type", "Name", "Created", "Modified --")
				for _, item := range items {
					itemType := "f"
					if item.IsDir {
						itemType = "d"
					}
					fmt.Fprintln(w, itemType, item.Name, item.Created.Format(time.DateTime), item.Modified.Format(time.DateTime))
				}
				w.Flush()
			}
		case "cd":
			if len(args) < 1 {
				fmt.Println("Usage: cd <path>")
			} else {
				if err := vfs.ChangeDir(args[0]); err != nil {
					fmt.Println("Error:", err)
				}
				rl.SetPrompt(fmt.Sprintf("vfs:%s> ", vfs.PrintWorkingDir()))
			}
		case "mkdir":
			if len(args) < 1 {
				fmt.Println("Usage: mkdir <directory>")
			} else {
				if err := vfs.Mkdir(args[0]); err != nil {
					fmt.Println("Error:", err)
				}
			}
		case "touch":
			if len(args) < 1 {
				fmt.Println("Usage: touch <file>")
			} else {
				if err := vfs.CreateFile(args[0]); err != nil {
					fmt.Println("Error:", err)
				}
			}
		case "cat":
			if len(args) < 1 {
				fmt.Println("Usage: cat <file>")
			} else {
				content, err := vfs.ReadFile(args[0])
				if err != nil {
					fmt.Println("Error:", err)
				} else {
					fmt.Println(content)
				}
			}
		case "write":
			if len(args) < 2 {
				fmt.Println("Usage: write <file> <content>")
			} else {
				file := args[0]
				content := strings.Join(args[1:], " ")
				if err := vfs.WriteFile(file, content); err != nil {
					fmt.Println("Error:", err)
				}
			}
		case "rm":
			if len(args) < 1 {
				fmt.Println("Usage: rm <file|directory>")
			} else {
				if err := vfs.Remove(args[0]); err != nil {
					fmt.Println("Error:", err)
				}
			}
		case "pwd":
			fmt.Println(vfs.PrintWorkingDir())
		case "exit":
			return
		case "exec":
			if len(args) < 1 {
				fmt.Println("Usage: exec <file>")
			} else {
				if err := vfs.ExecFile(args[0]); err != nil {
					fmt.Println("Error:", err)
				}
			}
		default:
			fmt.Println("Unknown command:", cmd)
		}
	}
}

func printHelp() {
	helpText := `Supported commands:
  ls <path>               - List directory contents
  cd <path>               - Change directory
  mkdir <dir>             - Create a new directory
  touch <file>            - Create an empty file
  cat <file>              - Display file contents
  write <file> <content>  - Write content to a file (creates file if not exists)
  rm <file|dir>           - Remove a file or an empty directory
  exec <file>			  - Execute a file from the virtual file system on the host file system
  pwd                     - Print current working directory
  help                    - Show this help message
  exit                    - Exit the CLI`
	fmt.Println(helpText)
}
