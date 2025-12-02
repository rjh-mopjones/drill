package ui

import (
	"drill/models"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	Events         []models.Event
	eventsViewport viewport.Model
	detailViewport viewport.Model
	selectedIndex  int
	width          int
	height         int
	ready          bool
	aggregateID    string
	Loading        bool
	err            error
	Services       []models.ServiceConfig
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
		aggregateID:   aggregateID,
		Loading:       true,
		selectedIndex: 0,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "esc":
			// Go back to entry screen
			entry := NewEntryModel(m.Services)
			return entry, func() tea.Msg {
				return tea.WindowSizeMsg{Width: m.width, Height: m.height}
			}
		case "up", "k":
			if m.selectedIndex > 0 {
				m.selectedIndex--
				m.updateDetailView()
				m.updateEventsView()
			}
		case "down", "j":
			if m.selectedIndex < len(m.Events)-1 {
				m.selectedIndex++
				m.updateDetailView()
				m.updateEventsView()
			}
		case "pgup":
			m.selectedIndex -= 10
			if m.selectedIndex < 0 {
				m.selectedIndex = 0
			}
			m.updateDetailView()
			m.updateEventsView()
		case "pgdown":
			m.selectedIndex += 10
			if m.selectedIndex >= len(m.Events) {
				m.selectedIndex = len(m.Events) - 1
			}
			m.updateDetailView()
			m.updateEventsView()
		case "home", "g":
			m.selectedIndex = 0
			m.updateDetailView()
			m.updateEventsView()
		case "end", "G":
			m.selectedIndex = len(m.Events) - 1
			m.updateDetailView()
			m.updateEventsView()
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		headerHeight := 4
		footerHeight := 3
		availableHeight := m.height - headerHeight - footerHeight

		leftWidth := m.width / 2
		rightWidth := m.width - leftWidth - 3

		if !m.ready {
			m.eventsViewport = viewport.New(leftWidth, availableHeight)
			m.detailViewport = viewport.New(rightWidth, availableHeight)
			m.ready = true

			// Sort data (for data loaded from entry screen)
			sort.Slice(m.Events, func(i, j int) bool {
				return m.Events[i].Metadata.PersistedAt.Before(m.Events[j].Metadata.PersistedAt)
			})
		} else {
			m.eventsViewport.Width = leftWidth
			m.eventsViewport.Height = availableHeight
			m.detailViewport.Width = rightWidth
			m.detailViewport.Height = availableHeight
		}

		m.updateEventsView()
		m.updateDetailView()

	case DataLoadedMsg:
		m.Loading = false
		m.Events = msg.Events

		// Sort by persistedAt
		sort.Slice(m.Events, func(i, j int) bool {
			return m.Events[i].Metadata.PersistedAt.Before(m.Events[j].Metadata.PersistedAt)
		})

		m.updateEventsView()
		m.updateDetailView()

	case ErrorMsg:
		m.Loading = false
		m.err = msg.Err
	}

	// Handle viewport scrolling for detail panel
	var cmd tea.Cmd
	m.detailViewport, cmd = m.detailViewport.Update(msg)

	return m, cmd
}

func (m *Model) updateEventsView() {
	m.eventsViewport.SetContent(m.renderEvents())

	// Auto-scroll to keep selected item visible
	visibleLines := m.eventsViewport.Height - 2 // account for header
	if m.selectedIndex >= visibleLines {
		m.eventsViewport.SetYOffset(m.selectedIndex - visibleLines + 1)
	} else {
		m.eventsViewport.SetYOffset(0)
	}
}

func (m *Model) updateDetailView() {
	m.detailViewport.SetContent(m.renderEventDetail())
	m.detailViewport.GotoTop()
}

func (m Model) renderEvents() string {
	if len(m.Events) == 0 {
		return "No events found"
	}

	const (
		timeWidth    = 22
		serviceWidth = 20
		evtWidth     = 22
		corrWidth    = 38
		cellPadding  = "  " // spacing between cells
	)

	var sb strings.Builder

	// Header row
	timeCol := lipgloss.NewStyle().Width(timeWidth).Render("Time")
	svcCol := lipgloss.NewStyle().Width(serviceWidth).Render("Service")
	evtCol := lipgloss.NewStyle().Width(evtWidth).Render("Event")
	corrCol := lipgloss.NewStyle().Width(corrWidth).Render("CorrelationID")

	header := lipgloss.JoinHorizontal(lipgloss.Left, timeCol, cellPadding, svcCol, cellPadding, evtCol, cellPadding, corrCol)
	sb.WriteString(TableHeaderStyle.Render(header))
	sb.WriteString("\n")

	for i, evt := range m.Events {
		// Format time
		timeStr := evt.Metadata.PersistedAt.Format("2006-01-02 15:04:05")
		timeCell := lipgloss.NewStyle().Width(timeWidth).Render(timeStr)

		// Format service name with unique color
		svcText := CreateServiceStyle(evt.ServiceName).Render(evt.ServiceName)
		svcCell := lipgloss.NewStyle().Width(serviceWidth).Render(svcText)

		// Event alias
		evtAlias := evt.Metadata.EventAlias
		if len(evtAlias) > evtWidth-2 {
			evtAlias = evtAlias[:evtWidth-5] + "..."
		}
		evtCell := lipgloss.NewStyle().Width(evtWidth).Render(evtAlias)

		// Correlation ID (full UUID)
		correlationText := CreateCorrelationStyle(evt.Metadata.CorrelationID).Render(evt.Metadata.CorrelationID)
		corrCell := lipgloss.NewStyle().Width(corrWidth).Render(correlationText)

		// Join all cells with padding
		row := lipgloss.JoinHorizontal(lipgloss.Left, timeCell, cellPadding, svcCell, cellPadding, evtCell, cellPadding, corrCell)

		// Highlight selected row
		if i == m.selectedIndex {
			row = SelectedRowStyle.Render(row)
		}

		sb.WriteString(row)
		sb.WriteString("\n")
	}

	return sb.String()
}

