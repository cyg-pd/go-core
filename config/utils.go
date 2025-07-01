package config

import (
	"os"
	"path"
	"path/filepath"
)

func getAvailablePath() []string {
	var availablePath = []string{}
	if p, err := os.Getwd(); err == nil && p != "" {
		availablePath = append(availablePath, allParentLevel(p)...)
	} else if absPath, err := filepath.Abs("."); err == nil {
		availablePath = append(availablePath, absPath)
	} else {
		availablePath = append(availablePath, ".")
	}
	return availablePath
}

func allParentLevel(p string) []string {
	paths := []string{}
	parentPath := p
	for {
		paths = append(paths, parentPath)
		if parentPath == "/" {
			break
		}
		parentPath = path.Dir(parentPath)
	}

	return paths
}
