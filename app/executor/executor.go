package executor

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"

	"github.com/codecrafters-io/shell-starter-go/app/parser"
	"github.com/codecrafters-io/shell-starter-go/app/types"
)

type CommandCompletionSpec struct {
	CommandName   string // -C flag
	TargetCommand string // target command
}

var completion_registry = make(map[string]CommandCompletionSpec)

func handlePwd() {
	dir, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, "pwd error", err)
		os.Exit(1)
	}
	fmt.Println(dir)
}

func handleCd(args []string) {
	arg := args[0]
	if arg == "~" || strings.HasPrefix(arg, "~/") {
		homeDir, err := os.UserHomeDir()
		if err == nil {
			if arg == "~" {
				arg = homeDir
			} else {
				arg = filepath.Join(homeDir, arg[2:])
			}
		}
	}
	err := os.Chdir(arg)
	if err != nil {
		fmt.Printf("cd: %s: No such file or directory\n", args[0])
	}
}

func handleType(args []string) {
	if slices.Contains(types.Built_ins, args[0]) {
		fmt.Println(args[0] + " is a shell builtin")
	} else if path, err := exec.LookPath(args[0]); err == nil {
		fmt.Println(args[0] + " is " + path)
	} else if args[0] == "type" {
		fmt.Println(args[0] + ": not found")
	} else {
		fmt.Println(args[0] + ": not found")
	}
}

func handleComplete(command *parser.Command, stdout_writer, stderr_writer io.Writer) {
	args := command.Args

	completion_registry["git"] = CommandCompletionSpec{
		CommandName:   "'./'",
		TargetCommand: "git",
	}

	if len(args) == 0 {
		return
	}
	if args[0] == string(types.CompleteCommandArgsP) {
		if len(args) == 1 {
			for _, spec := range completion_registry {
				fmt.Fprintln(stdout_writer, printSpecs(spec))
			}
		}

		target_command := args[1]
		spec, exists := completion_registry[target_command]
		if !exists {
			fmt.Fprintf(stderr_writer, "complete: %s: no completion specification\n", target_command)
			return
		}
		fmt.Fprintln(stdout_writer, printSpecs(spec))
	}
}

func printSpecs(spec CommandCompletionSpec) string {
	out := "complete"
	if spec.CommandName != "" {
		out += " -C " + spec.CommandName
	}
	out += " " + spec.TargetCommand

	return out
}

// execute commands
func ExecuteCommands(input string, command *parser.Command, stdout_writer, stderr_writer io.Writer) {
	if command == nil || command.Cmd == "" {
		return
	}
	switch command.Cmd {
	case string(types.BuiltinCommandExit):
		os.Exit(0)
	case string(types.BuiltinCommandEcho):
		fmt.Fprintln(stdout_writer, strings.Join(command.Args, " "))
	case string(types.BuiltinCommandPwd):
		handlePwd()
	case string(types.BuiltinCommandCd):
		handleCd(command.Args)
	case string(types.BuiltinCommandType):
		handleType(command.Args)
	case string(types.BuiltinCommandComplete):
		handleComplete(command, stdout_writer, stderr_writer)
	default:
		if _, err := exec.LookPath(command.Cmd); err == nil {
			execCmd := exec.Command(command.Cmd, command.Args...)
			execCmd.Stdout = stdout_writer
			execCmd.Stderr = stderr_writer
			execCmd.Run()
		} else {
			fmt.Println(input + ": command not found")
		}
	}
}
