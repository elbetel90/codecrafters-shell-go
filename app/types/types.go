package types

import "slices"

type BuiltinCommand string

const (
	BuiltinCommandExit     BuiltinCommand = "exit"
	BuiltinCommandEcho     BuiltinCommand = "echo"
	BuiltinCommandType     BuiltinCommand = "type"
	BuiltinCommandPwd      BuiltinCommand = "pwd"
	BuiltinCommandCd       BuiltinCommand = "cd"
	BuiltinCommandComplete BuiltinCommand = "complete"
	BuiltinCommandJobs     BuiltinCommand = "jobs"
	BuiltinCommandsHistory BuiltinCommand = "history"
	BuiltinCommandDeclare  BuiltinCommand = "declare"
)

var Built_ins = []string{
	string(BuiltinCommandExit),
	string(BuiltinCommandEcho),
	string(BuiltinCommandType),
	string(BuiltinCommandPwd),
	string(BuiltinCommandCd),
	string(BuiltinCommandComplete),
	string(BuiltinCommandJobs),
	string(BuiltinCommandsHistory),
	string(BuiltinCommandDeclare),
}

func IsBuiltin(name string) bool {
	return slices.Contains(Built_ins, name)
}

type StateTransition int

const (
	StateTransitionNormal StateTransition = iota
	StateTransitionSingleQoute
	StateTranstionDoubleQoute
	StateTransitionEscapeOutside
	StateTransitionEscapeDoubleQoute
)

type RedirectType int

const (
	RedirectTypeNone RedirectType = iota
	RedirectTypeStdout
	RedirectTypeStderr
	RedirectTypeAppendStdout
	RedirectTypeAppendStderr
)

type EnvVar string

const (
	EnvVarPath EnvVar = "PATH"
)

type CompleteCommandArgs string

const (
	CompleteCommandArgsP CompleteCommandArgs = "-p"
	CompleteCommandArgsC CompleteCommandArgs = "-C"
	CompleteCommandArgsR CompleteCommandArgs = "-r"
)

type HistoryCommandArgs string

const (
	HistoryCommandArgsR HistoryCommandArgs = "-r"
	HistoryCommandArgsW HistoryCommandArgs = "-w"
	HistoryCommandArgsA HistoryCommandArgs = "-a"
)

type DeclareCommandArgs string

const (
	DeclareCommandArgsP DeclareCommandArgs = "-p"
)

type CommandCompletionSpec struct {
	CommandName   string // -C flag
	TargetCommand string // target command
}

var CompletionRegistry = make(map[string]CommandCompletionSpec)

type Job struct {
	JobNumber int
	Pid       int
	Command   string
	Status    string
}

var Jobs []Job

type JobStatus string

const (
	JobStatusRunning JobStatus = "Running"
	JobStatusDone    JobStatus = "Done"
)

var History []string
var HistoryAppendIndex int

var ShellVariables = map[string]string{}
