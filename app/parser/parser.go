package parser

import (
	"fmt"
	"strings"

	"github.com/codecrafters-io/shell-starter-go/app/types"
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
	Cmd          string
	Args         []string
	OutputFile   string
	ErrorFile    string
	AppendStdout bool
	AppendStderr bool
}

func NewCommand() *Command {
	return &Command{}
}

func (c *Command) ParseCommand(input string) *Command {
	var sb strings.Builder

	var tokens []string

	has_token := false
	is_redirect_target := false
	is_append_out := false
	is_append_err := false

	stdout_file := ""
	stderr_file := ""

	redirect_mode := types.RedirectTypeNone

	st := stack{rune(types.StateTransitionNormal)}

	for i := 0; i < len(input); i++ {
		ch := rune(input[i])
		current_mode := st.Top()
		switch current_mode {
		case rune(types.StateTransitionNormal):
			switch ch {
			case '\'':
				st.Push(rune(types.StateTransitionSingleQoute))
				has_token = true
			case '"':
				st.Push(rune(types.StateTranstionDoubleQoute))
				has_token = true
			case ' ', '\t':
				if ch == '\t' {
					fmt.Println("called here")
				}
				if has_token {
					current_token := sb.String()
					if is_redirect_target {
						switch redirect_mode {
						case types.RedirectTypeStderr:
							stderr_file = current_token
							is_append_err = false
							is_append_out = false
						case types.RedirectTypeStdout:
							stdout_file = current_token
							is_append_err = false
							is_append_out = false
						case types.RedirectTypeAppendStdout:
							stdout_file = current_token
							is_append_err = false
							is_append_out = true
						case types.RedirectTypeAppendStderr:
							stderr_file = current_token
							is_append_err = true
							is_append_out = false
						}
						is_redirect_target = false
						redirect_mode = types.RedirectTypeNone
					} else {
						tokens = append(tokens, current_token)
					}
					sb.Reset()
					has_token = false
				}
			case '\\':
				st.Push(rune(types.StateTransitionEscapeOutside))
				has_token = true
			case '>':
				is_append := false
				if i+1 < len(input) && input[i+1] == '>' {
					is_append = true
					i++
				}
				buf := sb.String()
				switch buf {
				case "1":
					if is_append {
						redirect_mode = types.RedirectTypeAppendStdout
						is_append_out = true
						is_append_err = false
					} else {
						redirect_mode = types.RedirectTypeStdout
						is_append_out = false
						is_append_err = false
					}
				case "2":
					if is_append {
						redirect_mode = types.RedirectTypeAppendStderr
						is_append_out = false
						is_append_err = true
					} else {
						redirect_mode = types.RedirectTypeStderr
						is_append_out = false
						is_append_err = false
					}
				default:
					if len(buf) > 0 {
						tokens = append(tokens, buf)
					}
					if is_append {
						redirect_mode = types.RedirectTypeAppendStdout
						is_append_out = true
						is_append_err = false
					} else {
						redirect_mode = types.RedirectTypeStdout
						is_append_out = false
						is_append_err = false
					}
				}
				sb.Reset()
				has_token = false
				is_redirect_target = true
			default:
				sb.WriteRune(ch)
				has_token = true
			}
		case rune(types.StateTransitionSingleQoute):
			if ch == '\'' {
				st.Pop()
			} else {
				sb.WriteRune(ch)
			}
		case rune(types.StateTranstionDoubleQoute):
			switch ch {
			case '"':
				st.Pop()
			case '\\':
				st.Push(rune(types.StateTransitionEscapeDoubleQoute))
			default:
				sb.WriteRune(ch)
			}
		case rune(types.StateTransitionEscapeDoubleQoute):
			switch ch {
			case '"', '\\':
				sb.WriteRune(ch)
			default:
				sb.WriteRune('\\')
				sb.WriteRune(ch)
			}
			st.Pop()
		case rune(types.StateTransitionEscapeOutside):
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
			case types.RedirectTypeStderr:
				stderr_file = current_token
				is_append_err = false
				is_append_out = false
			case types.RedirectTypeStdout:
				stdout_file = current_token
				is_append_err = false
				is_append_out = false
			case types.RedirectTypeAppendStdout:
				stdout_file = current_token
				is_append_err = false
				is_append_out = true
			case types.RedirectTypeAppendStderr:
				stderr_file = current_token
				is_append_err = true
				is_append_out = false
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

	return &Command{
		Cmd:          tokens[0],
		Args:         tokens[1:],
		OutputFile:   stdout_file,
		ErrorFile:    stderr_file,
		AppendStdout: is_append_out,
		AppendStderr: is_append_err,
	}
}
