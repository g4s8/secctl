package main

import (
	"fmt"
	"io"
	"os"
	"path"
)

type TmpFile struct {
	path string
}

func NewTmpFile(suffix string) (*TmpFile, error) {
	name := fmt.Sprintf("k8s-secret-editor-%s", suffix)
	p := path.Join(os.TempDir(), name)
	f, err := os.Create(p)
	if err != nil {
		return nil, fmt.Errorf("error creating temp file: %w", err)
	}
	if err := f.Close(); err != nil {
		_ = os.Remove(f.Name())
		return nil, fmt.Errorf("error closing temp file: %w", err)
	}
	// Set restrictive permissions to protect sensitive secret data
	if err := os.Chmod(p, 0o600); err != nil {
		_ = os.Remove(p)
		return nil, fmt.Errorf("error setting file permissions: %w", err)
	}

	return &TmpFile{path: p}, nil
}

func (t *TmpFile) Write(data []byte) error {
	f, err := os.OpenFile(t.path, os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return fmt.Errorf("error opening temp file for writing: %w", err)
	}
	if _, err := f.Write(data); err != nil {
		return fmt.Errorf("error writing to temp file: %w", err)
	}
	if err := f.Sync(); err != nil {
		return fmt.Errorf("error syncing temp file: %w", err)
	}
	return nil
}

func (t *TmpFile) Read() ([]byte, error) {
	f, err := os.Open(t.path)
	if err != nil {
		return nil, fmt.Errorf("error opening temp file for reading: %w", err)
	}

	data, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("error reading from temp file: %w", err)
	}
	return data, nil
}

func (t *TmpFile) OpenEditor(editor interface{ Open(filePath string) error }) error {
	if err := editor.Open(t.path); err != nil {
		return fmt.Errorf("error opening editor: %w", err)
	}
	return nil
}

func (t *TmpFile) Close() error {
	if err := os.Remove(t.path); err != nil {
		return fmt.Errorf("error removing temp file: %w", err)
	}
	return nil
}
