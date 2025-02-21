package vfs

import (
	"testing"
	"time"
)

func TestRebuildFilesystem(t *testing.T) {
	// Create a new event log
	eventLog := &EventLog{}

	// Record events to simulate operations:
	// 1. Create a directory "dir1"
	eventLog.Append(FileSystemEvent{
		EventType: EventCreateDir,
		Path:      "dir1",
		Timestamp: time.Now(),
	})
	// 2. Create a file "file.txt"
	eventLog.Append(FileSystemEvent{
		EventType: EventCreateFile,
		Path:      "file.txt",
		Timestamp: time.Now(),
	})
	// 3. Write content to "file.txt"
	eventLog.Append(FileSystemEvent{
		EventType: EventWriteFile,
		Path:      "file.txt",
		Content:   "Hello, world!",
		Timestamp: time.Now(),
	})

	// Create a new, empty filesystem.
	vfsNew := NewVirtualFS(nil)

	// Replay the event log to rebuild the filesystem state.
	if err := eventLog.Replay(vfsNew); err != nil {
		t.Fatalf("Failed to replay event log: %v", err)
	}

	// Validate that the directory "dir1" exists.
	dirNode, err := vfsNew.resolvePath("dir1")
	if err != nil {
		t.Fatalf("Expected directory 'dir1' not found: %v", err)
	}
	if !dirNode.IsDir {
		t.Fatalf("Expected 'dir1' to be a directory")
	}

	// Validate that "file.txt" exists and has the correct content.
	fileNode, err := vfsNew.resolvePath("file.txt")
	if err != nil {
		t.Fatalf("Expected file 'file.txt' not found: %v", err)
	}
	if fileNode.IsDir {
		t.Fatalf("'file.txt' is a directory, expected a file")
	}
	if fileNode.Content != "Hello, world!" {
		t.Errorf("Content mismatch for 'file.txt': got %q, expected %q", fileNode.Content, "Hello, world!")
	}
}
