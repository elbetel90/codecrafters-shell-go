package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/codecrafters-io/shell-starter-go/app/executor"
	"github.com/codecrafters-io/shell-starter-go/app/parser"
	"github.com/codecrafters-io/shell-starter-go/app/types"
)

// Ensures gofmt doesn't remove the "fmt" import in stage 1 (feel free to remove this!)
var _ = fmt.Print

func main() {
	c := parser.NewCommand()

	histfile_var := os.Getenv("HISTFILE")
	if histfile_var != "" {
		executor.LoadHistory(histfile_var)
		types.HistoryAppendIndex = len(types.History)
	}

	// all_commands := c.GetAllCommands()

	for {
		executor.ReapJobs(os.Stdout)

		fmt.Print("$ ")
		input, err := c.ReadLine([]string{})
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return
		}

		input = strings.TrimRight(input, "\r\n")
		if len(input) == 0 {
			continue

		}
		types.History = append(types.History, input)
		commands := c.ParseCommand(input)
		if len(commands) > 1 {
			command_executor := executor.NewCommandExecutor(input, commands[0], os.Stdout, os.Stderr)
			command_executor.ExecutePipeline(commands, os.Stdout, os.Stderr)

			continue
		}
		command := commands[0]
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
