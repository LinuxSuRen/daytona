// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package projectconfig

import (
	"errors"
	"strings"

	"github.com/daytonaio/daytona/pkg/workspace/project/config"
)

type IProjectConfigService interface {
	Delete(projectConfig *config.ProjectConfig) error
	Find(projectConfigName string) (*config.ProjectConfig, error)
	FindDefault(url string) (*config.ProjectConfig, error)
	List() ([]*config.ProjectConfig, error)
	FilterByGitUrl(url string) ([]*config.ProjectConfig, error)
	SetDefault(projectConfigName string) error
	Save(projectConfig *config.ProjectConfig) error
}

type ProjectConfigServiceConfig struct {
	ConfigStore config.Store
}

type ProjectConfigService struct {
	configStore config.Store
}

func NewProjectConfigService(config ProjectConfigServiceConfig) IProjectConfigService {
	return &ProjectConfigService{
		configStore: config.ConfigStore,
	}
}

func (s *ProjectConfigService) List() ([]*config.ProjectConfig, error) {
	return s.configStore.List()
}

func (s *ProjectConfigService) FilterByGitUrl(url string) ([]*config.ProjectConfig, error) {
	projectConfigs, err := s.configStore.List()
	if err != nil {
		return nil, err
	}

	url = strings.TrimSuffix(url, "/")
	url = strings.TrimSuffix(url, ".git")

	var response []*config.ProjectConfig

	for _, pc := range projectConfigs {
		if pc.Repository == nil {
			continue
		}

		currentUrl := strings.TrimSuffix(pc.Repository.Url, "/")
		currentUrl = strings.TrimSuffix(currentUrl, ".git")

		if currentUrl != url {
			continue
		}

		response = append(response, pc)
	}

	return response, nil
}

func (s *ProjectConfigService) FindDefault(url string) (*config.ProjectConfig, error) {
	projectConfigs, err := s.FilterByGitUrl(url)
	if err != nil {
		return nil, err
	}

	for _, pc := range projectConfigs {
		if pc.IsDefault {
			return pc, nil
		}
	}

	return nil, config.ErrProjectConfigNotFound
}

func (s *ProjectConfigService) SetDefault(projectConfigName string) error {
	projectConfig, err := s.Find(projectConfigName)
	if err != nil {
		return err
	}

	if projectConfig == nil {
		return config.ErrProjectConfigNotFound
	}

	if projectConfig.Repository == nil {
		return errors.New("project config does not have a repository")
	}

	projectConfigs, err := s.FilterByGitUrl(projectConfig.Repository.Url)
	if err != nil {
		return err
	}

	for _, pc := range projectConfigs {
		if pc.Name == projectConfigName {
			pc.IsDefault = true
		} else {
			pc.IsDefault = false
		}
	}

	return nil
}

func (s *ProjectConfigService) Find(projectConfigName string) (*config.ProjectConfig, error) {
	return s.configStore.Find(projectConfigName)
}

func (s *ProjectConfigService) Save(projectConfig *config.ProjectConfig) error {
	return s.configStore.Save(projectConfig)
}

func (s *ProjectConfigService) Delete(projectConfig *config.ProjectConfig) error {
	return s.configStore.Delete(projectConfig)
}
