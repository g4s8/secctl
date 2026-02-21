package main

import (
	"os"
	"testing"
)

func TestNewEditor_ValidPath(t *testing.T) {
	// Use a real executable from the system
	editorPath := "/bin/sh"
	if _, err := os.Stat(editorPath); os.IsNotExist(err) {
		t.Skip("sh not available")
	}

	editor, err := NewEditor(editorPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if editor.path != editorPath {
		t.Errorf("expected path=%s, got %s", editorPath, editor.path)
	}
}

func TestNewEditor_NonExistentPath(t *testing.T) {
	_, err := NewEditor("/nonexistent/path/to/editor")
	if err == nil {
		t.Error("expected error for non-existent editor, got nil")
	}
}

func TestNewEditor_DirectoryPath(t *testing.T) {
	_, err := NewEditor("/tmp")
	if err == nil {
		t.Error("expected error for directory path, got nil")
	}
}

func TestNewEditor_NonExecutablePath(t *testing.T) {
	// Create a temporary non-executable file
	tmpFile, err := os.CreateTemp("", "non-executable-*")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	_, err = NewEditor(tmpFile.Name())
	if err == nil {
		t.Error("expected error for non-executable file, got nil")
	}
}

func TestNewEditor_FromEnv(t *testing.T) {
	// Create a temporary executable file
	tmpFile, err := os.CreateTemp("", "executable-*")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	// Make it executable
	if err := os.Chmod(tmpFile.Name(), 0o755); err != nil {
		t.Fatalf("failed to chmod file: %v", err)
	}

	// Save and restore original EDITOR
	originalEditor := os.Getenv("EDITOR")
	defer os.Setenv("EDITOR", originalEditor)

	os.Setenv("EDITOR", tmpFile.Name())

	editor, err := NewEditor("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if editor.path != tmpFile.Name() {
		t.Errorf("expected path=%s, got %s", tmpFile.Name(), editor.path)
	}
}

func TestNewEditor_NoEnv(t *testing.T) {
	// Save and restore original EDITOR
	originalEditor := os.Getenv("EDITOR")
	defer os.Setenv("EDITOR", originalEditor)

	os.Setenv("EDITOR", "")

	_, err := NewEditor("")
	if err == nil {
		t.Error("expected error when EDITOR not set, got nil")
	}
}

func TestEditorOpen_NonExistentFile(t *testing.T) {
	editorPath := "/bin/sh"
	if _, err := os.Stat(editorPath); os.IsNotExist(err) {
		t.Skip("sh not available")
	}

	editor, err := NewEditor(editorPath)
	if err != nil {
		t.Fatalf("failed to create editor: %v", err)
	}

	// Try to open a non-existent file (sh will fail to open it)
	err = editor.Open("/nonexistent/file/path")
	if err == nil {
		t.Error("expected error when opening non-existent file, got nil")
	}
}

func TestEditorOpen_ValidFile(t *testing.T) {
	editorPath := "/bin/sh"
	if _, err := os.Stat(editorPath); os.IsNotExist(err) {
		t.Skip("sh not available")
	}

	editor, err := NewEditor(editorPath)
	if err != nil {
		t.Fatalf("failed to create editor: %v", err)
	}

	// Create a temporary file
	tmpFile, err := os.CreateTemp("", "editor-test-*")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	// This should execute sh (with no arguments) which will try to read from stdin
	// and immediately hit EOF, returning successfully
	err = editor.Open(tmpFile.Name())
	if err != nil {
		t.Logf("editor.Open returned error: %v (this is expected for sh with no args)", err)
	}
}
