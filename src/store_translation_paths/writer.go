package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// writeUniqueLine writes a normalized newline-terminated pathspec once.
func writeUniqueLine(writer io.Writer, seen map[string]struct{}, pathspec string) error {
	line := filepath.ToSlash(filepath.Clean(pathspec))
	if line == "." {
		return errors.New("empty pathspec")
	}

	if _, ok := seen[line]; ok {
		return nil
	}

	if _, err := io.WriteString(writer, line+"\n"); err != nil {
		return err
	}

	seen[line] = struct{}{}
	return nil
}

// createOutputFile creates the temp file consumed later by changed-files.
func createOutputFile() (*os.File, error) {
	dir := filepath.Join(".git", "lokalise-action")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("cannot create output directory: %w", err)
	}

	path := filepath.Join(dir, "paths.txt")

	file, err := os.Create(path)
	if err != nil {
		return nil, fmt.Errorf("cannot create output file %q: %w", path, err)
	}

	return file, nil
}
