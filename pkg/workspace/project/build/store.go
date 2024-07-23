// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package build

import "errors"

type Store interface {
	List() ([]*ProjectBuild, error)
	Find(name string) (*ProjectBuild, error)
	Save(projectBuild *ProjectBuild) error
	Delete(projectBuild *ProjectBuild) error
}

var (
	ErrProjectBuildNotFound = errors.New("project build not found")
)

func IsProjectBuildNotFound(err error) bool {
	return err.Error() == ErrProjectBuildNotFound.Error()
}
