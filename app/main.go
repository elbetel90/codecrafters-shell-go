package main

import (
	"bufio"
	"errors"
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
	"cat",
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

func handleSingleQuote(arg string) (string, error) {
	var sb strings.Builder
	inSingleQoute := false

	for i := 0; i < len(arg); i++ {
		switch arg[i] {
		case '\'':
			inSingleQoute = !inSingleQoute
		case ' ', '\t':
			if inSingleQoute {
				sb.WriteByte(arg[i])
			}
		default:
			sb.WriteByte(arg[i])
		}
	}

	if inSingleQoute {
		return "", errors.New("syntax error: unclosed single quote")
	}

	return sb.String(), nil
}

func handleCat(args []string) {
	var newArgs []string
	for _, arg := range args {
		content, err := handleSingleQuote(arg)
		if err == nil {
			newArgs = append(newArgs, content)
		}
	}
	cmd := exec.Command("cat", newArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
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
		input = strings.TrimSpace(input)
		args := strings.Split(input, " ")
		command, args := args[0], args[1:]

		switch command {
		case "exit":
			os.Exit(0)
		case "echo":
			if strings.HasPrefix(args[0], "'") {
				content, err := handleSingleQuote(strings.Join(args, " "))
				if err == nil {
					fmt.Println(content)
				} else {
					fmt.Println(err)
				}
			} else {
				fmt.Println(input[5:])
			}
		case "pwd":
			handlePwd()
		case "cd":
			handleCd(args)
		case "type":
			handleType(args)
		case "cat":
			handleCat(args)

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
