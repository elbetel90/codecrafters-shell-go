package types

type BuiltinCommand string

const (
	BuiltinCommandExit     BuiltinCommand = "exit"
	BuiltinCommandEcho     BuiltinCommand = "echo"
	BuiltinCommandType     BuiltinCommand = "type"
	BuiltinCommandPwd      BuiltinCommand = "pwd"
	BuiltinCommandCd       BuiltinCommand = "cd"
	BuiltinCommandComplete BuiltinCommand = "complete"
)

var Built_ins = []string{
	string(BuiltinCommandExit),
	string(BuiltinCommandEcho),
	string(BuiltinCommandType),
	string(BuiltinCommandPwd),
	string(BuiltinCommandCd),
	string(BuiltinCommandComplete),
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
