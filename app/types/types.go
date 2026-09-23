package types

type BuiltinCommand string

const (
	BuiltinCommandExit     BuiltinCommand = "exit"
	BuiltinCommandEcho     BuiltinCommand = "echo"
	BuiltinCommandType     BuiltinCommand = "type"
	BuiltinCommandPwd      BuiltinCommand = "pwd"
	BuiltinCommandCd       BuiltinCommand = "cd"
	BuiltinCommandComplete BuiltinCommand = "complete"
	BuiltinCommandJobs     BuiltinCommand = "jobs"
)

var Built_ins = []string{
	string(BuiltinCommandExit),
	string(BuiltinCommandEcho),
	string(BuiltinCommandType),
	string(BuiltinCommandPwd),
	string(BuiltinCommandCd),
	string(BuiltinCommandComplete),
	string(BuiltinCommandJobs),
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
var NextJobNumber int = 1
