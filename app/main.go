package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
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
}

func main() {
	// TODO: Uncomment the code below to pass the first stage
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
		if command == "type" {
			if slices.Contains(built_ins, args[0]) {
				fmt.Println(args[0] + " is a shell builtin")
			} else if path, err := exec.LookPath(args[0]); err == nil {
				fmt.Println(args[0] + " is " + path)
			} else if args[0] == "type" {
				fmt.Println(args[0] + ": not found")
			} else {
				fmt.Println(args[0] + ": not found")
			}
		} else if _, err := exec.LookPath(command); err == nil {
			cmd := exec.Command(command, args...)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Run()
		} else if command == "pwd" {
			dir, err := os.Getwd()
			if err != nil {
				fmt.Fprintln(os.Stderr, "Error getting current directory", err)
				os.Exit(1)
			}
			fmt.Println(dir)
		} else if input == "exit" {
			break
		} else if strings.HasPrefix(input, "echo ") {
			fmt.Println(input[5:])
		} else {
			fmt.Println(input + ": command not found")
		}

	}

}
