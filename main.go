// Copyright 2025 The https://github.com/alphacoderun/console-example Authors. All rights reserved.
package main

import (
	"console-example/console"
	_ "embed"
	"fmt"
)

//go:embed LICENSE
var license string

//go:embed LICENSE-THIRD-PARTY
var licenseThirdParty string

func ExecuteCommand(tab string, command string) (string, error) {

	output := fmt.Sprintf("You have Executed command '%s' on tab '%s'", command, tab)
	if tab == "Tab2" {
		output = fmt.Sprintf("You have Executed command '%s' on tab '%s' with a random error message", command, tab)
	}
	if tab == "License" {
		if command == "license" {
			output = license
		} else if command == "third-party" {
			output = licenseThirdParty
		} else {
			output = "Enter 'license' or 'third-party' to see the respective license information."
		}
	}

	return output, nil
}

func main() {
	console := console.NewApp()
	tabNames := []string{"Tab1", "Tab2", "Tab3", "License"}
	console.Run(tabNames, ExecuteCommand)
}
