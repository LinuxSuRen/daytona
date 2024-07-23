// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package build

type ProjectBuildDevcontainer struct {
	DevContainerFilePath string `json:"devContainerFilePath"`
} // @name ProjectBuildDevcontainer

/*
type ProjectBuildDockerfile struct {
	Context    string            `json:"context"`
	Dockerfile string            `json:"dockerfile"`
	Args       map[string]string `json:"args"`
} // @name ProjectBuildDockerfile
*/

type ProjectBuild struct {
	Devcontainer *ProjectBuildDevcontainer `json:"devcontainer"`
	/*
		Dockerfile   *ProjectBuildDockerfile   `json:"dockerfile"`
	*/
} // @name ProjectBuild
