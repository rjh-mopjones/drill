package ui

import (
	"drill/cache"
	"drill/fetcher"
	"drill/mock"
	"drill/models"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/uuid"
)

type menuOption int

const (
	optionLoadAccount menuOption = iota
	optionMockMode
)

type EntryModel struct {
	menuSelection menuOption
	previousIndex int
	textInput     textinput.Model
	inputMode     bool
	cache         *cache.Cache
	services      []models.ServiceConfig
	width         int
	height        int
	err           error
	loading       bool
	loadingMsg    string
}

type LoadCompleteMsg struct {
	AggregateID string
	Events      []models.Event
	Commands    []models.Command
	IsMock      bool
}

type LoadErrorMsg struct {
	Err error
}

func NewEntryModel(services []models.ServiceConfig) EntryModel {
	ti := textinput.New()
	ti.Placeholder = "Enter Aggregate ID (UUID)"
	ti.CharLimit = 36
	ti.Width = 40

	c, _ := cache.Load()

	return EntryModel{
		menuSelection: optionLoadAccount,
		textInput:     ti,
		cache:         c,
		services:      services,
	}
}

func (m EntryModel) Init() tea.Cmd {
	return nil
}

func (m EntryModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.loading {
			return m, nil
		}

		if m.inputMode {
			switch msg.String() {
			case "enter":
				aggregateID := strings.TrimSpace(m.textInput.Value())
				if aggregateID == "" {
					m.err = fmt.Errorf("aggregate ID cannot be empty")
					return m, nil
				}
				if _, err := uuid.Parse(aggregateID); err != nil {
					m.err = fmt.Errorf("invalid UUID format")
					return m, nil
				}
				m.err = nil
				m.loading = true
				m.loadingMsg = "Loading data from services..."
				return m, m.loadFromServices(aggregateID)
			case "esc":
				m.inputMode = false
				m.textInput.Blur()
				return m, nil
			}
			var cmd tea.Cmd
			m.textInput, cmd = m.textInput.Update(msg)
			return m, cmd
		}

		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "up", "k":
			if m.menuSelection > 0 {
				m.menuSelection--
			}
		case "down", "j":
			if m.menuSelection < optionMockMode {
				m.menuSelection++
			}
		case "left", "h":
			m.previousIndex = -1 // Deselect previous
		case "right", "l":
			if len(m.cache.Requests) > 0 && m.previousIndex < 0 {
				m.previousIndex = 0
			}
		case "tab":
			if m.previousIndex >= 0 {
				m.previousIndex = -1
			} else if len(m.cache.Requests) > 0 {
				m.previousIndex = 0
			}
		case "enter":
			if m.previousIndex >= 0 && m.previousIndex < len(m.cache.Requests) {
				// Load from cache
				req := m.cache.Requests[m.previousIndex]
				return m, func() tea.Msg {
					return LoadCompleteMsg{
						AggregateID: req.AggregateID,
						Events:      req.Events,
						Commands:    req.Commands,
						IsMock:      req.IsMock,
					}
				}
			}
			switch m.menuSelection {
			case optionLoadAccount:
				m.inputMode = true
				m.textInput.Focus()
				return m, textinput.Blink
			case optionMockMode:
				m.loading = true
				m.loadingMsg = "Generating mock data..."
				return m, m.loadMockData()
			}
		}

		// Navigate previous requests when focused on right side
		if m.previousIndex >= 0 {
			switch msg.String() {
			case "up", "k":
				if m.previousIndex > 0 {
					m.previousIndex--
				}
			case "down", "j":
				if m.previousIndex < len(m.cache.Requests)-1 {
					m.previousIndex++
				}
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case LoadCompleteMsg:
		m.loading = false
		// Save to cache
		m.cache.AddRequest(msg.AggregateID, msg.Events, msg.Commands, msg.IsMock)
		m.cache.Save()
		// Return the data view model
		dataModel := NewModel(msg.AggregateID)
		dataModel.Commands = msg.Commands
		dataModel.Events = msg.Events
		dataModel.Loading = false
		dataModel.Services = m.services
		return dataModel, func() tea.Msg {
			return tea.WindowSizeMsg{Width: m.width, Height: m.height}
		}

	case LoadErrorMsg:
		m.loading = false
		m.err = msg.Err
	}

	return m, nil
}

func (m EntryModel) loadMockData() tea.Cmd {
	return func() tea.Msg {
		aggregateID := uuid.New().String()
		events, commands := mock.GenerateMockData(aggregateID)
		return LoadCompleteMsg{
			AggregateID: aggregateID,
			Events:      events,
			Commands:    commands,
			IsMock:      true,
		}
	}
}

