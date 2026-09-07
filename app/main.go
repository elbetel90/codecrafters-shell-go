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
}

func parseCommand(input string) *Command {
	var tokens []string
	has_token := false
	var sb strings.Builder
	is_redirect_target := false
	last_was_unqouted_one := false
	output_file := ""

	st := stack{0}

	for _, ch := range input {
		current_mode := st.Top()
		switch current_mode {
		case rune(StateTransitionNormal):
			switch ch {
			case '\'':
				st.Push(rune(StateTransitionSingleQoute))
				has_token = true
				last_was_unqouted_one = false
			case '"':
				st.Push(rune(StateTranstionDoubleQoute))
				has_token = true
				last_was_unqouted_one = false
			case ' ', '\t':
				if has_token {
					current_token := sb.String()
					if is_redirect_target {
						output_file = current_token
						is_redirect_target = false
					} else {
						tokens = append(tokens, current_token)
					}
					sb.Reset()
					has_token = false
				}
				last_was_unqouted_one = false
			case '\\':
				st.Push(rune(StateTransitionEscapeOutside))
				has_token = true
				last_was_unqouted_one = false
			case '>':
				if last_was_unqouted_one {
					buf := sb.String()
					trimmed_buf := buf[:len(buf)-1]
					if len(trimmed_buf) > 0 {
						tokens = append(tokens, trimmed_buf)
					}
				} else if has_token {
					tokens = append(tokens, sb.String())
				}
				sb.Reset()
				has_token = false
				is_redirect_target = true
				last_was_unqouted_one = false
			default:
				sb.WriteRune(ch)
				has_token = true
				last_was_unqouted_one = (ch == '1')
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
			output_file = current_token
			is_redirect_target = false
		} else {
			tokens = append(tokens, current_token)
		}
	} else if is_redirect_target {

	}

	if len(tokens) == 0 {
		return nil
	}

	return &Command{Cmd: tokens[0], Args: tokens[1:], OutputFile: output_file}
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

// parsing, file setup, and process execution

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
			fmt.Println(strings.Join(command.Args, " "))
		case "pwd":
			handlePwd()
		case "cd":
			handleCd(command.Args)
		case "type":
			handleType(command.Args)
		default:
			if _, err := exec.LookPath(command.Cmd); err == nil {
				cmd := exec.Command(command.Cmd, command.Args...)
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				cmd.Run()
			} else {
				fmt.Println(input + ": command not found")

			}
		}
	}

}
