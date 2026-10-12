package main

import (
	"errors"
	"strings"
)

// processAllFiles emits GitHub Action outputs.
func processAllFiles(
	allFiles []string,
	writeOutput writeFunc,
) error {
	if len(allFiles) == 0 {
		if !writeOutput("has_files", "false") {
			return errors.New("cannot write has_files to GITHUB_OUTPUT")
		}
		return nil
	}

	if !writeOutput("ALL_FILES", strings.Join(allFiles, ",")) {
		return errors.New("cannot write ALL_FILES to GITHUB_OUTPUT")
	}

	if !writeOutput("has_files", "true") {
		return errors.New("cannot write has_files to GITHUB_OUTPUT")
	}

	return nil
}
