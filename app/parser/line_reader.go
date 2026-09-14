package parser

import (
	"os"
	"strings"

	"github.com/codecrafters-io/shell-starter-go/app/types"
	"golang.org/x/term"
)

func ReadLine() (string, error) {
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return "", err
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	var line_byte []byte
	buf := make([]byte, 1)

	for {
		_, err := os.Stdin.Read(buf)
		if err != nil {
			return "", err
		}
		b := buf[0]

		switch b {
		case '\r', '\n':
			os.Stdout.WriteString("\r\n")
			return string(line_byte), nil
		case '\t':
			line := string(line_byte)
			if !strings.Contains(line, " ") {
				var matches []string
				for _, b_name := range types.Built_ins {
					if strings.HasPrefix(b_name, line) {
						matches = append(matches, b_name)
					}
				}

				if len(matches) == 1 {
					completion := matches[0][len(line):] + " "
					os.Stdout.WriteString(completion)
					line_byte = append(line_byte, completion...)
				} else if len(matches) == 0 {
					os.Stdout.WriteString("\x07")
				}
			}
		case '\x7f', '\b':
			if len(line_byte) > 0 {
				line_byte = line_byte[:len(line_byte)-1]
				os.Stdout.WriteString("\b \b")
			}
		case '\x1b':
			seq := make([]byte, 2)
			n, _ := os.Stdin.Read(seq)
			if n == 2 && seq[0] == '[' {
				switch seq[1] {
				case 'A', 'B':
					// ignore for now
				case 'C', 'D':
					// ignore for now
				}
			}
		case '\x03':
			os.Stdout.WriteString("\r\n")
			return "", nil
		default:
			if b >= 32 && b <= 126 {
				line_byte = append(line_byte, byte(b))
			}
			os.Stdout.Write(buf)
		}
	}

}
