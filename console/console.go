// Copyright 2025 The https://github.com/alphacoderun/console-example Authors. All rights reserved.
package console

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	// Styles
	docStyle       = lipgloss.NewStyle().Margin(1, 2)
	statusBar      = lipgloss.NewStyle().Background(lipgloss.Color("0")).Foreground(lipgloss.Color("6")).Padding(0, 1)
	activeTabStyle = lipgloss.NewStyle().Background(lipgloss.Color("0")).Foreground(lipgloss.Color("5")).Padding(0, 1).Bold(true)

	inactiveTabStyle = lipgloss.NewStyle().Background(lipgloss.Color("0")).Foreground(lipgloss.Color("2")).Padding(0, 1)

	viewportStyle  = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(0, 1)
	inputTextStyle = lipgloss.NewStyle().Border(lipgloss.HiddenBorder())

	tabBorderStyle = lipgloss.NewStyle().Border(lipgloss.NormalBorder(), false, false, false, true).BorderForeground(lipgloss.Color("8"))
	helpStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
)

type model struct {
	tabs         []string
	activeTab    int
	tabContents  []string // Raw string content for each tab
	tabViewports []viewport.Model
	inputText    textinput.Model
	width        int
	height       int
	ready        bool // To handle initial sizing
	execute      ExecuteCommand
}

// Inialize the model with tabs and an execute function
// @tabs: List of tab names
// @execute: Function to execute commands in the active tab
func initialModel(tabs []string, execute ExecuteCommand) model {
	// tabs := execute.GetTabNames() //[]string{"Info", "Data Stream", "Logs"}
	tabContents := make([]string, len(tabs))
	tabViewports := make([]viewport.Model, len(tabs))

	// Initial content for tab viewports
	for i := 0; i < len(tabContents); i++ {
		tabContents[i] = "" // Initialize with empty content
	}

	input := textinput.New()
	input.Placeholder = "Type something and press Enter..."
	input.Focus()

	m := model{
		tabs:         tabs,
		activeTab:    0,
		tabContents:  tabContents,
		tabViewports: tabViewports,
		inputText:    input,
		execute:      execute,
	}
	return m
}

func (m model) Init() tea.Cmd {
	return textarea.Blink
}

