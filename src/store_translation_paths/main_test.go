package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"
	"testing"
)

func TestRunWith(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		t.Parallel()

		wantCfg := envConfig{
			Paths:       []string{"translations"},
			BaseLang:    "en",
			FileExts:    []string{"json"},
			NamePattern: "",
			FlatNaming:  true,
		}

		validateCalled := false
		createCalled := false
		storeCalled := false
		closeCalled := false

		var createdFile *os.File

		validate := func() (envConfig, error) {
			validateCalled = true
			return wantCfg, nil
		}

		createFile := func() (*os.File, error) {
			createCalled = true

			f, err := os.CreateTemp(t.TempDir(), "pathspecs-*.txt")
			if err != nil {
				t.Fatalf("failed to create temp file: %v", err)
			}
			createdFile = f
			return f, nil
		}

		store := func(cfg envConfig, writer io.Writer) error {
			storeCalled = true

			assertConfigEqual(t, cfg, wantCfg)

			if writer != createdFile {
				t.Fatalf("writer mismatch. want=%v got=%v", createdFile, writer)
			}

			return nil
		}

		closeFile := func(file *os.File) error {
			closeCalled = true
			if file != createdFile {
				return fmt.Errorf("closeFile got unexpected file. want=%v got=%v", createdFile, file)
			}

			return file.Close()
		}

		err := runWith(validate, createFile, store, closeFile)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !validateCalled {
			t.Fatal("validate was not called")
		}
		if !createCalled {
			t.Fatal("createFile was not called")
		}
		if !storeCalled {
			t.Fatal("store was not called")
		}
		if !closeCalled {
			t.Fatal("closeFile was not called")
		}
	})

	t.Run("returns validate error and stops", func(t *testing.T) {
		t.Parallel()

		validate := func() (envConfig, error) {
			return envConfig{}, errors.New("bad env")
		}

		createFile := func() (*os.File, error) {
			t.Fatal("createFile should not be called")
			return nil, nil
		}

		store := func(envConfig, io.Writer) error {
			t.Fatal("store should not be called")
			return nil
		}

		closeFile := func(*os.File) error {
			t.Helper()
			t.Fatal("closeFile should not be called")
			return nil
		}

		err := runWith(validate, createFile, store, closeFile)
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if !strings.Contains(err.Error(), "bad env") {
			t.Fatalf("expected error containing %q, got %q", "bad env", err.Error())
		}
	})

	t.Run("wraps create output file error and stops", func(t *testing.T) {
		t.Parallel()

		validate := func() (envConfig, error) {
			return envConfig{
				Paths:      []string{"translations"},
				BaseLang:   "en",
				FileExts:   []string{"json"},
				FlatNaming: true,
			}, nil
		}

		createFile := func() (*os.File, error) {
			return nil, errors.New("permission denied")
		}

		store := func(envConfig, io.Writer) error {
			t.Fatal("store should not be called")
			return nil
		}

		closeFile := func(*os.File) error {
			t.Helper()
			t.Fatal("closeFile should not be called")
			return nil
		}

		err := runWith(validate, createFile, store, closeFile)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "cannot create output file") {
			t.Fatalf("expected wrapped error containing %q, got %q", "cannot create output file", err.Error())
		}
		if !strings.Contains(err.Error(), "permission denied") {
			t.Fatalf("expected wrapped error containing %q, got %q", "permission denied", err.Error())
		}
	})

	t.Run("wraps store error and still closes file", func(t *testing.T) {
		t.Parallel()

		wantCfg := envConfig{
			Paths:      []string{"translations"},
			BaseLang:   "en",
			FileExts:   []string{"json"},
			FlatNaming: true,
		}

		storeErr := errors.New("disk full")

		var createdFile *os.File
		closeCalled := false

		validate := func() (envConfig, error) {
			return wantCfg, nil
		}

		createFile := func() (*os.File, error) {
			f, err := os.CreateTemp(t.TempDir(), "pathspecs-*.txt")
			if err != nil {
				t.Fatalf("failed to create temp file: %v", err)
			}

			createdFile = f
			return f, nil
		}

		store := func(cfg envConfig, writer io.Writer) error {
			assertConfigEqual(t, cfg, wantCfg)

			if writer != createdFile {
				t.Fatalf("writer mismatch. want=%v got=%v", createdFile, writer)
			}

			return storeErr
		}

		closeFile := func(file *os.File) error {
			closeCalled = true

			if file != createdFile {
				return fmt.Errorf(
					"closeFile got unexpected file. want=%v got=%v",
					createdFile,
					file,
				)
			}

			return file.Close()
		}

		err := runWith(validate, createFile, store, closeFile)
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		if !errors.Is(err, storeErr) {
			t.Fatalf("expected error wrapping %v, got %v", storeErr, err)
		}

		if !strings.Contains(err.Error(), "cannot store translation paths") {
			t.Fatalf(
				"expected error containing %q, got %q",
				"cannot store translation paths",
				err.Error(),
			)
		}

		if !closeCalled {
			t.Fatal("closeFile was not called")
		}
	})
}

func assertConfigEqual(t *testing.T, got, want envConfig) {
	t.Helper()

	if !slices.Equal(got.Paths, want.Paths) {
		t.Fatalf("paths mismatch. want=%v got=%v", want.Paths, got.Paths)
	}

	if got.BaseLang != want.BaseLang {
		t.Fatalf("baseLang mismatch. want=%q got=%q", want.BaseLang, got.BaseLang)
	}

	if !slices.Equal(got.FileExts, want.FileExts) {
		t.Fatalf("fileExts mismatch. want=%v got=%v", want.FileExts, got.FileExts)
	}

	if got.NamePattern != want.NamePattern {
		t.Fatalf(
			"namePattern mismatch. want=%q got=%q",
			want.NamePattern,
			got.NamePattern,
		)
	}

	if got.FlatNaming != want.FlatNaming {
		t.Fatalf(
			"flatNaming mismatch. want=%v got=%v",
			want.FlatNaming,
			got.FlatNaming,
		)
	}
}
