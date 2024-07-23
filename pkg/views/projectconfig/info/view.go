// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package info

import (
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/daytonaio/daytona/pkg/apiclient"
	"github.com/daytonaio/daytona/pkg/views"
	"golang.org/x/term"
)

const propertyNameWidth = 16

var propertyNameStyle = lipgloss.NewStyle().
	Foreground(views.LightGray)

var propertyValueStyle = lipgloss.NewStyle().
	Foreground(views.Light).
	Bold(true)

func Render(projectConfig *apiclient.ProjectConfig, forceUnstyled bool) {
	var output string
	output += "\n\n"

	output += views.GetStyledMainTitle("Project Config Info") + "\n\n"

	output += getInfoLine("Name", *projectConfig.Name) + "\n"

	if projectConfig.Repository != nil {
		output += getInfoLine("Repository", *projectConfig.Repository.Url) + "\n"
	}

	if GetLabelFromBuild(projectConfig.Build) != "" {
		output += getInfoLine("Build", GetLabelFromBuild(projectConfig.Build)) + "\n"
	}

	if projectConfig.Image != nil && *projectConfig.Image != "" {
		output += getInfoLine("Image", *projectConfig.Image) + "\n"
	}

	if projectConfig.User != nil && *projectConfig.User != "" {
		output += getInfoLine("User", *projectConfig.User) + "\n"
	}

	terminalWidth, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		fmt.Println(output)
		return
	}
	if terminalWidth < views.TUITableMinimumWidth || forceUnstyled {
		renderUnstyledInfo(output)
		return
	}

	renderTUIView(output, views.GetContainerBreakpointWidth(terminalWidth))
}

func renderUnstyledInfo(output string) {
	fmt.Println(output)
}

func renderTUIView(output string, width int) {
	output = lipgloss.NewStyle().PaddingLeft(3).Render(output)

	content := lipgloss.
		NewStyle().Width(width).
		Render(output)

	fmt.Println(content)
}

func getInfoLine(key, value string) string {
	return propertyNameStyle.Render(fmt.Sprintf("%-*s", propertyNameWidth, key)) + propertyValueStyle.Render(value) + "\n"
}

func GetLabelFromBuild(build *apiclient.ProjectBuild) string {
	if build == nil {
		return "Automatic"
	}

	if build.Devcontainer != nil && build.Devcontainer.DevContainerFilePath != nil {
		return fmt.Sprintf("Devcontainer (%s)", *build.Devcontainer.DevContainerFilePath)
	}

	return ""
}