func (m EntryModel) loadFromServices(aggregateID string) tea.Cmd {
	return func() tea.Msg {
		if len(m.services) == 0 {
			return LoadErrorMsg{Err: fmt.Errorf("no services configured. Use -services flag")}
		}

		f := fetcher.NewFetcher(m.services)
		events, commands, err := f.FetchAll(aggregateID)
		if err != nil {
			return LoadErrorMsg{Err: err}
		}

		return LoadCompleteMsg{
			AggregateID: aggregateID,
			Events:      events,
			Commands:    commands,
			IsMock:      false,
		}
	}
}

func (m EntryModel) View() string {
	if m.loading {
		return m.renderLoading()
	}

	leftPanel := m.renderMenu()
	rightPanel := m.renderPreviousRequests()

	// Calculate widths
	leftWidth := m.width/2 - 2
	rightWidth := m.width/2 - 2

	leftStyle := lipgloss.NewStyle().
		Width(leftWidth).
		Height(m.height-6).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#5c6bc0")).
		Padding(1, 2)

	rightStyle := lipgloss.NewStyle().
		Width(rightWidth).
		Height(m.height-6).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#5c6bc0")).
		Padding(1, 2)

	if m.previousIndex >= 0 {
		rightStyle = rightStyle.BorderForeground(lipgloss.Color("#ffcc00"))
	} else {
		leftStyle = leftStyle.BorderForeground(lipgloss.Color("#ffcc00"))
	}

	content := lipgloss.JoinHorizontal(lipgloss.Top,
		leftStyle.Render(leftPanel),
		rightStyle.Render(rightPanel),
	)

	title := TitleStyle.Render("Drill - Event Source Debugger")

	help := HelpStyle.Render("Tab/Arrows: navigate | Enter: select | q: quit")

	return lipgloss.JoinVertical(lipgloss.Left, title, content, help)
}

func (m EntryModel) renderLoading() string {
	style := lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Align(lipgloss.Center, lipgloss.Center)

	return style.Render(m.loadingMsg + "\n\nPlease wait...")
}

func (m EntryModel) renderMenu() string {
	var sb strings.Builder

	menuTitle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#ffffff")).
		MarginBottom(2).
		Render("Select an Option")

	sb.WriteString(menuTitle)
	sb.WriteString("\n\n")

	options := []string{
		"Load Account (Enter UUID)",
		"Run Mock Mode",
	}

	for i, opt := range options {
		style := lipgloss.NewStyle().Padding(0, 2)
		if m.previousIndex < 0 && menuOption(i) == m.menuSelection {
			style = style.
				Background(lipgloss.Color("#5c6bc0")).
				Foreground(lipgloss.Color("#ffffff")).
				Bold(true)
		}
		sb.WriteString(style.Render(opt))
		sb.WriteString("\n")
	}

	if m.inputMode {
		sb.WriteString("\n")
		sb.WriteString(m.textInput.View())
		sb.WriteString("\n")
		sb.WriteString(HelpStyle.Render("Press Enter to submit, Esc to cancel"))
	}

	if m.err != nil {
		sb.WriteString("\n\n")
		errStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#ff5252"))
		sb.WriteString(errStyle.Render("Error: " + m.err.Error()))
	}

	return sb.String()
}

func (m EntryModel) renderPreviousRequests() string {
	var sb strings.Builder

	menuTitle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#ffffff")).
		MarginBottom(2).
		Render("Previous Requests")

	sb.WriteString(menuTitle)
	sb.WriteString("\n\n")

	if len(m.cache.Requests) == 0 {
		sb.WriteString(HelpStyle.Render("No previous requests"))
		return sb.String()
	}

	for i, req := range m.cache.Requests {
		style := lipgloss.NewStyle().Padding(0, 1)
		if m.previousIndex == i {
			style = style.
				Background(lipgloss.Color("#5c6bc0")).
				Foreground(lipgloss.Color("#ffffff")).
				Bold(true)
		}

		mockLabel := ""
		if req.IsMock {
			mockLabel = " [MOCK]"
		}

		line := fmt.Sprintf("%s%s", req.AggregateID[:8]+"...", mockLabel)
		sb.WriteString(style.Render(line))
		sb.WriteString("\n")

		// Show timestamp and counts
		details := HelpStyle.Render(fmt.Sprintf("  %s | %d cmds, %d events",
			req.Timestamp.Format("Jan 02 15:04"),
			len(req.Commands),
			len(req.Events),
		))
		sb.WriteString(details)
		sb.WriteString("\n")
	}

	return sb.String()
}