// Update handles messages and updates the model accordingly.
// It processes input from the textarea, viewport updates, and global keybindings.
// It also manages the active tab and viewport content.
// @msg: tea.Msg - The message to process
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	// Handle input area updates first
	m.inputText, cmd = m.inputText.Update(msg)
	cmds = append(cmds, cmd)

	// Handle viewport updates for the active tab
	if m.ready && len(m.tabViewports) > m.activeTab {
		// Pass general key messages (like arrows for scrolling) to the active viewport.
		// Specific Ctrl+Left/Right will be handled below to ensure they attempt horizontal scroll.
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Global keybindings
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyTab:
			m.activeTab = (m.activeTab + 1) % len(m.tabs)
			if m.ready { // Ensure viewports are initialized
				m.tabViewports[m.activeTab].SetContent(m.tabContents[m.activeTab])
				// m.tabViewports[m.activeTab].GotoTop() // Optionally reset scroll on tab switch
			}
			m.inputText.Focus() // Keep or return focus to input area
			cmds = append(cmds, textarea.Blink)

		case tea.KeyEnter:
			if m.inputText.Focused() && m.inputText.Value() != "" {
				newText := m.inputText.Value()
				if newText == "clear" {
					// Clear the content of the active tab
					m.tabContents[m.activeTab] = ""
					m.tabViewports[m.activeTab].SetContent(m.tabContents[m.activeTab])
					m.tabViewports[m.activeTab].GotoTop() // Reset scroll to top
					m.inputText.Reset()                   // Clear input field
					m.inputText.Focus()                   // Return focus to input area
					cmds = append(cmds, textarea.Blink)
					return m, nil
				}
				output, err := m.execute(m.tabs[m.activeTab], newText)
				// Add a newline if current text isn't empty and doesn't end with one
				if m.tabContents[m.activeTab] != "" && !strings.HasSuffix(m.tabContents[m.activeTab], "\n") {
					m.tabContents[m.activeTab] += "\n"
				}
				if err != nil {
					m.tabContents[m.activeTab] += output + "\n" + err.Error()
				} else {
					m.tabContents[m.activeTab] += output + "\n"
				}
				m.tabViewports[m.activeTab].SetContent(m.tabContents[m.activeTab])
				m.tabViewports[m.activeTab].GotoBottom() // Scroll to show the new text
				m.inputText.Reset()
				m.inputText.Focus()
				cmds = append(cmds, textarea.Blink)
			}

		// Horizontal scrolling for active viewport with Ctrl+Left/Right
		// We send a simple KeyLeft/KeyRight to the viewport, it should handle XOffset.
		case tea.KeyCtrlLeft:
			if m.ready && len(m.tabViewports) > m.activeTab {
				m.tabViewports[m.activeTab].ScrollLeft(5)
				return m, nil
			}
		case tea.KeyCtrlRight:
			if m.ready && len(m.tabViewports) > m.activeTab {
				m.tabViewports[m.activeTab].ScrollRight(5)
				return m, nil
			}

		default:
			// Pass other relevant keys (like arrows, pgup/pgdown) to the active viewport for scrolling
			// if the input area is not the primary recipient or if the key is a navigation key.
			// For simplicity, if it's an arrow/nav key, let the viewport try to handle it.
			// This allows natural scrolling of the viewport when not typing in inputArea.
			// A more robust focus model might be needed for complex interactions.
			if m.ready && len(m.tabViewports) > m.activeTab {
				// //isInputFocused := m.inputArea.Focused()
				isArrowKey := msg.Type == tea.KeyLeft || msg.Type == tea.KeyRight || msg.Type == tea.KeyUp || msg.Type == tea.KeyDown || msg.Type == tea.KeyPgUp || msg.Type == tea.KeyPgDown

				// If input is not focused OR it's an arrow key (which textarea might not fully consume for its own single line)
				// let the viewport try to process it.
				// This logic might need refinement based on desired focus behavior.
				// A simple approach: always let viewport try to update on arrow keys.
				if isArrowKey {
					m.tabViewports[m.activeTab], cmd = m.tabViewports[m.activeTab].Update(msg)
					cmds = append(cmds, cmd)
				}
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		headerHeight := lipgloss.Height(m.renderStatusBar()) + lipgloss.Height(m.renderTabs())
		inputAreaRenderedHeight := 5 // Approximate height for input area with borders

		viewportContentHeight := m.height - docStyle.GetVerticalMargins() - headerHeight - inputAreaRenderedHeight - viewportStyle.GetVerticalBorderSize()

		// Set widths accounting for docStyle margins and component borders
		contentWidth := m.width - docStyle.GetHorizontalMargins()
		viewportWidth := contentWidth - viewportStyle.GetHorizontalBorderSize()
		inputWidth := contentWidth - inputTextStyle.GetHorizontalBorderSize()

		if !m.ready { // First layout
			for i := range m.tabs {
				vp := viewport.New(viewportWidth, viewportContentHeight)
				vp.SetContent(m.tabContents[i])
				m.tabViewports[i] = vp
			}
			m.ready = true
		} else { // Resize
			for i := range m.tabViewports {
				m.tabViewports[i].Width = viewportWidth
				m.tabViewports[i].Height = viewportContentHeight
				m.tabViewports[i].SetContent(m.tabContents[i]) // Re-set content to re-flow
			}
		}

		m.inputText.Width = inputWidth
		if m.inputText.Width < 0 {
			m.inputText.Width = 0 // Ensure input area width is not negative
		}
	}

	// Ensure active viewport is up-to-date (e.g., after SetContent)
	if m.ready && len(m.tabViewports) > m.activeTab {
		// m.tabViewports[m.activeTab], cmd = m.tabViewports[m.activeTab].Update(nil) // Not always needed if SetContent triggers re-render
		// cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// Render the status bar with active tab and total tabs count.
// It also indicates that the application is ready and provides a quit message.
// @return: string - Rendered status bar with active tab and total tabs coun
func (m model) renderStatusBar() string {
	if !m.ready {
		return statusBar.Render("Initializing...")
	}
	statusText := fmt.Sprintf("Active: %s | Total: %d | Ctrl+C to quit | Type clear to clear tab", m.tabs[m.activeTab], len(m.tabs))
	return statusBar.Width(m.width - docStyle.GetHorizontalMargins()).Render(statusText)
}

// Render the tabs with active and inactive styles.
// It highlights the active tab and provides a border around the tabs.
// It returns a string representation of the rendered tabs.
// The tabs are rendered horizontally, and the active tab is highlighted.
// @return: string - Rendered tabs with styles applied
// @note: The tabs are rendered with a border at the bottom, and the active tab is bolded.
func (m model) renderTabs() string {
	if !m.ready {
		return ""
	}
	var renderedTabs []string
	for i, t := range m.tabs {
		style := inactiveTabStyle
		if i == m.activeTab {
			style = activeTabStyle
		}
		renderedTabs = append(renderedTabs, style.Render(t))
	}
	return tabBorderStyle.Width(m.width - docStyle.GetHorizontalMargins()).Render(lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs...))
}

// View renders the entire application view.
// It combines the status bar, tabs, active viewport content, and input area.
// It returns a string representation of the complete view.
// @return: string - Rendered view of the application
func (m model) View() string {
	if !m.ready || len(m.tabViewports) == 0 { // Ensure viewports are initialized
		return docStyle.Render("Loading...")
	}

	sBar := m.renderStatusBar()
	tabs := m.renderTabs()

	activeViewportView := viewportStyle.Render(m.tabViewports[m.activeTab].View())
	input := inputTextStyle.Render(m.inputText.View())

	help := helpStyle.Render("Use Tab to switch, Ctrl+Left/Right to scroll horizontally, Arrows to scroll vertically. Enter in input field.")

	return docStyle.Render(lipgloss.JoinVertical(
		lipgloss.Left,
		sBar,
		tabs,
		activeViewportView,
		input,
		help,
	))
}
