package main

// A simple example that shows how to retrieve a value from a Bubble Tea
// program after the Bubble Tea has exited.

import (
	"fmt"
	"github.com/charmbracelet/bubbles/textinput"
	"os"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type View string

type (
	errMsg error
)

const (
	COMMAND_VIEW View = "COMMAND_VIEW"
	QUERY_VIEW   View = "QUERY_VIEW"
)

var choices = []string{"Query", "Update"}

type model struct {
	activeView View

	textInput textinput.Model
	cursor    int
	choice    string
	err       errMsg
}

func (m *model) Init() tea.Cmd {
	// TODO: Figure out why this mofo don't blink
	return textinput.Blink
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	if m.activeView == COMMAND_VIEW {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "ctrl+c", "q", "esc":
				return m, tea.Quit

			case "enter":
				// Send the choice on the channel and exit.
				m.choice = choices[m.cursor]
				if m.choice == "Query" {
					m.activeView = QUERY_VIEW
				}

			case "down", "j":
				m.cursor++
				if m.cursor >= len(choices) {
					m.cursor = 0
				}

			case "up", "k":
				m.cursor--
				if m.cursor < 0 {
					m.cursor = len(choices) - 1
				}
			}
		case errMsg:
			m.err = msg
			return m, nil
		}
	} else {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "ctrl+c", "q", "esc":
				return m, tea.Quit

			case "enter":
				cmd := exec.Command("./llm/run.sh", m.textInput.Value())
				stdout, err := cmd.Output()

				if err != nil {
					fmt.Println(err.Error())
					return m, tea.Quit
				}

				// Print the output
				fmt.Println(string(stdout))
			}
		case errMsg:
			m.err = msg
			return m, nil
		}
	}

	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m *model) View() string {
	switch m.activeView {
	case COMMAND_VIEW:
		view := renderCommandView(m)
		return view.String()
	case QUERY_VIEW:
		return renderQueryView(m)
	default:
		return "unknown view"
	}
}

func main() {
	ti := textinput.New()

	p := tea.NewProgram(&model{
		textInput:  ti,
		activeView: QUERY_VIEW,
	})

	// Run returns the model as a tea.Model.
	m, err := p.Run()
	if err != nil {
		fmt.Println("Oh no:", err)
		os.Exit(1)
	}

	// Assert the final tea.Model to our local model and print the choice.
	if m, ok := m.(*model); ok && m.choice != "" {
		fmt.Println("\n---\nExited\n")
	}
}

func renderCommandView(m *model) strings.Builder {
	s := strings.Builder{}
	s.WriteString("What kind of Bubble Tea would you like to order?\n\n")

	for i := 0; i < len(choices); i++ {
		if m.cursor == i {
			s.WriteString("(•) ")
		} else {
			s.WriteString("( ) ")
		}
		s.WriteString(choices[i])
		s.WriteString("\n")
	}

	s.WriteString("\n(press q to quit)\n")
	return s
}

func renderQueryView(m *model) string {
	m.textInput.CharLimit = 100
	m.textInput.Width = 20
	m.textInput.Focus()
	m.textInput.Placeholder = "How do I make it blink?"

	return fmt.Sprintf(
		"What is your query?\n\n%s\n\n%s",
		m.textInput.View(),
		"(esc to quit)",
	) + "\n"
}
