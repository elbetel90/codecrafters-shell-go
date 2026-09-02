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

	var builtIns = []string{
		"exit", "echo", "type",
	}

	for {
		fmt.Print("$ ")
		command, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error reading input: ", err)
			os.Exit(1)
		}
		command = strings.TrimSpace(command)
		if command == "exit" {
			break
		}
		if strings.HasPrefix(command, "echo ") {
			fmt.Println(command[5:])
		}
		if strings.HasPrefix(command, "type ") {
			types := command[5:]
			if slices.Contains(builtIns, types) {
				fmt.Println(types + " is a shell builtin")
			} else {
				fmt.Println(command + ": not found")
			}

		} else {
			fmt.Println(command + ": command not found")
		}

	}

}
