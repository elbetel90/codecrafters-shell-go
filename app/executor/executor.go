package executor

import (
	"cmp"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"
	"syscall"

	"github.com/codecrafters-io/shell-starter-go/app/parser"
	"github.com/codecrafters-io/shell-starter-go/app/types"
	"github.com/codecrafters-io/shell-starter-go/app/utils"
)

type CommandExecutor struct {
	Input        string
	Commmand     *parser.Command
	StdoutWriter io.Writer
	StderrWriter io.Writer
}

func NewCommandExecutor(input string, command *parser.Command, stdout_writer, stderr_writer io.Writer) *CommandExecutor {
	return &CommandExecutor{
		input,
		command,
		stdout_writer,
		stdout_writer,
	}
}

// execute commands
func (ce *CommandExecutor) ExecuteCommands(input string, command *parser.Command, stdout_writer, stderr_writer io.Writer) {
	if command == nil || command.Cmd == "" {
		return
	}
	switch command.Cmd {
	case string(types.BuiltinCommandExit):
		AppendHistory(os.Getenv("HISTFILE"))
		os.Exit(0)
	case string(types.BuiltinCommandEcho):
		fmt.Fprintln(stdout_writer, strings.Join(command.Args, " "))
	case string(types.BuiltinCommandPwd):
		ce.handlePwd()
	case string(types.BuiltinCommandCd):
		ce.handleCd(command.Args)
	case string(types.BuiltinCommandType):
		ce.handleType(command.Args)
	case string(types.BuiltinCommandComplete):
		ce.handleComplete(command, stdout_writer, stderr_writer)
	case string(types.BuiltinCommandJobs):
		ce.handleJobs(stdout_writer)
	case string(types.BuiltinCommandsHistory):
		ce.handleHistory(stdout_writer, command)
	case string(types.BuiltinCommandDeclare):
		ce.handleDeclare(stdout_writer, stderr_writer, command)
	default:
		if len(command.Args) > 0 && command.Args[len(command.Args)-1] == "&" {
			job_number, pid, err := ce.runBackgroudJobs(command)
			if err != nil {
				fmt.Fprintln(stderr_writer, err)
				return
			}
			fmt.Fprintln(stdout_writer, ce.printJobDetail(job_number, pid))
		} else {
			if _, err := exec.LookPath(command.Cmd); err == nil {
				exec_cmd := exec.Command(command.Cmd, command.Args...)
				exec_cmd.Stdout = stdout_writer
				exec_cmd.Stderr = stderr_writer
				exec_cmd.Run()
			} else {
				fmt.Println(input + ": command not found")
			}
		}
	}
}

func (ce *CommandExecutor) runBuiltInPipeline(command *parser.Command, stdout_writer, stderr_writer io.Writer) {
	switch command.Cmd {
	case string(types.BuiltinCommandEcho):
		fmt.Fprintln(stdout_writer, strings.Join(command.Args, " "))
	case string(types.BuiltinCommandType):
		for _, arg := range command.Args {
			if slices.Contains(types.Built_ins, arg) {
				fmt.Fprintln(stdout_writer, arg+" is a shell builtin")
			} else if path, err := exec.LookPath(arg); err == nil {
				fmt.Fprintln(stdout_writer, arg+" is "+path)
			} else {
				fmt.Fprintln(stdout_writer, arg+": not found")
			}
		}
	case string(types.BuiltinCommandPwd):
		dir, err := os.Getwd()
		if err != nil {
			fmt.Fprintln(stderr_writer, "pwd error", err)
			return
		}
		fmt.Fprintln(stdout_writer, dir)
	}
}

