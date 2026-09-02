package main

import (
	"bufio"
	"fmt"
	"os"
	"slices"
	"strings"
)

// Ensures gofmt doesn't remove the "fmt" import in stage 1 (feel free to remove this!)
var _ = fmt.Print

func main() {
	// TODO: Uncomment the code below to pass the first stage
	reader := bufio.NewReader(os.Stdin)
	var built_ins = []string{
		"exit", "echo", "type",
	}

	for {
		fmt.Print("$ ")
		command, err := reader.ReadString('\n')
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error reading input: ", err)
			os.Exit(1)
		}
		command = strings.TrimSpace(command)
		args := strings.Split(command, " ")
		if args[0] == "type" && slices.Contains(built_ins, args[1]) {
			fmt.Println(args[1] + " is a shell command")
		} else if args[0] == "type" {
			fmt.Println(args[1] + ": command not found")
		} else if command == "exit" {
			break
		} else if strings.HasPrefix(command, "echo ") {
			fmt.Println(command[5:])
		} else {
			fmt.Println(command + ": command not found")
		}

	}

}
