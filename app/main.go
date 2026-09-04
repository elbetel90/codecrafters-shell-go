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

func parseCommand(input string) (cmd string, args []string) {
	var tokens []string
	hasToken := false
	escapeNext := false
	var sb strings.Builder

	st := stack{0}

	for _, ch := range input {
		current_mode := st.Top()
		switch current_mode {
		case 0:
			switch ch {
			case '\'':
				st.Push('\'')
				hasToken = true
			case '"':
				st.Push('"')
				hasToken = true
			case ' ', '\t':
				if hasToken {
					tokens = append(tokens, sb.String())
					sb.Reset()
					hasToken = false
				}
			case '\\':
				st.Push('\\')
				hasToken = true
				escapeNext = true
			default:
				sb.WriteRune(ch)
				hasToken = true
			}
		case '\'':
			if escapeNext {
				continue
			}
			if ch == '\'' {
				st.Pop()
			} else {
				sb.WriteRune(ch)
			}
		case '"':
			if escapeNext {
				continue
			}
			if ch == '"' {
				st.Pop()
			} else {
				sb.WriteRune(ch)
			}
		case '\\':
			if ch == '\\' {
				escapeNext = false
				// st.Pop()
			} else {
				sb.WriteRune(ch)
			}
		}

	}

	if st.Top() != 0 {
		// return "", nil, fmt.Errorf("syntax error: unclosed single quote")
	}

	if hasToken {
		tokens = append(tokens, sb.String())
	}

	if len(tokens) == 0 {
		return "", nil
	}

	return tokens[0], tokens[1:]
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

		command, args := parseCommand(input)

		switch command {
		case "exit":
			os.Exit(0)
		case "echo":
			fmt.Println(strings.Join(args, " "))
		case "pwd":
			handlePwd()
		case "cd":
			handleCd(args)
		case "type":
			handleType(args)

		default:
			if _, err := exec.LookPath(command); err == nil {
				cmd := exec.Command(command, args...)
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				cmd.Run()
			} else {
				fmt.Println(input + ": command not found")

			}
		}
	}

}
