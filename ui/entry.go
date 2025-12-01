package ui

import (
	"drill/cache"
	"drill/fetcher"
	"drill/mock"
	"drill/models"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
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
	menuSelection   menuOption
	previousIndex   int
	textInput       textinput.Model
	inputMode       bool
	cache           *cache.Cache
	services        []models.ServiceConfig
	width           int
	height          int
	err             error
	loading         bool
	loadingMsg      string
	progress        progress.Model
	progressPercent float64
	progressSteps   []string
	currentStep     int
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

type ProgressMsg struct {
	Percent float64
	Step    string
}

type FetchStepMsg struct {
	ServiceName string
	StepType    string // "events" or "commands"
	Done        bool
}

func NewEntryModel(services []models.ServiceConfig) EntryModel {
	ti := textinput.New()
	ti.Placeholder = "Enter Aggregate ID (UUID)"
	ti.CharLimit = 36
	ti.Width = 40

	c, _ := cache.Load()

	p := progress.New(
		progress.WithDefaultGradient(),
		progress.WithWidth(40),
	)

	return EntryModel{
		menuSelection: optionLoadAccount,
		textInput:     ti,
		cache:         c,
		services:      services,
		progress:      p,
		previousIndex: -1,
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
				m.initProgressSteps(false)
				m.loadingMsg = "Connecting to services..."
				return m, tea.Batch(m.tickProgress(), m.loadFromServices(aggregateID))
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
				m.initProgressSteps(true)
				m.loadingMsg = "Connecting to mock services..."
				return m, tea.Batch(m.tickProgress(), m.loadMockDataWithProgress())
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

	case ProgressMsg:
		if m.currentStep < len(m.progressSteps) {
			m.loadingMsg = m.progressSteps[m.currentStep]
			m.progressPercent = float64(m.currentStep+1) / float64(len(m.progressSteps))
			m.currentStep++
			return m, m.tickProgress()
		}
		return m, nil

	case progress.FrameMsg:
		progressModel, cmd := m.progress.Update(msg)
		m.progress = progressModel.(progress.Model)
		return m, cmd
	}

	return m, nil
}

func (m *EntryModel) initProgressSteps(isMock bool) {
	var services []models.ServiceConfig
	if isMock {
		services = mock.MockServices
	} else {
		services = m.services
	}

	m.progressSteps = make([]string, 0)
	for _, svc := range services {
		m.progressSteps = append(m.progressSteps, fmt.Sprintf("Fetching events from %s...", svc.Name))
		m.progressSteps = append(m.progressSteps, fmt.Sprintf("Fetching commands from %s...", svc.Name))
	}
	m.currentStep = 0
	m.progressPercent = 0
}

func (m EntryModel) tickProgress() tea.Cmd {
	return tea.Tick(200*time.Millisecond, func(t time.Time) tea.Msg {
		return ProgressMsg{}
	})
}

func (m EntryModel) loadMockDataWithProgress() tea.Cmd {
	return func() tea.Msg {
		aggregateID := uuid.New().String()
		events, commands := mock.GenerateMockData(aggregateID)

		// Simulate network delay
		time.Sleep(200 * time.Millisecond)

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
			return LoadErrorMsg{Err: fmt.Errorf("no services configured. Set DRILL_SERVICES env var")}
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
	title := TitleStyle.Render("Drill - Event Source Debugger")

	// Progress content
	var content strings.Builder

	content.WriteString("\n\n")
	content.WriteString(m.loadingMsg)
	content.WriteString("\n\n")
	content.WriteString(m.progress.ViewAs(m.progressPercent))
	content.WriteString("\n\n")

	stepInfo := fmt.Sprintf("Step %d of %d", m.currentStep, len(m.progressSteps))
	content.WriteString(HelpStyle.Render(stepInfo))

	contentStyle := lipgloss.NewStyle().
		Width(m.width).
		Height(m.height-4).
		Align(lipgloss.Center, lipgloss.Center)

	return lipgloss.JoinVertical(lipgloss.Left, title, contentStyle.Render(content.String()))
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

		line := fmt.Sprintf("%s%s", req.AggregateID, mockLabel)
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
