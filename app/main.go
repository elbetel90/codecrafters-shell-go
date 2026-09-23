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
	all_commands := parser.GetAllCommands()

	for {
		executor.ReapJobs(os.Stdout)

		fmt.Print("$ ")
		input, err := parser.ReadLine(all_commands)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return
		}

		input = strings.TrimRight(input, "\r\n")
		if len(input) == 0 {
			continue

		}
		c := parser.NewCommand()
		command := c.ParseCommand(input)
		stdout_writer, stderr_writer, cleanup, err := executor.GetOutputWriters(command)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return
		}

		command_executor := executor.NewCommandExecutor(input, command, stdout_writer, stderr_writer)

		command_executor.ExecuteCommands(input, command, stdout_writer, stderr_writer)

		cleanup()
	}

}
