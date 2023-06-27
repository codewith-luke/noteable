package main

// A simple example that shows how to retrieve a value from a Bubble Tea
// program after the Bubble Tea has exited.

import (
	"fmt"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"os"
	"os/exec"
	"strings"
	time "time"
)

type View string

type (
	errMsg error
)

const (
	COMMAND_VIEW View = "COMMAND_VIEW"
	CHAT_VIEW    View = "QUERY_VIEW"
)

var choices = []string{"Query", "Update"}

type model struct {
	activeView View

	viewport    viewport.Model
	messages    []string
	textarea    textarea.Model
	senderStyle lipgloss.Style

	cursor int
	choice string
	err    errMsg
}

func (m model) Init() tea.Cmd {
	// TODO: Figure out why this mofo don't blink
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
					m.activeView = CHAT_VIEW
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
		var (
			tiCmd tea.Cmd
			vpCmd tea.Cmd
		)

		m.textarea, tiCmd = m.textarea.Update(msg)
		m.viewport, vpCmd = m.viewport.Update(msg)

		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.Type {
			case tea.KeyCtrlC, tea.KeyEsc:
				fmt.Println(m.textarea.Value())
				return m, tea.Quit
			case tea.KeyEnter:
				m.messages = append(m.messages, m.senderStyle.Render("You: ")+m.textarea.Value())
				m.viewport.SetContent(strings.Join(m.messages, "\n"))

				ch := make(chan string)

				go (func(msg string) {
					cmd := exec.Command("./llm/run.sh", msg)
					stdout, err := cmd.CombinedOutput()

					if err != nil {
						fmt.Println(err.Error())
					}

					ch <- string(stdout)
				})(m.textarea.Value())

				m.viewport.GotoBottom()
				m.textarea.Reset()

				select {
				case <-time.After(5 * time.Second):
					return m, tea.Batch(tiCmd, vpCmd)
				case res := <-ch:
					m.messages = append(m.messages, m.senderStyle.Render("Sys: ")+res)
					m.viewport.SetContent(strings.Join(m.messages, "\n"))

					return m, tea.Batch(tiCmd, vpCmd)
				}
			}

		case errMsg:
			m.err = msg
			return m, nil
		}

		return m, tea.Batch(tiCmd, vpCmd)
	}

	return m, cmd
}

func (m model) View() string {
	switch m.activeView {
	case COMMAND_VIEW:
		view := renderCommandView(m)
		return view.String()
	case CHAT_VIEW:
		return renderChatView(m)
	default:
		return "unknown view"
	}
}

func initialModel() model {
	ta := textarea.New()
	ta.Placeholder = "Send a message..."
	ta.Focus()

	ta.Prompt = "┃ "
	ta.CharLimit = 280

	ta.SetWidth(20)
	ta.SetHeight(3)

	// Remove cursor line styling
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()

	ta.ShowLineNumbers = false

	vp := viewport.New(100, 20)
	vp.SetContent(`Welcome to the chat room!
Type a message and press Enter to send.`)

	ta.KeyMap.InsertNewline.SetEnabled(false)

	return model{
		textarea:    ta,
		messages:    []string{},
		viewport:    vp,
		senderStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("5")),
		activeView:  COMMAND_VIEW,
		err:         nil,
	}
}

func main() {
	appModel := initialModel()
	p := tea.NewProgram(appModel)

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

func renderCommandView(m model) strings.Builder {
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

func renderChatView(m model) string {
	return fmt.Sprintf(
		"%s\n\n%s",
		m.viewport.View(),
		m.textarea.View(),
	) + "\n\n"
}
