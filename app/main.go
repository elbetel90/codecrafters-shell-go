package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/codecrafters-io/shell-starter-go/app/executor"
	"github.com/codecrafters-io/shell-starter-go/app/parser"
)

// Ensures gofmt doesn't remove the "fmt" import in stage 1 (feel free to remove this!)
var _ = fmt.Print

func main() {
	// all_commands := parser.GetAllCommands()

	for {
		fmt.Print("$ ")
		input, err := parser.ReadLine([]string{})
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return
		}

		input = strings.TrimRight(input, "\r\n")
		if len(input) == 0 {
			continue

		}
		command := parser.ParseCommand(input)
		stdout_writer, stderr_writer, cleanup, err := executor.GetOutputWriters(command)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return
		}

		executor.ExecuteCommands(input, command, stdout_writer, stderr_writer)

		cleanup()
	}

}
