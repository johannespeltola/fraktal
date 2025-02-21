package cli

import (
	"fraktal/vfs"
	"strings"
	"testing"
)

func TestEmptyInput(t *testing.T) {
	v := vfs.NewVirtualFS(nil)
	completer := &FSCompleter{vfs: v}

	// Empty string should not result in suggestions
	input := []rune("")
	suggestions, pos := completer.Do(input, 0)
	if suggestions != nil || pos != 0 {
		t.Errorf("Expected no suggestions and pos=0 for empty input, got suggestions=%v, pos=%d", suggestions, pos)
	}
}

func TestCommandAutocomplete(t *testing.T) {
	v := vfs.NewVirtualFS(nil)
	completer := &FSCompleter{vfs: v}

	inputStr := "c"
	input := []rune(inputStr)
	pos := len(inputStr)
	suggestions, start := completer.Do(input, pos)
	if start != pos {
		t.Errorf("Expected start pos %d, got %d", pos, start)
	}

	// Input "c" should produce suggestions "cd" and "cat"
	validCommands := []string{"cd", "cat"}
	// Verify that suggestion are valid commands.
	found := make(map[string]bool)
	for _, sug := range suggestions {
		candidate := inputStr + string(sug)
		for _, vc := range validCommands {
			if candidate == vc {
				found[vc] = true
			}
		}
	}
	for _, cmd := range validCommands {
		if !found[cmd] {
			t.Errorf("Expected completion for command %q not found", cmd)
		}
	}
}

func TestFileAutocomplete(t *testing.T) {
	v := vfs.NewVirtualFS(nil)
	// Populate vfs with files and directories
	if err := v.CreateFile("file1.txt"); err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	if err := v.CreateFile("anotherfile.txt"); err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	if err := v.Mkdir("folder1"); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}

	completer := &FSCompleter{vfs: v}

	// Use an input that represents a file command with an argument.
	inputStr := "cat f"
	input := []rune(inputStr)
	pos := len(inputStr)
	suggestions, _ := completer.Do(input, pos)

	// "f" should generate suggestions "file1.txt" and "folder1".
	expectedCandidates := []string{"file1.txt", "folder1"}
	prefix := "f"

	for _, candidate := range expectedCandidates {
		if strings.HasPrefix(candidate, prefix) {
			expectedSuffix := candidate[len(prefix):]
			found := false
			for _, sug := range suggestions {
				if string(sug) == expectedSuffix {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected file completion suffix %q for candidate %q not found", expectedSuffix, candidate)
			}
		}
	}
}
