package vfs

import (
	"errors"
	"fmt"
	"fraktal/memfd"
	"os"
	"syscall"
	"time"
)

// Creates a new directory at the specified path
func (vfs *VirtualFS) Mkdir(path string) error {
	absPath, err := vfs.AbsolutePath(path)
	if err != nil {
		return err
	}

	parent, name, err := vfs.traverseToParent(absPath)
	if err != nil {
		return err
	}
	if _, exists := parent.Children[name]; exists {
		return fmt.Errorf("directory or file already exists: %s", name)
	}
	newDir := &FSNode{
		Name:     name,
		IsDir:    true,
		Children: make(map[string]*FSNode),
		Parent:   parent,
		Created:  time.Now(),
		Modified: time.Now(),
	}
	parent.Children[name] = newDir
	parent.Modified = time.Now()

	// Record event
	if !vfs.IsRestoring {
		event := FileSystemEvent{
			EventType: EventCreateDir,
			Path:      absPath,
			Timestamp: time.Now(),
		}
		vfs.EventLog.Append(event)
	}

	return nil
}

// CreateFile creates an empty file at the specified path
func (vfs *VirtualFS) CreateFile(path string) error {
	absPath, err := vfs.AbsolutePath(path)
	if err != nil {
		return err
	}

	parent, name, err := vfs.traverseToParent(absPath)
	if err != nil {
		return err
	}
	if _, exists := parent.Children[name]; exists {
		return fmt.Errorf("file or directory already exists: %s", name)
	}
	newFile := &FSNode{
		Name:     name,
		IsDir:    false,
		Content:  "",
		Parent:   parent,
		Created:  time.Now(),
		Modified: time.Now(),
	}
	parent.Children[name] = newFile
	parent.Modified = time.Now()

	// Record event
	if !vfs.IsRestoring {
		event := FileSystemEvent{
			EventType: EventCreateFile,
			Path:      absPath,
			Timestamp: time.Now(),
		}
		vfs.EventLog.Append(event)
	}

	return nil
}

// Returns the content of the file at the given path
func (vfs *VirtualFS) ReadFile(path string) (string, error) {
	node, err := vfs.resolvePath(path)
	if err != nil {
		return "", err
	}
	if node.IsDir {
		return "", fmt.Errorf("cannot read directory: %s", path)
	}
	return node.Content, nil
}

// Writes content to the file at the given path
// If the file does not exist, it is created
func (vfs *VirtualFS) WriteFile(path string, content string) error {
	absPath, err := vfs.AbsolutePath(path)
	if err != nil {
		return err
	}

	node, err := vfs.resolvePath(absPath)
	if err != nil {
		// File does not exist, so create it.
		err = vfs.CreateFile(absPath)
		if err != nil {
			return err
		}
		node, err = vfs.resolvePath(absPath)
		if err != nil {
			return err
		}
	}
	if node.IsDir {
		return fmt.Errorf("cannot write to directory: %s", path)
	}
	node.Content = content
	node.Modified = time.Now()

	// Record event
	if !vfs.IsRestoring {
		event := FileSystemEvent{
			EventType: EventWriteFile,
			Path:      absPath,
			Content:   content,
			Timestamp: time.Now(),
		}
		vfs.EventLog.Append(event)
	}

	return nil
}

// Deletes a file/directory at the given path
// Directories must be empty to be removed
func (vfs *VirtualFS) Remove(path string) error {
	node, err := vfs.resolvePath(path)
	if err != nil {
		return err
	}
	if node == vfs.Root {
		return errors.New("cannot remove root directory")
	}
	if node.IsDir && len(node.Children) > 0 {
		return fmt.Errorf("directory not empty: %s", path)
	}
	parent := node.Parent
	delete(parent.Children, node.Name)
	parent.Modified = time.Now()

	// Record event
	if !vfs.IsRestoring {
		event := FileSystemEvent{
			EventType: EventDelete,
			Path:      path,
			Timestamp: time.Now(),
		}
		vfs.EventLog.Append(event)
	}

	return nil
}

// Returns a list of names of all items in the directory at the given path
func (vfs *VirtualFS) ListDir(path string) ([]*FSNode, error) {
	node, err := vfs.resolvePath(path)
	if err != nil {
		return nil, err
	}
	if !node.IsDir {
		return nil, fmt.Errorf("not a directory: %s", path)
	}
	var items []*FSNode
	for _, item := range node.Children {
		items = append(items, item)
	}
	return items, nil
}

// Changes the current working directory
func (vfs *VirtualFS) ChangeDir(path string) error {
	node, err := vfs.resolvePath(path)
	if err != nil {
		return err
	}
	if !node.IsDir {
		return fmt.Errorf("not a directory: %s", path)
	}
	vfs.Cwd = node
	return nil
}

func (vfs *VirtualFS) ExecFile(path string) error {
	node, err := vfs.resolvePath(path)
	if err != nil {
		return err
	}
	if node.IsDir {
		return fmt.Errorf("%s is a directory", path)
	}

	// Create a script using the file content
	script := []byte("#!/bin/sh\n" + node.Content + "\n")

	// Create an in-memory file using memfd_create
	fd, err := memfd.Create(node.Name, 0)
	if err != nil {
		return fmt.Errorf("memfdCreate failed: %v", err)
	}

	// Write the script content to the file descriptor
	// This is a write call but it's to a volatile file in memory
	if _, err := syscall.Write(fd, script); err != nil {
		return fmt.Errorf("write failed: %v", err)
	}

	// Make the in-memory file executable.
	if err := syscall.Fchmod(fd, 0755); err != nil {
		return fmt.Errorf("fchmod failed: %v", err)
	}

	// Obtain a path to the in-memory file via /proc/self/fd/<fd>.
	fdPath := fmt.Sprintf("/proc/self/fd/%d", fd)

	// Execute the file
	args := []string{fdPath}
	// Inherit the environment
	env := os.Environ()
	pid, err := syscall.ForkExec(fdPath, args, &syscall.ProcAttr{
		Dir:   "",
		Env:   env,
		Files: []uintptr{uintptr(syscall.Stdin), uintptr(syscall.Stdout), uintptr(syscall.Stderr)},
	})
	if err != nil {
		return fmt.Errorf("forkExec failed: %v", err)
	}

	// Wait for the child process to finish.
	var ws syscall.WaitStatus
	_, err = syscall.Wait4(pid, &ws, 0, nil)
	if err != nil {
		return fmt.Errorf("wait4 failed: %v", err)
	}

	// Close the file descriptor
	syscall.Close(fd)

	return nil
}
