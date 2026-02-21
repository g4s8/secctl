package main

import (
	"fmt"
	"os"
	"os/exec"
)

type Editor struct {
	path string
}

func NewEditor(path string) (*Editor, error) {
	if path == "" {
		path = os.Getenv("EDITOR")
		if path == "" {
			return nil, fmt.Errorf("EDITOR environment variable is not set")
		}
	}

	fi, err := os.Stat(path) // Check if the editor exists
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("editor '%s' does not exist", path)
		}
		return nil, fmt.Errorf("error checking editor '%s': %w", path, err)
	}

	// Check if the editor is executable
	if fi.IsDir() {
		return nil, fmt.Errorf("editor '%s' is a directory, not an executable", path)
	}
	if fi.Mode()&0o111 == 0 {
		return nil, fmt.Errorf("editor '%s' is not executable", path)
	}
	return &Editor{path: path}, nil
}

func (e *Editor) Open(filePath string) error {
	cmd := exec.Command(e.path, filePath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error opening editor '%s': %w", e.path, err)
	}
	return nil
}
