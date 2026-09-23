package executor

import (
	"cmp"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"syscall"

	"github.com/codecrafters-io/shell-starter-go/app/parser"
	"github.com/codecrafters-io/shell-starter-go/app/types"
)

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

	if len(args) == 0 {
		return
	}

	if args[0] == string(types.CompleteCommandArgsP) {
		if len(args) == 1 {
			for _, spec := range types.CompletionRegistry {
				fmt.Fprintln(stdout_writer, printSpecs(command.Cmd, spec))
			}
		}

		target_command := args[1]
		spec, exists := types.CompletionRegistry[target_command]
		if !exists {
			fmt.Fprintf(stderr_writer, "complete: %s: no completion specification\n", target_command)
			return
		}
		fmt.Fprintln(stdout_writer, printSpecs(command.Cmd, spec))
	} else if args[0] == string(types.CompleteCommandArgsC) {
		if len(args) == 1 || len(args) == 2 {
			return
		}
		target := args[len(args)-1]
		command_name := args[len(args)-2]
		types.CompletionRegistry[target] = types.CommandCompletionSpec{
			CommandName:   command_name,
			TargetCommand: target,
		}
	} else if args[0] == string(types.CompleteCommandArgsR) {
		if len(args) < 2 {
			return
		}
		target := args[1]
		delete(types.CompletionRegistry, target)
	}
}

func handleJobs(stdout_writer io.Writer) {
	slices.SortFunc(types.Jobs, func(a, b types.Job) int {
		return cmp.Compare(a.JobNumber, b.JobNumber)
	})
	updateJobStatuses()
	n := len(types.Jobs)
	for i, job := range types.Jobs {
		marker := jobMarker(i, n)
		if job.Status == string(types.JobStatusRunning) {
			fmt.Fprintf(stdout_writer, "[%d]%s  %-24s%s\n", job.JobNumber, marker, string(types.JobStatusRunning), job.Command+" &")
		} else {
			fmt.Fprintf(stdout_writer, "[%d]%s  %-24s%s\n", job.JobNumber, marker, string(types.JobStatusDone), job.Command)
		}
	}
	remaining := types.Jobs[:0]
	for _, job := range types.Jobs {
		if job.Status != string(types.JobStatusDone) {
			remaining = append(remaining, job)
		}
	}
	types.Jobs = remaining
}

func runBackgroudJobs(command *parser.Command) (int, int, error) {
	if len(command.Args) < 1 {
		return 0, 0, nil
	}
	cmd := exec.Command(command.Cmd, command.Args[:len(command.Args)-1]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Start()
	if err != nil {
		return 0, 0, err
	}

	args_without_amp := command.Args[:len(command.Args)-1]

	job := types.Job{
		JobNumber: types.NextJobNumber,
		Pid:       cmd.Process.Pid,
		Command:   command.Cmd + " " + strings.Join(args_without_amp, " "),
		Status:    string(types.JobStatusRunning),
	}
	types.NextJobNumber++

	types.Jobs = append(types.Jobs, job)

	return job.JobNumber, cmd.Process.Pid, nil
}

func printSpecs(cmd string, spec types.CommandCompletionSpec) string {
	out := cmd
	if spec.CommandName != "" {
		out += " -C '" + spec.CommandName + "'"
	}
	out += " " + spec.TargetCommand

	return out
}

func printJobDetail(job_number, pid int) string {
	out := fmt.Sprintf("[%d] %d", job_number, pid)
	return out
}

func updateJobStatuses() {
	for i, job := range types.Jobs {
		var ws syscall.WaitStatus
		wpid, err := syscall.Wait4(job.Pid, &ws, syscall.WNOHANG, nil)
		if err != nil {
			continue
		} else if wpid == job.Pid {
			if ws.Exited() {
				types.Jobs[i].Status = string(types.JobStatusDone)
			}
		}
	}
}

func jobMarker(i, n int) string {
	switch i {
	case n - 1:
		return "+"
	case n - 2:
		return "-"
	}

	return " "
}

func ReapJobs(stdout_writer io.Writer) {
	slices.SortFunc(types.Jobs, func(a, b types.Job) int {
		return cmp.Compare(a.JobNumber, b.JobNumber)
	})

	updateJobStatuses()
	n := len(types.Jobs)

	for i, job := range types.Jobs {
		if job.Status != string(types.JobStatusDone) {
			continue
		}
		marker := jobMarker(i, n)
		fmt.Fprintf(stdout_writer, "[%d]%s  %-24s%s\n", job.JobNumber, marker, string(types.JobStatusDone), job.Command)
	}

	remaining_jobs := types.Jobs[:0]
	for _, job := range types.Jobs {
		if job.Status != string(types.JobStatusDone) {
			remaining_jobs = append(remaining_jobs, job)
		}
	}

	types.Jobs = remaining_jobs

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
	case string(types.BuiltinCommandJobs):
		handleJobs(stdout_writer)
	default:
		if len(command.Args) > 0 && command.Args[len(command.Args)-1] == "&" {
			job_number, pid, err := runBackgroudJobs(command)
			if err != nil {
				fmt.Fprintln(stderr_writer, err)
				return
			}
			fmt.Fprintln(stdout_writer, printJobDetail(job_number, pid))
		} else {
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
}
