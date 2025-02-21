package vfs

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
)

// File or directory in the virtual filesystem
type FSNode struct {
	Name     string
	IsDir    bool
	Content  string
	Children map[string]*FSNode
	Parent   *FSNode
	Created  time.Time
	Modified time.Time
}

type VirtualFS struct {
	Root        *FSNode   // fs root
	Cwd         *FSNode   // Current working directory
	EventLog    *EventLog // Event log
	IsRestoring bool      // Flag for system restore in progress (no events will be recorded)
}

func (vfs *VirtualFS) AbsolutePath(path string) (string, error) {
	// If the path is already absolute, return it
	if strings.HasPrefix(path, "/") {
		return path, nil
	}

	// Otherwise prepend current working directory
	cwd := vfs.PrintWorkingDir()
	if cwd == "/" {
		return "/" + path, nil
	}
	return cwd + "/" + path, nil
}

// Returns the absolute path of the current working directory
func (vfs *VirtualFS) PrintWorkingDir() string {
	var parts []string
	node := vfs.Cwd
	for node != nil && node != vfs.Root {
		parts = append([]string{node.Name}, parts...)
		node = node.Parent
	}
	return "/" + strings.Join(parts, "/")
}

// Initialize a new virtual filesystem with a root directory
func NewVirtualFS(redisClient *redis.Client) *VirtualFS {
	root := &FSNode{
		Name:     "",
		IsDir:    true,
		Children: make(map[string]*FSNode),
		Created:  time.Now(),
		Modified: time.Now(),
	}

	vfs := &VirtualFS{
		Root:        root,
		Cwd:         root,
		EventLog:    NewEventLog(redisClient),
		IsRestoring: false,
	}

	// No redis client provided - soft exit for easy testing
	if redisClient == nil {
		return vfs
	}

	eventLog, err := RestoreEventLog(redisClient)
	if err != nil {
		fmt.Printf("Failed to restore event log: %v\n", err)
		return vfs
	}

	// Restore filesystem
	vfs.IsRestoring = true
	vfs.EventLog = eventLog
	err = eventLog.Replay(vfs)
	if err != nil {
		fmt.Printf("Failed to replay events: %v\n", err)
	}
	vfs.IsRestoring = false

	return vfs
}