func (ce *CommandExecutor) ExecutePipeline(commands []*parser.Command, stdout_writer, stderr_writer io.Writer) {
	n := len(commands)
	if n == 0 {
		return
	}

	type pipePair struct {
		r, w *os.File
	}

	pipes := make([]pipePair, n-1)
	for i := range pipes {
		r, w, err := os.Pipe()
		if err != nil {
			fmt.Fprintln(stderr_writer, err)
			return
		}
		pipes[i] = pipePair{r, w}
	}

	stage_stdout := make([]io.Writer, n)
	stage_stdin := make([]*os.File, n)
	stage_stdin[0] = os.Stdin
	stage_stdout[n-1] = stdout_writer
	for i := 0; i < n-1; i++ {
		stage_stdout[i] = pipes[i].w
		stage_stdin[i+1] = pipes[i].r
	}

	goroutine_owns_write := make([]bool, n-1)

	var wg sync.WaitGroup
	var external_cmds []*exec.Cmd
	for i, command := range commands {
		if types.IsBuiltin(command.Cmd) {
			wg.Add(1)
			w := stage_stdout[i]
			cmd := command
			if i < n-1 {
				goroutine_owns_write[i] = true
			}
			go func() {
				defer wg.Done()
				if f, ok := w.(*os.File); ok && f != os.Stdout {
					defer f.Close()
				}
				ce.runBuiltInPipeline(cmd, w, stderr_writer)
			}()
		} else {
			if _, err := exec.LookPath(command.Cmd); err != nil {
				fmt.Fprintln(stderr_writer, command.Cmd+": command not found")
				for _, p := range pipes {
					p.r.Close()
					p.w.Close()
				}
				wg.Wait()
				return
			}
			exec_cmd := exec.Command(command.Cmd, command.Args...)
			exec_cmd.Stdin = stage_stdin[i]
			exec_cmd.Stdout = stage_stdout[i]
			exec_cmd.Stderr = stderr_writer
			external_cmds = append(external_cmds, exec_cmd)
			if err := exec_cmd.Start(); err != nil {
				fmt.Fprintln(stderr_writer, err)
				continue
			}
			wg.Add(1)
			go func(cmd *exec.Cmd) {
				defer wg.Done()
				cmd.Wait()
			}(exec_cmd)
		}
	}

	for i, p := range pipes {
		p.r.Close()
		if !goroutine_owns_write[i] {
			p.w.Close()
		}
	}

	wg.Wait()
}

func (ce *CommandExecutor) handlePwd() {
	dir, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, "pwd error", err)
		os.Exit(1)
	}
	fmt.Println(dir)
}

