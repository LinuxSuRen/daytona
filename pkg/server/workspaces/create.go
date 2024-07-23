// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package workspaces

import (
	"fmt"
	"io"
	"regexp"

	"github.com/daytonaio/daytona/pkg/apikey"
	"github.com/daytonaio/daytona/pkg/build"
	"github.com/daytonaio/daytona/pkg/containerregistry"
	"github.com/daytonaio/daytona/pkg/gitprovider"
	"github.com/daytonaio/daytona/pkg/logs"
	"github.com/daytonaio/daytona/pkg/provider"
	"github.com/daytonaio/daytona/pkg/server/workspaces/dto"
	"github.com/daytonaio/daytona/pkg/workspace"
	"github.com/daytonaio/daytona/pkg/workspace/project"
	"github.com/daytonaio/daytona/pkg/workspace/project/config"
)

func (s *WorkspaceService) CreateWorkspace(req dto.CreateWorkspaceDTO) (*workspace.Workspace, error) {
	_, err := s.workspaceStore.Find(req.Name)
	if err == nil {
		return nil, ErrWorkspaceAlreadyExists
	}

	isAlphaNumeric := regexp.MustCompile(`^[a-zA-Z0-9-]+$`).MatchString
	if !isAlphaNumeric(req.Name) {
		return nil, ErrInvalidWorkspaceName
	}

	w := &workspace.Workspace{
		Id:     req.Id,
		Name:   req.Name,
		Target: req.Target,
	}

	apiKey, err := s.apiKeyService.Generate(apikey.ApiKeyTypeWorkspace, w.Id)
	if err != nil {
		return nil, err
	}
	w.ApiKey = apiKey

	w.Projects = []*project.Project{}

	for _, p := range req.Projects {
		var projectConfig *config.ProjectConfig
		if p.NewProjectConfig == nil && p.ExistingProjectConfig == nil {
			return nil, ErrInvalidProjectConfig
		}

		if p.NewProjectConfig != nil {
			// todo: make util that turns create project config into project config
			fmt.Println(projectConfig)
		}

		isValidProjectName := regexp.MustCompile(`^[a-zA-Z0-9-_.]+$`).MatchString
		if !isValidProjectName(p.NewProjectConfig.Name) {
			return nil, ErrInvalidProjectName
		}

		if p.NewProjectConfig.Source.Repository != nil && p.NewProjectConfig.Source.Repository.Sha == "" {
			sha, err := s.gitProviderService.GetLastCommitSha(p.NewProjectConfig.Source.Repository)
			if err != nil {
				return nil, err
			}
			p.NewProjectConfig.Source.Repository.Sha = sha
		}

		apiKey, err := s.apiKeyService.Generate(apikey.ApiKeyTypeProject, fmt.Sprintf("%s/%s", w.Id, p.NewProjectConfig.Name))
		if err != nil {
			return nil, err
		}

		var pc *config.ProjectConfig

		if p.ExistingProjectConfig == nil {
			projectImage := s.defaultProjectImage
			if p.NewProjectConfig.Image != nil {
				projectImage = *p.NewProjectConfig.Image
			}

			projectUser := s.defaultProjectUser
			if p.NewProjectConfig.User != nil {
				projectUser = *p.NewProjectConfig.User
			}

			pc = &config.ProjectConfig{
				Name:       p.NewProjectConfig.Name,
				Image:      projectImage,
				User:       projectUser,
				Build:      p.NewProjectConfig.Build,
				Repository: p.NewProjectConfig.Source.Repository,
				EnvVars:    p.NewProjectConfig.EnvVars,
			}
		} else {
			pc, err = s.projectConfigService.Find(p.ExistingProjectConfig.ConfigName)
			if err != nil {
				return nil, err
			}
			pc.Repository.Branch = &p.ExistingProjectConfig.Branch
		}

		p := &project.Project{
			ProjectConfig: *pc,
			WorkspaceId:   w.Id,
			ApiKey:        apiKey,
			Target:        w.Target,
		}
		w.Projects = append(w.Projects, p)
	}

	err = s.workspaceStore.Save(w)
	if err != nil {
		return nil, err
	}

	return s.createWorkspace(w)
}

