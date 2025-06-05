// Copyright 2025 The https://github.com/alphacoderun/console-example Authors. All rights reserved.
package console

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

// App interface defines the methods for the console application.
type App interface {
	Run(tabNames []string, execute ExecuteCommand) error
}

// ExecuteCommand is a function type that takes a tab name and a command string,
// and returns the output as a string or an error.
// This function is used to execute commands in the context of a specific tab.
// It allows the application to interact with the underlying system or service
// associated with the tab, providing a way to run commands and retrieve results.
// @tab: The name of the active tab where the command should be executed.
// @command: The command to be executed in the specified tab.
// @return: A string containing the output of the command execution, or an error if the execution fails.
type ExecuteCommand func(tab string, command string) (string, error)

// appImpl is the implementation of the App interface.
type appImpl struct {
}

// Run starts the application and blocks until it exits.
func (a appImpl) Run(tabNames []string, execute ExecuteCommand) error {
	p := tea.NewProgram(initialModel(tabNames, execute), tea.WithAltScreen(), tea.WithMouseCellMotion()) // AltScreen & Mouse (optional)
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("error running program: %w", err)
	}
	return nil
}

// NewApp creates a new instance of the console application.
// It initializes the application with default settings and returns an App interface.
// This function is the entry point for creating a new console application instance.
func NewApp() App {
	app := appImpl{}
	return app
}
