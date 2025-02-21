package vfs

import (
	"testing"
)

func TestMkdirAndResolvePath(t *testing.T) {
	vfs := NewVirtualFS(nil)
	// Create a directory named "test"
	if err := vfs.Mkdir("test"); err != nil {
		t.Fatalf("Mkdir failed: %v", err)
	}
	// Resolve the "test" directory
	node, err := vfs.resolvePath("test")
	if err != nil {
		t.Fatalf("ResolvePath failed: %v", err)
	}
	if !node.IsDir {
		t.Fatalf("Expected 'test' to be a directory")
	}
}

func TestCreateFileAndReadWrite(t *testing.T) {
	vfs := NewVirtualFS(nil)
	// Create a file named "file.txt"
	if err := vfs.CreateFile("file.txt"); err != nil {
		t.Fatalf("CreateFile failed: %v", err)
	}

	// Write content to "file.txt"
	content := "Hello, VirtualFS!"
	if err := vfs.WriteFile("file.txt", content); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	// Read the content back
	readContent, err := vfs.ReadFile("file.txt")
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	if readContent != content {
		t.Fatalf("ReadFile content mismatch: got %q, want %q", readContent, content)
	}
}

func TestListDir(t *testing.T) {
	vfs := NewVirtualFS(nil)
	// Create a directory and a file in the root
	if err := vfs.Mkdir("dir1"); err != nil {
		t.Fatalf("Mkdir failed: %v", err)
	}
	if err := vfs.CreateFile("file1.txt"); err != nil {
		t.Fatalf("CreateFile failed: %v", err)
	}

	// List the root directory
	items, err := vfs.ListDir(".")
	if err != nil {
		t.Fatalf("ListDir failed: %v", err)
	}

	expected := map[string]bool{"dir1": true, "file1.txt": true}
	for _, item := range items {
		if !expected[item.Name] {
			t.Errorf("Unexpected item: %s", item.Name)
		}
		delete(expected, item.Name)
	}
	if len(expected) != 0 {
		t.Errorf("Missing expected items: %v", expected)
	}
}

func TestRemove(t *testing.T) {
	vfs := NewVirtualFS(nil)
	// Create and remove a file
	if err := vfs.CreateFile("file.txt"); err != nil {
		t.Fatalf("CreateFile failed: %v", err)
	}
	if err := vfs.Remove("file.txt"); err != nil {
		t.Fatalf("Remove failed: %v", err)
	}
	if _, err := vfs.resolvePath("file.txt"); err == nil {
		t.Fatalf("Expected error when resolving removed file, got nil")
	}

	// Create and remove an empty directory
	if err := vfs.Mkdir("dir1"); err != nil {
		t.Fatalf("Mkdir failed: %v", err)
	}
	if err := vfs.Remove("dir1"); err != nil {
		t.Fatalf("Remove directory failed: %v", err)
	}
	if _, err := vfs.resolvePath("dir1"); err == nil {
		t.Fatalf("Expected error when resolving removed directory, got nil")
	}
}

func TestChangeDirAndPrintWorkingDir(t *testing.T) {
	vfs := NewVirtualFS(nil)
	// Create a directory structure /a/b
	if err := vfs.Mkdir("a"); err != nil {
		t.Fatalf("Mkdir failed: %v", err)
	}
	if err := vfs.ChangeDir("a"); err != nil {
		t.Fatalf("ChangeDir to /a failed: %v", err)
	}
	if err := vfs.Mkdir("b"); err != nil {
		t.Fatalf("Mkdir failed: %v", err)
	}
	if err := vfs.ChangeDir("b"); err != nil {
		t.Fatalf("ChangeDir to /a/b failed: %v", err)
	}
	pwd := vfs.PrintWorkingDir()
	expected := "/a/b"
	if pwd != expected {
		t.Errorf("PrintWorkingDir mismatch: got %s, want %s", pwd, expected)
	}
}

func TestErrorHandling(t *testing.T) {
	vfs := NewVirtualFS(nil)
	// Create a file and try to list it as a directory
	if err := vfs.CreateFile("file.txt"); err != nil {
		t.Fatalf("CreateFile failed: %v", err)
	}
	if _, err := vfs.ListDir("file.txt"); err == nil {
		t.Fatalf("Expected error when listing a file as directory")
	}
	// Create a directory and try to read it as a file
	if err := vfs.Mkdir("dir1"); err != nil {
		t.Fatalf("Mkdir failed: %v", err)
	}
	if _, err := vfs.ReadFile("dir1"); err == nil {
		t.Fatalf("Expected error when reading a directory")
	}
}

