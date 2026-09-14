package executor

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/codecrafters-io/shell-starter-go/app/parser"
)

// file setup
func openFile(path string, is_append bool) (*os.File, error) {
	if path == "" {
		return nil, nil
	}

	flag := os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	if is_append {
		flag = os.O_WRONLY | os.O_CREATE | os.O_APPEND
	}
	file, err := os.OpenFile(path, flag, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open output file: %w", err)
	}

	return file, err
}

// get standard output writers for output and errors
func GetOutputWriters(command *parser.Command) (io.Writer, io.Writer, func(), error) {
	var stdout_writer io.Writer = os.Stdout
	var stderr_writer io.Writer = os.Stderr
	var closers []func()

	clean_up := func() {
		for _, close_fn := range closers {
			close_fn()
		}
	}

	if command.OutputFile != "" {
		if err := os.MkdirAll(filepath.Dir(command.OutputFile), 0755); err != nil {
			fmt.Fprintf(os.Stderr, "error creating directory: %v\n", err)
			return nil, nil, clean_up, nil
		}
		f, err := openFile(command.OutputFile, command.AppendStdout)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error opening file: %v\n", err)
			return nil, nil, clean_up, nil
		}
		out_closer := f
		closers = append(closers, func() { out_closer.Close() })
		stdout_writer = f
	}

	if command.ErrorFile != "" {
		if err := os.MkdirAll(filepath.Dir(command.ErrorFile), 0755); err != nil {
			fmt.Fprintf(os.Stderr, "error creating directory: %v\n", err)
			return nil, nil, clean_up, nil
		}
		f, err := openFile(command.ErrorFile, command.AppendStderr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error opening file: %v\n", err)
			return nil, nil, clean_up, nil
		}
		err_closer := f
		closers = append(closers, func() { err_closer.Close() })
		stderr_writer = f
	}

	return stdout_writer, stderr_writer, clean_up, nil
}
