package main

import (
	"fmt"
	"io"
	"os"
)

type TmpFile struct {
	file *os.File
}

func NewTmpFile() (*TmpFile, error) {
	file, err := os.CreateTemp(os.TempDir(), "k8s-secret-editor-")
	if err != nil {
		return nil, err
	}

	// Set restrictive permissions to protect sensitive secret data
	if err := os.Chmod(file.Name(), 0600); err != nil {
		_ = file.Close()
		_ = os.Remove(file.Name())
		return nil, fmt.Errorf("error setting file permissions: %w", err)
	}

	return &TmpFile{file: file}, nil
}

func (t *TmpFile) Write(data []byte) error {
	if _, err := t.file.Write(data); err != nil {
		return fmt.Errorf("error writing to temp file: %w", err)
	}
	return nil
}

func (t *TmpFile) Read() ([]byte, error) {
	// Move the file pointer back to the beginning before reading
	if _, err := t.file.Seek(0, io.SeekStart); err != nil {
		return nil, fmt.Errorf("error seeking to beginning of temp file: %w", err)
	}

	data, err := io.ReadAll(t.file)
	if err != nil {
		return nil, fmt.Errorf("error reading from temp file: %w", err)
	}
	return data, nil
}

func (t *TmpFile) OpenEditor(editor interface{ Open(filePath string) error }) error {
	return editor.Open(t.file.Name())
}

func (t *TmpFile) Close() error {
	name := t.file.Name()
	if err := t.file.Close(); err != nil {
		return err
	}
	// Remove temp file after closing to avoid leaking secrets
	if err := os.Remove(name); err != nil {
		return fmt.Errorf("error removing temp file: %w", err)
	}
	return nil
}
