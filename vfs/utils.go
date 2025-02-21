package vfs

import (
	"fmt"
	"strings"
)

// Resolves an absolute or relative path to an FSNode.
func (vfs *VirtualFS) resolvePath(path string) (*FSNode, error) {
	if path == "" {
		return vfs.Cwd, nil
	}

	var node *FSNode
	if strings.HasPrefix(path, "/") {
		node = vfs.Root
		path = strings.TrimPrefix(path, "/")
	} else {
		node = vfs.Cwd
	}

	if path == "" {
		return node, nil
	}

	parts := strings.Split(path, "/")
	for _, part := range parts {
		if part == "" || part == "." {
			continue
		}
		if part == ".." {
			if node.Parent != nil {
				node = node.Parent
			}
			continue
		}
		child, exists := node.Children[part]
		if !exists {
			return nil, fmt.Errorf("path not found: %s", part)
		}
		node = child
	}
	return node, nil
}

// Resolve a path to parent directory and returns the final element name.
func (vfs *VirtualFS) traverseToParent(path string) (*FSNode, string, error) {
	path = strings.TrimSuffix(path, "/")
	parts := strings.Split(path, "/")
	if len(parts) == 0 {
		return nil, "", fmt.Errorf("invalid path")
	}
	name := parts[len(parts)-1]
	parentPath := strings.Join(parts[:len(parts)-1], "/")
	parent, err := vfs.resolvePath(parentPath)
	if err != nil {
		return nil, "", err
	}
	if !parent.IsDir {
		return nil, "", fmt.Errorf("not a directory: %s", parentPath)
	}
	return parent, name, nil
}