func (m Model) renderEventDetail() string {
	if len(m.Events) == 0 || m.selectedIndex >= len(m.Events) {
		return "No event selected"
	}

	evt := m.Events[m.selectedIndex]

	var sb strings.Builder

	// Title
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#ffffff")).MarginBottom(1)
	sb.WriteString(titleStyle.Render(evt.Metadata.EventAlias))
	sb.WriteString("\n\n")

	// Labels
	labelStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#888888"))
	valueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#ffffff"))

	// Event ID
	sb.WriteString(labelStyle.Render("Event ID:"))
	sb.WriteString("\n")
	sb.WriteString(valueStyle.Render(evt.Metadata.EventID))
	sb.WriteString("\n\n")

	// Service
	sb.WriteString(labelStyle.Render("Service:"))
	sb.WriteString("\n")
	svcStyled := CreateServiceStyle(evt.ServiceName).Render(evt.ServiceName)
	sb.WriteString(svcStyled)
	sb.WriteString("\n\n")

	// Persisted At
	sb.WriteString(labelStyle.Render("Persisted At:"))
	sb.WriteString("\n")
	sb.WriteString(valueStyle.Render(evt.Metadata.PersistedAt.Format("2006-01-02 15:04:05.000")))
	sb.WriteString("\n\n")

	// Correlation ID
	sb.WriteString(labelStyle.Render("Correlation ID:"))
	sb.WriteString("\n")
	corrStyled := CreateCorrelationStyle(evt.Metadata.CorrelationID).Render(evt.Metadata.CorrelationID)
	sb.WriteString(corrStyled)
	sb.WriteString("\n\n")

	// Aggregate ID
	sb.WriteString(labelStyle.Render("Aggregate ID:"))
	sb.WriteString("\n")
	sb.WriteString(valueStyle.Render(evt.Metadata.AggregateID))
	sb.WriteString("\n\n")

	// Payload
	sb.WriteString(labelStyle.Render("Payload:"))
	sb.WriteString("\n")

	// Pretty print JSON payload
	if evt.Payload != "" {
		var prettyJSON map[string]interface{}
		if err := json.Unmarshal([]byte(evt.Payload), &prettyJSON); err == nil {
			formatted, _ := json.MarshalIndent(prettyJSON, "", "  ")
			sb.WriteString(valueStyle.Render(string(formatted)))
		} else {
			sb.WriteString(valueStyle.Render(evt.Payload))
		}
	} else {
		sb.WriteString(HelpStyle.Render("(empty)"))
	}

	return sb.String()
}

func (m Model) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v\n\nPress q to quit.", m.err)
	}

	if m.Loading {
		return "Loading data...\n\nPress q to quit."
	}

	if !m.ready {
		return "Initializing..."
	}

	// Title
	title := TitleStyle.Render(fmt.Sprintf("Event Debugger - Aggregate: %s", m.aggregateID))

	// Create panel headers
	leftWidth := m.width / 2
	rightWidth := m.width - leftWidth - 3

	eventsHeader := HeaderStyle.Width(leftWidth).Align(lipgloss.Center).Render("EVENTS")
	detailHeader := HeaderStyle.Width(rightWidth).Align(lipgloss.Center).Render("EVENT DETAIL")

	headers := lipgloss.JoinHorizontal(lipgloss.Top, eventsHeader, " ", detailHeader)

	// Create bordered viewports
	eventsBorder := BorderStyle.BorderForeground(lipgloss.Color("#ffcc00"))
	detailBorder := BorderStyle

	eventsBox := eventsBorder.Width(leftWidth).Render(m.eventsViewport.View())
	detailBox := detailBorder.Width(rightWidth).Render(m.detailViewport.View())

	panels := lipgloss.JoinHorizontal(lipgloss.Top, eventsBox, " ", detailBox)

	// Stats and help
	stats := fmt.Sprintf("Events: %d | Selected: %d/%d", len(m.Events), m.selectedIndex+1, len(m.Events))
	help := HelpStyle.Render("j/k: navigate | Esc: back | q: quit")

	return lipgloss.JoinVertical(lipgloss.Left,
		title,
		headers,
		panels,
		stats,
		help,
	)
}
