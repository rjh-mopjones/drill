package ui

import (
	"drill/models"
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type focus int

const (
	focusCommands focus = iota
	focusEvents
)

type Model struct {
	commands         []models.Command
	events           []models.Event
	commandsViewport viewport.Model
	eventsViewport   viewport.Model
	focus            focus
	width            int
	height           int
	ready            bool
	aggregateID      string
	loading          bool
	err              error
}

type DataLoadedMsg struct {
	Events   []models.Event
	Commands []models.Command
}

type ErrorMsg struct {
	Err error
}

func NewModel(aggregateID string) Model {
	return Model{
		aggregateID: aggregateID,
		loading:     true,
		focus:       focusCommands,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "tab":
			if m.focus == focusCommands {
				m.focus = focusEvents
			} else {
				m.focus = focusCommands
			}
		case "up", "k":
			if m.focus == focusCommands {
				m.commandsViewport.LineUp(1)
			} else {
				m.eventsViewport.LineUp(1)
			}
		case "down", "j":
			if m.focus == focusCommands {
				m.commandsViewport.LineDown(1)
			} else {
				m.eventsViewport.LineDown(1)
			}
		case "pgup":
			if m.focus == focusCommands {
				m.commandsViewport.HalfViewUp()
			} else {
				m.eventsViewport.HalfViewUp()
			}
		case "pgdown":
			if m.focus == focusCommands {
				m.commandsViewport.HalfViewDown()
			} else {
				m.eventsViewport.HalfViewDown()
			}
		case "home", "g":
			if m.focus == focusCommands {
				m.commandsViewport.GotoTop()
			} else {
				m.eventsViewport.GotoTop()
			}
		case "end", "G":
			if m.focus == focusCommands {
				m.commandsViewport.GotoBottom()
			} else {
				m.eventsViewport.GotoBottom()
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		headerHeight := 4
		footerHeight := 3
		availableHeight := m.height - headerHeight - footerHeight

		halfWidth := (m.width - 3) / 2

		if !m.ready {
			m.commandsViewport = viewport.New(halfWidth, availableHeight)
			m.eventsViewport = viewport.New(halfWidth, availableHeight)
			m.ready = true
		} else {
			m.commandsViewport.Width = halfWidth
			m.commandsViewport.Height = availableHeight
			m.eventsViewport.Width = halfWidth
			m.eventsViewport.Height = availableHeight
		}

		m.commandsViewport.SetContent(m.renderCommands())
		m.eventsViewport.SetContent(m.renderEvents())

	case DataLoadedMsg:
		m.loading = false
		m.commands = msg.Commands
		m.events = msg.Events

		// Sort by persistedAt
		sort.Slice(m.commands, func(i, j int) bool {
			return m.commands[i].PersistedAt.Before(m.commands[j].PersistedAt)
		})
		sort.Slice(m.events, func(i, j int) bool {
			return m.events[i].PersistedAt.Before(m.events[j].PersistedAt)
		})

		m.commandsViewport.SetContent(m.renderCommands())
		m.eventsViewport.SetContent(m.renderEvents())

	case ErrorMsg:
		m.loading = false
		m.err = msg.Err
	}

	return m, tea.Batch(cmds...)
}

func (m Model) renderCommands() string {
	if len(m.commands) == 0 {
		return "No commands found"
	}

	const (
		timeWidth    = 22
		serviceWidth = 22
		cmdWidth     = 25
		statusWidth  = 12
		corrWidth    = 38
	)

	var sb strings.Builder

	// Header row - build with fixed-width columns
	timeCol := lipgloss.NewStyle().Width(timeWidth).Render("Time")
	svcCol := lipgloss.NewStyle().Width(serviceWidth).Render("Service")
	cmdCol := lipgloss.NewStyle().Width(cmdWidth).Render("Command")
	statusCol := lipgloss.NewStyle().Width(statusWidth).Render("Status")
	corrCol := lipgloss.NewStyle().Width(corrWidth).Render("Correlation")

	header := lipgloss.JoinHorizontal(lipgloss.Left, timeCol, svcCol, cmdCol, statusCol, corrCol)
	sb.WriteString(TableHeaderStyle.Render(header))
	sb.WriteString("\n")

	for _, cmd := range m.commands {
		// Format time - pad to fixed width
		timeStr := cmd.PersistedAt.Format("2006-01-02 15:04:05")
		timeCell := lipgloss.NewStyle().Width(timeWidth).Render(timeStr)

		// Format service name with unique color
		svcText := CreateServiceStyle(cmd.ServiceName).Render(cmd.ServiceName)
		svcCell := lipgloss.NewStyle().Width(serviceWidth).Render(svcText)

		// Command alias
		cmdCell := lipgloss.NewStyle().Width(cmdWidth).Render(cmd.CommandAlias)

		// Format status with color
		var statusText string
		if cmd.CommandStatus == models.CommandFailed {
			statusText = FailedCommandStyle.Render("FAILED")
		} else {
			statusText = SuccessCommandStyle.Render("SUCCEEDED")
		}
		statusCell := lipgloss.NewStyle().Width(statusWidth).Render(statusText)

		// Format correlation ID with unique color - show full UUID
		correlationText := CreateCorrelationStyle(cmd.CorrelationID).Render(cmd.CorrelationID)
		correlationStr := lipgloss.NewStyle().Width(corrWidth).Render(correlationText)

		// Join all cells
		row := lipgloss.JoinHorizontal(lipgloss.Left, timeCell, svcCell, cmdCell, statusCell, correlationStr)
		sb.WriteString(row)
		sb.WriteString("\n")
	}

	return sb.String()
}

func (m Model) renderEvents() string {
	if len(m.events) == 0 {
		return "No events found"
	}

	const (
		timeWidth    = 22
		serviceWidth = 22
		evtWidth     = 25
		corrWidth    = 38
	)

	var sb strings.Builder

	// Header row - build with fixed-width columns
	timeCol := lipgloss.NewStyle().Width(timeWidth).Render("Time")
	svcCol := lipgloss.NewStyle().Width(serviceWidth).Render("Service")
	evtCol := lipgloss.NewStyle().Width(evtWidth).Render("Event")
	corrCol := lipgloss.NewStyle().Width(corrWidth).Render("Correlation")

	header := lipgloss.JoinHorizontal(lipgloss.Left, timeCol, svcCol, evtCol, corrCol)
	sb.WriteString(TableHeaderStyle.Render(header))
	sb.WriteString("\n")

	for _, evt := range m.events {
		// Format time - pad to fixed width
		timeStr := evt.PersistedAt.Format("2006-01-02 15:04:05")
		timeCell := lipgloss.NewStyle().Width(timeWidth).Render(timeStr)

		// Format service name with unique color
		svcText := CreateServiceStyle(evt.ServiceName).Render(evt.ServiceName)
		svcCell := lipgloss.NewStyle().Width(serviceWidth).Render(svcText)

		// Event alias
		evtCell := lipgloss.NewStyle().Width(evtWidth).Render(evt.EventAlias)

		// Format correlation ID with unique color - show full UUID
		correlationText := CreateCorrelationStyle(evt.CorrelationID).Render(evt.CorrelationID)
		correlationStr := lipgloss.NewStyle().Width(corrWidth).Render(correlationText)

		// Join all cells
		row := lipgloss.JoinHorizontal(lipgloss.Left, timeCell, svcCell, evtCell, correlationStr)
		sb.WriteString(row)
		sb.WriteString("\n")
	}

	return sb.String()
}

func (m Model) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v\n\nPress q to quit.", m.err)
	}

	if m.loading {
		return "Loading data...\n\nPress q to quit."
	}

	if !m.ready {
		return "Initializing..."
	}

	// Title
	title := TitleStyle.Render(fmt.Sprintf("Event Debugger - Aggregate: %s", m.aggregateID))

	// Create table headers
	halfWidth := (m.width - 3) / 2

	commandsTitle := "COMMANDS"
	eventsTitle := "EVENTS"

	if m.focus == focusCommands {
		commandsTitle = "> " + commandsTitle + " <"
	}
	if m.focus == focusEvents {
		eventsTitle = "> " + eventsTitle + " <"
	}

	commandsHeader := HeaderStyle.Width(halfWidth).Align(lipgloss.Center).Render(commandsTitle)
	eventsHeader := HeaderStyle.Width(halfWidth).Align(lipgloss.Center).Render(eventsTitle)

	headers := lipgloss.JoinHorizontal(lipgloss.Top, commandsHeader, " ", eventsHeader)

	// Create bordered viewports
	commandsBorder := BorderStyle
	eventsBorder := BorderStyle

	if m.focus == focusCommands {
		commandsBorder = commandsBorder.BorderForeground(lipgloss.Color("#ffcc00"))
	} else {
		eventsBorder = eventsBorder.BorderForeground(lipgloss.Color("#ffcc00"))
	}

	commandsBox := commandsBorder.Width(halfWidth).Render(m.commandsViewport.View())
	eventsBox := eventsBorder.Width(halfWidth).Render(m.eventsViewport.View())

	tables := lipgloss.JoinHorizontal(lipgloss.Top, commandsBox, " ", eventsBox)

	// Stats and help
	stats := fmt.Sprintf("Commands: %d | Events: %d", len(m.commands), len(m.events))
	help := HelpStyle.Render("Tab: switch panels | j/k or arrows: scroll | pgup/pgdown: page | q: quit")

	return lipgloss.JoinVertical(lipgloss.Left,
		title,
		headers,
		tables,
		stats,
		help,
	)
}