func (s *WorkspaceService) createBuild(p *project.Project, gc *gitprovider.GitProviderConfig, logWriter io.Writer) (*project.Project, error) {
	// FIXME: skip build completely for now
	return p, nil

	if p.Build != nil { // nolint:govet
		lastBuildResult, err := s.builderFactory.CheckExistingBuild(*p)
		if err != nil {
			return nil, err
		}
		if lastBuildResult != nil {
			p.Image = lastBuildResult.ImageName
			p.User = lastBuildResult.User
			return p, nil
		}

		builder, err := s.builderFactory.Create(*p, gc)
		if err != nil {
			return nil, err
		}

		if builder == nil {
			return p, nil
		}

		buildResult, err := builder.Build()
		if err != nil {
			s.handleBuildError(p, builder, logWriter, err)
			return p, nil
		}

		err = builder.Publish()
		if err != nil {
			s.handleBuildError(p, builder, logWriter, err)
			return p, nil
		}

		err = builder.SaveBuildResults(*buildResult)
		if err != nil {
			s.handleBuildError(p, builder, logWriter, err)
			return p, nil
		}

		err = builder.CleanUp()
		if err != nil {
			logWriter.Write([]byte(fmt.Sprintf("Error cleaning up build: %s\n", err.Error())))
		}

		p.Image = buildResult.ImageName
		p.User = buildResult.User

		return p, nil
	}

	return p, nil
}

func (s *WorkspaceService) createProject(p *project.Project, target *provider.ProviderTarget, logWriter io.Writer) error {
	logWriter.Write([]byte(fmt.Sprintf("Creating project %s\n", p.Name)))

	cr, err := s.containerRegistryService.FindByImageName(p.Image)
	if err != nil && !containerregistry.IsContainerRegistryNotFound(err) {
		return err
	}

	gc, err := s.gitProviderService.GetConfigForUrl(p.Repository.Url)
	if err != nil && !gitprovider.IsGitProviderNotFound(err) {
		return err
	}

	err = s.provisioner.CreateProject(p, target, cr, gc)
	if err != nil {
		return err
	}

	logWriter.Write([]byte(fmt.Sprintf("Project %s created\n", p.Name)))

	return nil
}

func (s *WorkspaceService) createWorkspace(ws *workspace.Workspace) (*workspace.Workspace, error) {
	target, err := s.targetStore.Find(ws.Target)
	if err != nil {
		return ws, err
	}

	wsLogger := s.loggerFactory.CreateWorkspaceLogger(ws.Id, logs.LogSourceServer)
	defer wsLogger.Close()

	wsLogger.Write([]byte(fmt.Sprintf("Creating workspace %s (%s)\n", ws.Name, ws.Id)))

	err = s.provisioner.CreateWorkspace(ws, target)
	if err != nil {
		return nil, err
	}

	for i, p := range ws.Projects {
		projectLogger := s.loggerFactory.CreateProjectLogger(ws.Id, p.Name, logs.LogSourceServer)
		defer projectLogger.Close()

		gc, _ := s.gitProviderService.GetConfigForUrl(p.Repository.Url)

		projectWithEnv := *p
		projectWithEnv.EnvVars = project.GetProjectEnvVars(p, s.serverApiUrl, s.serverUrl)

		for k, v := range p.EnvVars {
			projectWithEnv.EnvVars[k] = v
		}

		var err error

		p, err = s.createBuild(&projectWithEnv, gc, projectLogger)
		if err != nil {
			return nil, err
		}

		ws.Projects[i] = p
		err = s.workspaceStore.Save(ws)
		if err != nil {
			return nil, err
		}

		err = s.createProject(p, target, projectLogger)
		if err != nil {
			return nil, err
		}
	}

	wsLogger.Write([]byte("Workspace creation complete. Pending start...\n"))

	err = s.startWorkspace(ws, target, wsLogger)
	if err != nil {
		return nil, err
	}

	return ws, nil
}

func (s *WorkspaceService) handleBuildError(p *project.Project, builder build.IBuilder, logWriter io.Writer, err error) {
	logWriter.Write([]byte("################################################\n"))
	logWriter.Write([]byte(fmt.Sprintf("#### BUILD FAILED FOR PROJECT %s: %s\n", p.Name, err.Error())))
	logWriter.Write([]byte("################################################\n"))

	cleanupErr := builder.CleanUp()
	if cleanupErr != nil {
		logWriter.Write([]byte(fmt.Sprintf("Error cleaning up build: %s\n", cleanupErr.Error())))
	}

	logWriter.Write([]byte("Creating project with default image\n"))
	p.Image = s.defaultProjectImage
	p.User = s.defaultProjectUser
}
