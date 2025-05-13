package main

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

// Check represents a single check to run
type Check struct {
	Name        string
	Description string
	Command     string
	Args        []string
	Result      *CheckResult
}

// CheckResult holds the result of a check
type CheckResult struct {
	Success bool
	Output  string
}

func main() {
	// Define styles
	var (
		magenta = lipgloss.Color("5")
		white   = lipgloss.Color("15")
		green   = lipgloss.Color("2")
		red     = lipgloss.Color("1")

		tableHeaderStyle = lipgloss.NewStyle().Foreground(magenta).Bold(true).Align(lipgloss.Center)
		cellStyle        = lipgloss.NewStyle().Padding(0, 1).Foreground(white)
		failedStyle      = lipgloss.NewStyle().Foreground(red)
		successStyle     = lipgloss.NewStyle().Foreground(green)

		listHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(magenta).
				MarginTop(1).
				MarginBottom(1)

		listItemStyle = lipgloss.NewStyle().
				PaddingLeft(2)
	)

	// Define checks
	checks := []Check{
		{
			Name:        "go fmt",
			Description: "Check code formatting",
			Command:     "go",
			Args:        []string{"fmt", "./..."},
		},
		{
			Name:        "go vet",
			Description: "Check source code for suspicious constructs",
			Command:     "go",
			Args:        []string{"vet", "./..."},
		},
		{
			Name:        "go test",
			Description: "Check that tests pass",
			Command:     "go",
			Args:        []string{"test", "./..."},
		},
		{
			Name:        "gosec",
			Description: "Check for potential security issues",
			Command:     "gosec",
			Args:        []string{"./..."},
		},
		{
			Name:        "staticcheck",
			Description: "Check for bugs & performance issues",
			Command:     "staticcheck",
			Args:        []string{"./..."},
		},
	}

	// Run all checks
	for i := range checks {
		checks[i].Result = runCheck(checks[i].Command, checks[i].Args)
	}

	// Create table rows
	var tableRows [][]string
	for _, check := range checks {
		status := failedStyle.Render("failed") // ✖
		if check.Result.Success {
			status = successStyle.Render("passed") // ✔
		}
		tableRows = append(tableRows, []string{status, check.Name, check.Description})
	}

	// Create and print table
	t := table.New().
		Border(lipgloss.RoundedBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(magenta)).
		StyleFunc(func(row, col int) lipgloss.Style {
			switch {
			case row == table.HeaderRow:
				return tableHeaderStyle
			default:
				return cellStyle
			}
		}).
		Headers("STATUS", "NAME", "DESCRIPTION").
		Rows(tableRows...)
	fmt.Println(t)

	// if there are failed checks, print them
	failedChecks := 0
	for _, check := range checks {
		if !check.Result.Success {
			failedChecks++
		}
	}

	if failedChecks > 0 {
		fmt.Println(listHeaderStyle.Render("Failed Checks:"))

		for _, check := range checks {
			if !check.Result.Success {
				fmt.Printf("%s %s\n", failedStyle.Render("✖"), check.Name)

				output := check.Result.Output
				if output == "" {
					output = "No output (but command failed)"
				}

				lines := strings.Split(output, "\n")
				for _, line := range lines {
					if line != "" {
						fmt.Println(listItemStyle.Render(line))
					}
				}
			}
		}
	} else {
		fmt.Println(listHeaderStyle.Render("All checks passed! 🎉"))
	}
}

func runCheck(command string, args []string) *CheckResult {
	cmd := exec.Command(command, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	// Combine stdout and stderr
	output := stdout.String()
	if stderr.String() != "" {
		if output != "" {
			output += "\n"
		}
		output += stderr.String()
	}

	// Trim whitespace
	output = strings.TrimSpace(output)

	// Check success - a command is successful if it returns without error
	// Note: go fmt specifically is successful if there's no output
	success := err == nil
	if command == "go" && args[0] == "fmt" {
		success = output == ""
	}

	return &CheckResult{
		Success: success,
		Output:  output,
	}
}
