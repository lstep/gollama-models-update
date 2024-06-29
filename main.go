package main

import (
	"bufio"
	"fmt"
	"os/exec"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/lipgloss"
	tea "github.com/charmbracelet/bubbletea"
)

var (
	successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	errorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	infoStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("12"))
	boldStyle    = lipgloss.NewStyle().Bold(true)
)

type model struct {
	spinner  spinner.Model
	models   []string
	current  int
	done     bool
	failed   []string
	updating bool
}

func initialModel() model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	return model{spinner: s}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, func() tea.Msg {
		models, err := getInstalledModels()
		if err != nil {
			return errorStyle.Render(fmt.Sprintf("Error getting installed models: %v", err))
		}
		return models
	})
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "q" {
			return m, tea.Quit
		}
	case []string:
		m.models = msg
		return m, m.updateNext
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case string:
		if strings.HasPrefix(msg, "Error:") {
			m.failed = append(m.failed, m.models[m.current])
		}
		m.current++
		if m.current >= len(m.models) {
			m.done = true
			return m, tea.Quit
		}
		return m, m.updateNext
	}
	return m, nil
}

func (m model) View() string {
	if len(m.models) == 0 {
		return fmt.Sprintf("%s Fetching installed models...", m.spinner.View())
	}
	if m.done {
		var output strings.Builder
		output.WriteString(infoStyle.Render("All model update attempts completed.\n\n"))
		if len(m.failed) == 0 {
			output.WriteString(successStyle.Render("All models updated successfully!\n"))
		} else {
			output.WriteString(errorStyle.Render("The following models failed to update:\n"))
			for _, model := range m.failed {
				output.WriteString(errorStyle.Render(fmt.Sprintf("- %s\n", model)))
			}
			output.WriteString("\n" + infoStyle.Render("Please check these models manually.\n"))
		}
		return lipgloss.NewStyle().Margin(0, 2).Render(output.String())
	}
	return fmt.Sprintf("%s Updating %s... (%d/%d)",
		m.spinner.View(),
		boldStyle.Render(m.models[m.current]),
		m.current+1,
		len(m.models))
}

func (m model) updateNext() tea.Msg {
	model := m.models[m.current]
	cmd := exec.Command("ollama", "pull", model)
	output, err := cmd.CombinedOutput()
	if err != nil || strings.Contains(string(output), "Error: pull model manifest: file does not exist") {
		return fmt.Sprintf("Error: Failed to update %s", model)
	}
	return fmt.Sprintf("Updated %s successfully", model)
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
	}
}

func getInstalledModels() ([]string, error) {
	cmd := exec.Command("ollama", "ls")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var models []string
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	// Skip the header line
	scanner.Scan()
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) > 0 {
			models = append(models, fields[0])
		}
	}

	return models, nil
}
