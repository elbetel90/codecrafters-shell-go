package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
)

// Ensures gofmt doesn't remove the "fmt" import in stage 1 (feel free to remove this!)
var _ = fmt.Print

var built_ins = []string{
	"exit",
	"echo",
	"type",
	"pwd",
	"cd",
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
)

type stack []rune

func (s *stack) Push(r rune) {
	*s = append(*s, r)
}

func (s *stack) Pop() rune {
	top := (*s)[len(*s)-1]
	*s = (*s)[:len(*s)-1]
	return top
}

func (s stack) Top() rune {
	return s[len(s)-1]
}

type Command struct {
	Cmd        string
	Args       []string
	OutputFile string
	ErrorFile  string
}

func parseCommand(input string) *Command {
	var tokens []string
	has_token := false
	var sb strings.Builder
	is_redirect_target := false
	stdout_file := ""
	stderr_file := ""

	redirect_mode := RedirectTypeNone

	st := stack{rune(StateTransitionNormal)}

	for _, ch := range input {
		current_mode := st.Top()
		switch current_mode {
		case rune(StateTransitionNormal):
			switch ch {
			case '\'':
				st.Push(rune(StateTransitionSingleQoute))
				has_token = true
			case '"':
				st.Push(rune(StateTranstionDoubleQoute))
				has_token = true
			case ' ', '\t':
				if has_token {
					current_token := sb.String()
					if is_redirect_target {
						switch redirect_mode {
						case RedirectTypeStderr:
							stderr_file = current_token
						case RedirectTypeStdout:
							stdout_file = current_token
						}
						is_redirect_target = false
						redirect_mode = RedirectTypeNone
					} else {
						tokens = append(tokens, current_token)
					}
					sb.Reset()
					has_token = false
				}
			case '\\':
				st.Push(rune(StateTransitionEscapeOutside))
				has_token = true
			case '>':
				buf := sb.String()
				switch buf {
				case "1":
					redirect_mode = RedirectTypeStdout
				case "2":
					redirect_mode = RedirectTypeStderr
				default:
					if len(buf) > 0 {
						tokens = append(tokens, buf)
					}
					redirect_mode = RedirectTypeStdout
				}
				sb.Reset()
				has_token = false
				is_redirect_target = true
			default:
				sb.WriteRune(ch)
				has_token = true
			}
		case rune(StateTransitionSingleQoute):
			if ch == '\'' {
				st.Pop()
			} else {
				sb.WriteRune(ch)
			}
		case rune(StateTranstionDoubleQoute):
			switch ch {
			case '"':
				st.Pop()
			case '\\':
				st.Push(rune(StateTransitionEscapeDoubleQoute))
			default:
				sb.WriteRune(ch)
			}
		case rune(StateTransitionEscapeDoubleQoute):
			switch ch {
			case '"', '\\':
				sb.WriteRune(ch)
			default:
				sb.WriteRune('\\')
				sb.WriteRune(ch)
			}
			st.Pop()
		case rune(StateTransitionEscapeOutside):
			sb.WriteRune(ch)
			st.Pop()
			has_token = true
		}

	}

	if st.Top() != 0 {
		// return "", nil, fmt.Errorf("syntax error: unclosed single quote")
	}

	if has_token {
		current_token := sb.String()
		if is_redirect_target {
			switch redirect_mode {
			case RedirectTypeStderr:
				stderr_file = current_token
			case RedirectTypeStdout:
				stdout_file = current_token
			}
			is_redirect_target = false
		} else {
			tokens = append(tokens, current_token)
		}
	}

	if is_redirect_target {
		// we will print here later
	}

	if len(tokens) == 0 {
		return nil
	}

	return &Command{Cmd: tokens[0], Args: tokens[1:], OutputFile: stdout_file, ErrorFile: stderr_file}
}

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
	if slices.Contains(built_ins, args[0]) {
		fmt.Println(args[0] + " is a shell builtin")
	} else if path, err := exec.LookPath(args[0]); err == nil {
		fmt.Println(args[0] + " is " + path)
	} else if args[0] == "type" {
		fmt.Println(args[0] + ": not found")
	} else {
		fmt.Println(args[0] + ": not found")
	}
}

func handleEcho(args []string, outputFile string) {
	output := strings.Join(args, " ") + "\n"

	if outputFile != "" {
		outFile, err := openFile(outputFile)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return
		}
		defer outFile.Close()

		outFile.WriteString(output)
	} else {
		fmt.Print(output)
	}
}

// file setup
func openFile(path string) (*os.File, error) {
	if path == "" {
		return nil, nil
	}

	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return nil, fmt.Errorf("failed to open output file: %w", err)
	}

	return file, err
}

func main() {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("$ ")
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error reading input: ", err)
			os.Exit(1)
		}

		input = strings.TrimRight(input, "\r\n")
		if len(input) == 0 {
			continue
		}

		command := parseCommand(input)

		switch command.Cmd {
		case "exit":
			os.Exit(0)
		case "echo":
			handleEcho(command.Args, command.OutputFile)
		case "pwd":
			handlePwd()
		case "cd":
			handleCd(command.Args)
		case "type":
			handleType(command.Args)
		default:
			if _, err := exec.LookPath(command.Cmd); err == nil {
				cmd := exec.Command(command.Cmd, command.Args...)
				cmd.Stdin = os.Stdin
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr

				if command.OutputFile != "" {
					stdout_file, err := openFile(command.OutputFile)
					if err != nil {
						// fmt.Fprintln(os.Stderr, err)
						return
					}
					defer stdout_file.Close()
					cmd.Stdout = stdout_file
				}

				if command.ErrorFile != "" {
					stderr_file, err := openFile(command.ErrorFile)
					if err != nil {
						// fmt.Fprintln(os.Stderr, err)
						return
					}
					defer stderr_file.Close()
					cmd.Stderr = stderr_file
				}

				if err := cmd.Run(); err != nil {
					// fmt.Printf("%s: command failed: %v\n", command, err)
				}
			} else {
				fmt.Println(input + ": command not found")

			}
		}
	}

}