func (ce *CommandExecutor) handleCd(args []string) {
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

func (ce *CommandExecutor) handleType(args []string) {
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

func (ce *CommandExecutor) handleComplete(command *parser.Command, stdout_writer, stderr_writer io.Writer) {
	args := command.Args

	if len(args) == 0 {
		return
	}

	if args[0] == string(types.CompleteCommandArgsP) {
		if len(args) == 1 {
			for _, spec := range types.CompletionRegistry {
				fmt.Fprintln(stdout_writer, ce.printSpecs(command.Cmd, spec))
			}
		}

		target_command := args[1]
		spec, exists := types.CompletionRegistry[target_command]
		if !exists {
			fmt.Fprintf(stderr_writer, "complete: %s: no completion specification\n", target_command)
			return
		}
		fmt.Fprintln(stdout_writer, ce.printSpecs(command.Cmd, spec))
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

func (ce *CommandExecutor) handleJobs(stdout_writer io.Writer) {
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

func (ce *CommandExecutor) handleHistory(stdout_writer io.Writer, command *parser.Command) {
	histories := types.History
	if len(command.Args) >= 1 {
		history_cmd_arg := command.Args[0]
		if history_cmd_arg == string(types.HistoryCommandArgsR) {
			if len(command.Args) < 2 {
				return
			}
			LoadHistory(command.Args[1])
			return
		} else if history_cmd_arg == string(types.HistoryCommandArgsW) {
			if len(command.Args) < 2 {
				return
			}
			data := strings.Join(types.History, "\n") + "\n"
			WriteHistory(command.Args[1], data)
			return
		} else if history_cmd_arg == string(types.HistoryCommandArgsA) {
			if len(command.Args) < 2 {
				return
			}
			AppendHistory(command.Args[1])
			return
		}
		limit, err := strconv.Atoi(history_cmd_arg)
		if err != nil {
			return
		}
		start := max(len(types.History)-limit, 0)
		histories = types.History[start:]
	}
	offset := len(types.History) - len(histories)
	for i, history := range histories {
		fmt.Fprintf(stdout_writer, "%5d  %s\n", offset+i+1, history)
	}
}

func (ce *CommandExecutor) handleDeclare(stdout_writer, stderr_writer io.Writer, command *parser.Command) {
	if len(command.Args) == 0 {
		return
	}
	if command.Args[0] == string(types.DeclareCommandArgsP) {
		if len(command.Args) < 2 {
			return
		}
		name := command.Args[1]
		value, ok := types.ShellVariables[name]
		if !ok {
			fmt.Fprintf(stderr_writer, "%s: %s: not found\n", command.Cmd, name)
			return
		}
		formatted_output := fmt.Sprintf("%s -- %s=%q", command.Cmd, name, value)
		fmt.Fprintf(stdout_writer, "%s\n", formatted_output)
		return
	}
	parts := strings.SplitN(command.Args[0], "=", 2)
	if len(parts) == 1 {
		return
	}
	if !utils.IsValidIdentifier(parts[0]) {
		fmt.Fprintf(stderr_writer, "%s: `%s': not a valid identifier\n", command.Cmd, command.Args[0])
		return
	}
	types.ShellVariables[parts[0]] = parts[1]
}

func (ce *CommandExecutor) runBackgroudJobs(command *parser.Command) (int, int, error) {
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
		JobNumber: ce.nextJobNumber(),
		Pid:       cmd.Process.Pid,
		Command:   command.Cmd + " " + strings.Join(args_without_amp, " "),
		Status:    string(types.JobStatusRunning),
	}

	types.Jobs = append(types.Jobs, job)

	return job.JobNumber, cmd.Process.Pid, nil
}

func (ce *CommandExecutor) nextJobNumber() int {
	if len(types.Jobs) == 0 {
		return 1
	}
	max := 0
	for _, job := range types.Jobs {
		if job.JobNumber > max {
			max = job.JobNumber
		}
	}
	return max + 1
}

func (ce *CommandExecutor) printSpecs(cmd string, spec types.CommandCompletionSpec) string {
	out := cmd
	if spec.CommandName != "" {
		out += " -C '" + spec.CommandName + "'"
	}
	out += " " + spec.TargetCommand

	return out
}

func (ce *CommandExecutor) printJobDetail(job_number, pid int) string {
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

func LoadHistory(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.Trim(line, "\r")
		if line == "" {
			continue
		}
		types.History = append(types.History, line)
	}
}

func WriteHistory(path, data string) {
	err := os.WriteFile(path, []byte(data), 0644)
	if err != nil {
		return
	}
}

func AppendHistory(path string) {
	if path == "" {
		return
	}
	new_lines := types.History[types.HistoryAppendIndex:]
	if len(new_lines) == 0 {
		return
	}
	data := strings.Join(new_lines, "\n") + "\n"
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return
	}
	_, err = f.WriteString(data)
	f.Close()
	if err != nil {
		return
	}
	types.HistoryAppendIndex = len(types.History)
}

func expandWord(word string) string {
	var sb strings.Builder
	for i := 0; i < len(word); i++ {
		if word[i] != '$' {
			sb.WriteByte(word[i])
			continue
		}
		j := i + 1
		for j < len(word) && utils.IsValidIdentifier(word[i+1:j+1]) {
			j++
		}
		name := word[i+1 : j]
		if name == "" {
			sb.WriteByte('$')
			continue
		}
		if value, ok := types.ShellVariables[name]; ok {
			sb.WriteString(value)
		}
		i = j - 1
	}
	return sb.String()
}

func ExpandCommand(command *parser.Command) {
	if command == nil {
		return
	}
	command.Cmd = expandWord(command.Cmd)
	for i, arg := range command.Args {
		command.Args[i] = expandWord(arg)
	}
}