func TestRelativePaths(t *testing.T) {
	vfs := NewVirtualFS(nil)
	// Create a directory structure /a/b
	if err := vfs.Mkdir("a"); err != nil {
		t.Fatalf("Mkdir failed: %v", err)
	}
	if err := vfs.ChangeDir("a"); err != nil {
		t.Fatalf("ChangeDir to /a failed: %v", err)
	}
	if err := vfs.Mkdir("b"); err != nil {
		t.Fatalf("Mkdir failed: %v", err)
	}
	if err := vfs.ChangeDir("b"); err != nil {
		t.Fatalf("ChangeDir to /a/b failed: %v", err)
	}
	// Move back to /a using relative path ".."
	if err := vfs.ChangeDir(".."); err != nil {
		t.Fatalf("ChangeDir '..' failed: %v", err)
	}
	pwd := vfs.PrintWorkingDir()
	expected := "/a"
	if pwd != expected {
		t.Errorf("PrintWorkingDir mismatch after '..': got %s, want %s", pwd, expected)
	}
}

func TestCreateFileAbsolutePath(t *testing.T) {
	vfs := NewVirtualFS(nil)
	// Create a directory "foo" at the root.
	if err := vfs.Mkdir("foo"); err != nil {
		t.Fatalf("Mkdir failed: %v", err)
	}
	// Change directory to "foo".
	if err := vfs.ChangeDir("foo"); err != nil {
		t.Fatalf("ChangeDir failed: %v", err)
	}
	// Now create a file with an absolute path that starts with /foo.
	if err := vfs.CreateFile("/foo/bar.txt"); err != nil {
		t.Fatalf("CreateFile failed: %v", err)
	}
	// Resolve the file using an absolute path.
	node, err := vfs.resolvePath("/foo/bar.txt")
	if err != nil {
		t.Fatalf("ResolvePath failed: %v", err)
	}
	if node == nil || node.IsDir {
		t.Errorf("Expected file node, got %v", node)
	}
}

func TestCreateFileRelativePath(t *testing.T) {
	vfs := NewVirtualFS(nil)
	// Create a directory "foo" and change into it.
	if err := vfs.Mkdir("foo"); err != nil {
		t.Fatalf("Mkdir failed: %v", err)
	}
	if err := vfs.ChangeDir("foo"); err != nil {
		t.Fatalf("ChangeDir failed: %v", err)
	}
	// Create a file with a relative path.
	if err := vfs.CreateFile("bar.txt"); err != nil {
		t.Fatalf("CreateFile failed: %v", err)
	}
	// The file should be located in the current directory.
	node, err := vfs.resolvePath("bar.txt")
	if err != nil {
		t.Fatalf("ResolvePath failed: %v", err)
	}
	if node == nil || node.IsDir {
		t.Errorf("Expected file node, got %v", node)
	}
}

func TestCreateFileAndRestore(t *testing.T) {
	vfs1 := NewVirtualFS(nil)

	// Create a directory and change into it
	if err := vfs1.Mkdir("foo"); err != nil {
		t.Fatalf("Mkdir failed: %v", err)
	}
	if err := vfs1.ChangeDir("foo"); err != nil {
		t.Fatalf("ChangeDir failed: %v", err)
	}
	// Create a file using a relative path.
	if err := vfs1.CreateFile("bar.txt"); err != nil {
		t.Fatalf("CreateFile failed: %v", err)
	}

	// Events should be logged
	if len(vfs1.EventLog.events) != 2 {
		t.Fatal("Expected at two events to be logged")
	}

	event := vfs1.EventLog.events[1]
	// event.Path should be absolute
	expectedPath := "/foo/bar.txt"
	if event.Path != expectedPath {
		t.Errorf("Expected event path %q, got %q", expectedPath, event.Path)
	}

	// Simulate restoring the filesystem
	vfs2 := NewVirtualFS(nil)
	// Replay events from vfs1.EventLog
	if err := vfs1.EventLog.Replay(vfs2); err != nil {
		t.Fatalf("Replay failed: %v", err)
	}
	// File should now be created at /foo/bar.txt in vfs2
	node, err := vfs2.resolvePath("/foo/bar.txt")
	if err != nil || node == nil || node.IsDir {
		t.Fatalf("File /foo/bar.txt was not restored correctly")
	}
}
