package main

import (
	"errors"
	"fmt"
	"os"
)

func main() {
	os.Exit(runMain())
}

func runMain() int {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		return 1
	}

	return 0
}

func run() error {
	return runWith(
		validateEnvironment,
		createOutputFile,
		storeTranslationPaths,
		(*os.File).Close,
	)
}

func runWith(
	validate func() (envConfig, error),
	createFile func() (*os.File, error),
	store storePathsFunc,
	closeFile func(*os.File) error,
) (err error) {
	// Read and validate inputs from the environment.
	cfg, err := validate()
	if err != nil {
		return err
	}

	// We persist the generated pathspecs to a file that is later consumed by
	// tj-actions/changed-files via `files_from_source_file`.
	file, err := createFile()
	if err != nil {
		return fmt.Errorf("cannot create output file: %w", err)
	}

	defer func() {
		if closeErr := closeFile(file); closeErr != nil {
			err = errors.Join(err, fmt.Errorf("cannot close output file: %w", closeErr))
		}
	}()

	// Emit one pathspec per line. Consumers expect newline-separated patterns.
	// Each line can be a direct file path or a glob (git pathspec-style).
	if err := store(cfg, file); err != nil {
		return fmt.Errorf("cannot store translation paths: %w", err)
	}

	return nil
}
