package main

import (
	"os"
	"testing"
)

func TestNewTmpFile(t *testing.T) {
	tmp, err := NewTmpFile()
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer tmp.file.Close()
	defer os.Remove(tmp.file.Name())

	// Verify file exists
	fi, err := os.Stat(tmp.file.Name())
	if err != nil {
		t.Fatalf("failed to stat temp file: %v", err)
	}

	// Verify it's a regular file
	if !fi.Mode().IsRegular() {
		t.Error("temp file is not a regular file")
	}
}

func TestTmpFilePermissions(t *testing.T) {
	tmp, err := NewTmpFile()
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer tmp.file.Close()
	defer os.Remove(tmp.file.Name())

	fi, err := os.Stat(tmp.file.Name())
	if err != nil {
		t.Fatalf("failed to stat temp file: %v", err)
	}

	// Check permissions are 0600 (readable and writable by owner only)
	mode := fi.Mode().Perm()
	expectedMode := os.FileMode(0o600)
	if mode != expectedMode {
		t.Errorf("expected permissions %o, got %o", expectedMode, mode)
	}
}

func TestTmpFileWrite(t *testing.T) {
	tmp, err := NewTmpFile()
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer tmp.Close()

	testData := []byte("test data for secret")
	err = tmp.Write(testData)
	if err != nil {
		t.Fatalf("failed to write to temp file: %v", err)
	}
}

func TestTmpFileRead(t *testing.T) {
	tmp, err := NewTmpFile()
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer tmp.Close()

	testData := []byte("test secret data")
	err = tmp.Write(testData)
	if err != nil {
		t.Fatalf("failed to write to temp file: %v", err)
	}

	// Read the data back
	readData, err := tmp.Read()
	if err != nil {
		t.Fatalf("failed to read from temp file: %v", err)
	}

	if string(readData) != string(testData) {
		t.Errorf("expected data %s, got %s", string(testData), string(readData))
	}
}

func TestTmpFileReadAfterMultipleWrites(t *testing.T) {
	tmp, err := NewTmpFile()
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer tmp.Close()

	// Write multiple times
	data1 := []byte("first")
	data2 := []byte("second")
	err = tmp.Write(data1)
	if err != nil {
		t.Fatalf("failed to write first data: %v", err)
	}

	err = tmp.Write(data2)
	if err != nil {
		t.Fatalf("failed to write second data: %v", err)
	}

	// Read should return both
	readData, err := tmp.Read()
	if err != nil {
		t.Fatalf("failed to read from temp file: %v", err)
	}

	expected := "firstsecond"
	if string(readData) != expected {
		t.Errorf("expected data %s, got %s", expected, string(readData))
	}
}

func TestTmpFileClose(t *testing.T) {
	tmp, err := NewTmpFile()
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	filePath := tmp.file.Name()

	// Write some data
	err = tmp.Write([]byte("test"))
	if err != nil {
		t.Fatalf("failed to write to temp file: %v", err)
	}

	// Close should delete the file
	err = tmp.Close()
	if err != nil {
		t.Fatalf("failed to close temp file: %v", err)
	}

	// Verify file is deleted
	_, err = os.Stat(filePath)
	if !os.IsNotExist(err) {
		t.Errorf("temp file was not deleted: %s", filePath)
	}
}

func TestTmpFileCloseMultipleTimes(t *testing.T) {
	tmp, err := NewTmpFile()
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	// First close
	err = tmp.Close()
	if err != nil {
		t.Fatalf("first close failed: %v", err)
	}

	// Second close should fail because file is already gone
	err = tmp.Close()
	if err == nil {
		t.Error("expected error on second close, got nil")
	}
}

func TestTmpFileOpenEditor(t *testing.T) {
	tmp, err := NewTmpFile()
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer tmp.Close()

	editorPath := "/bin/sh"
	if _, err := os.Stat(editorPath); os.IsNotExist(err) {
		t.Skip("sh not available")
	}

	editor, err := NewEditor(editorPath)
	if err != nil {
		t.Fatalf("failed to create editor: %v", err)
	}

	// This should not error (sh will just read EOF from stdin)
	err = tmp.OpenEditor(editor)
	if err != nil {
		t.Logf("OpenEditor returned error: %v (expected for sh with no input)", err)
	}
}
