package main

// A simple example that shows how to retrieve a value from a Bubble Tea
// program after the Bubble Tea has exited.

import (
	"bufio"
	"fmt"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"os"
	"os/exec"
	"strings"
)

type View string

type (
	errMsg error
)

const (
	CommandView View = "COMMAND_VIEW"
	ChatView    View = "QUERY_VIEW"

	TriggerQuestion = "trigger_question"
)

var choices = []string{"Query", "Update"}

type model struct {
	activeView View

	viewport    viewport.Model
	messages    []string
	textarea    textarea.Model
	senderStyle lipgloss.Style
	question    string

	cursor int
	choice string
	err    errMsg

	debugger *os.File
}

func (m model) debug(msg string) {
	message := fmt.Sprintf("DEBUG: %s\n", msg)
	m.debugger.WriteString(message)
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd   tea.Cmd
		tiCmd tea.Cmd
		vpCmd tea.Cmd
	)

	m.textarea, tiCmd = m.textarea.Update(msg)
	m.viewport, vpCmd = m.viewport.Update(msg)

	if msg == TriggerQuestion {
		m.debug(m.question)
		out := make(chan string)
		done := make(chan bool)

		go (func() {
			m.debug("running script")
			runScript(m.question, out, done)
		})()

		select {
		case <-done:
			m.debug("done")
			m.question = ""
			return m, tea.Batch(tiCmd, vpCmd, cmd)
		case outM := <-out:
			m.debug(outM)
			m.messages = append(m.messages, m.senderStyle.Render("Sys:")+outM)
			m.viewport.SetContent(strings.Join(m.messages, "\n"))
			m.textarea.Reset()
			m.viewport.GotoBottom()
		}
	}

	if m.activeView == CommandView {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "ctrl+c", "q", "esc":
				return m, tea.Quit

			case "enter":
				// Send the choice on the channel and exit.
				m.choice = choices[m.cursor]
				if m.choice == "Query" {
					m.activeView = ChatView
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
			switch msg.Type {
			case tea.KeyCtrlC, tea.KeyEsc:
				return m, tea.Quit
			case tea.KeyEnter:
				userMsg := m.textarea.Value()
				m.messages = append(m.messages, m.senderStyle.Render("You: ")+userMsg)
				m.messages = append(m.messages, m.senderStyle.Render("Sys: ")+"I'm thinking...")
				m.viewport.SetContent(strings.Join(m.messages, "\n"))
				m.textarea.Reset()
				m.viewport.GotoBottom()
				m.question = userMsg
				return m, tea.Batch(tiCmd, vpCmd, func() tea.Msg {
					return TriggerQuestion
				})
			}

		case errMsg:
			m.err = msg
			return m, nil
		}
	}

	return m, cmd
}

func (m model) View() string {
	switch m.activeView {
	case CommandView:
		view := renderCommandView(m)
		return view.String()
	case ChatView:
		return renderChatView(m)
	default:
		return "unknown view"
	}
}

func initialModel() model {
	f, err := tea.LogToFile("debug.log", "debug")
	f.Truncate(0)
	if err != nil {
		fmt.Println("fatal:", err)
		os.Exit(1)
	}

	ta := textarea.New()
	ta.Placeholder = "Send a message..."
	ta.Focus()

	ta.Prompt = "┃ "
	ta.CharLimit = 280

	ta.SetWidth(50)
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
		activeView:  CommandView,
		err:         nil,
		debugger:    f,
	}
}

func main() {
	appModel := initialModel()
	p := tea.NewProgram(appModel)

	defer appModel.debugger.Close()
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

func runScript(msg string, out chan string, done chan bool) {
	cmd := exec.Command("./llm/run.sh", msg)
	cmdReader, _ := cmd.StdoutPipe()
	scanner := bufio.NewScanner(cmdReader)

	cmd.Start()
	for scanner.Scan() {
		out <- scanner.Text()
	}

	err := cmd.Wait()

	done <- true

	if err != nil {
		fmt.Printf("cmd.Run() failed with %s\n", err)
	}
}
