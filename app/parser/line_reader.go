package parser

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"

	"github.com/codecrafters-io/shell-starter-go/app/types"
	"github.com/codecrafters-io/shell-starter-go/app/utils"
	"golang.org/x/term"
)

func GetAllCommands() []string {
	var commands []string
	seen := make(map[string]bool)

	for _, b := range types.Built_ins {
		if !seen[b] {
			seen[b] = true
			commands = append(commands, b)
		}
	}

	env_path := os.Getenv(string(types.EnvVarPath))
	exec_dirs := filepath.SplitList(env_path)

	for _, exec_dir := range exec_dirs {
		dir_entries, err := os.ReadDir(exec_dir)
		if err != nil {
			continue
		}

		for _, dir_entry := range dir_entries {
			if dir_entry.IsDir() {
				continue
			}
			dir_info, err := dir_entry.Info()
			if err != nil {
				continue
			}
			perm := dir_info.Mode().Perm()
			if perm&0111 != 0 {
				name := dir_entry.Name()
				if !seen[name] {
					seen[name] = true
					commands = append(commands, name)
				}
			}

		}
	}
	slices.Sort(commands)
	return commands
}

func handleTabCompletion(line_byte *[]byte, all_commands []string, last_was_tab *bool) error {
	line := string(*line_byte)

	if !strings.Contains(line, " ") {
		var matches []string
		for _, cmd := range all_commands {
			if strings.HasPrefix(cmd, line) {
				matches = append(matches, cmd)
			}
		}
		sort.Strings(matches)

		switch len(matches) {
		case 0:
			os.Stdout.WriteString("\x07")
		case 1:
			match := matches[0]
			current_len := len(*line_byte)
			if current_len < len(match) {
				suffix := match[current_len:] + " "
				os.Stdout.WriteString(suffix)
				*line_byte = append(*line_byte, []byte(suffix)...)
			} else if current_len == len(match) {
				os.Stdout.WriteString(" ")
				*line_byte = append(*line_byte, ' ')
			}
		default:
			lcp := utils.LongestCommandPrefix(matches)
			current_len := len(*line_byte)

			if len(lcp) > len(line) {
				suffix := lcp[current_len:]
				os.Stdout.WriteString(suffix)
				*line_byte = append(*line_byte, []byte(suffix)...)
			} else {
				if !*last_was_tab {
					os.Stdout.WriteString("\x07")
					*last_was_tab = true
				} else {
					os.Stdout.WriteString("\r\n")
					for i, match := range matches {
						os.Stdout.WriteString(match)
						if i < len(matches)-1 {
							os.Stdout.WriteString(" ")
						}
					}
					os.Stdout.WriteString("\r\n$ " + string(*line_byte))
					*last_was_tab = false
				}
			}
		}
	}

	return nil
}

func ReadLine(all_commands []string) (string, error) {
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return "", err
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	var line_byte []byte
	buf := make([]byte, 1)
	last_was_tab := false

	for {
		_, err := os.Stdin.Read(buf)
		if err != nil {
			return "", err
		}
		b := buf[0]
		if b != '\t' {
			last_was_tab = false
		}

		switch b {
		case '\r', '\n':
			os.Stdout.WriteString("\r\n")
			return string(line_byte), nil
		case '\t':
			line := string(line_byte)
			if len(line) > 0 && line[len(line)-1] == ' ' {
				os.Stdout.WriteString("\x07")
				continue
			}
			err := handleTabCompletion(&line_byte, all_commands, &last_was_tab)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				return "", err
			}
		case '\x7f', '\b':
			if len(line_byte) > 0 {
				line_byte = line_byte[:len(line_byte)-1]
				os.Stdout.WriteString("\b \b")
			} else {
				os.Stdout.WriteString("\x07")
			}
		case '\x1b':
			seq := make([]byte, 2)
			n, _ := os.Stdin.Read(seq)
			if n == 2 && seq[0] == '[' {
				switch seq[1] {
				case 'A', 'B', 'C', 'D':
					// ignore for now
				}
			}
		case '\x03':
			os.Stdout.WriteString("\r\n")
			return "", nil
		default:
			if b >= 32 && b <= 126 {
				line_byte = append(line_byte, byte(b))
				os.Stdout.Write([]byte{b})
			}
		}
	}

}
