// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"context"
	"errors"
	"log"
	"net/url"

	config_const "github.com/daytonaio/daytona/cmd/daytona/config"
	apiclient_util "github.com/daytonaio/daytona/internal/util/apiclient"
	"github.com/daytonaio/daytona/pkg/apiclient"
	gitprovider_view "github.com/daytonaio/daytona/pkg/views/gitprovider"
	views_util "github.com/daytonaio/daytona/pkg/views/util"
	"github.com/daytonaio/daytona/pkg/views/workspace/create"
	"github.com/daytonaio/daytona/pkg/views/workspace/selection"
)

type RepositoryWizardConfig struct {
	ApiClient            *apiclient.APIClient
	UserGitProviders     []apiclient.GitProvider
	MultiProject         bool
	SkipBranchSelection  bool
	ProjectOrder         int
	DisabledGitProviders map[string]bool
	DisabledNamespaces   map[string]bool
	SelectedRepos        map[string]bool
}

func getRepositoryFromWizard(config RepositoryWizardConfig) (*apiclient.GitRepository, error) {
	var providerId string
	var namespaceId string
	var checkoutOptions []selection.CheckoutOption

	ctx := context.Background()

	if len(config.UserGitProviders) == 0 {
		return create.GetRepositoryFromUrlInput(config.MultiProject, config.ProjectOrder, config.ApiClient, config.SelectedRepos)
	}

	supportedProviders := config_const.GetSupportedGitProviders()
	var gitProviderViewList []gitprovider_view.GitProviderView

	for _, gitProvider := range config.UserGitProviders {
		for _, supportedProvider := range supportedProviders {
			if *gitProvider.Id == supportedProvider.Id {
				gitProviderViewList = append(gitProviderViewList,
					gitprovider_view.GitProviderView{
						Id:       *gitProvider.Id,
						Name:     supportedProvider.Name,
						Username: *gitProvider.Username,
					},
				)
			}
		}
	}
	providerId = selection.GetProviderIdFromPrompt(gitProviderViewList, config.ProjectOrder, config.DisabledGitProviders)
	if providerId == "" {
		return nil, errors.New("must select a provider")
	}

	if providerId == selection.CustomRepoIdentifier {
		return create.GetRepositoryFromUrlInput(config.MultiProject, config.ProjectOrder, config.ApiClient, config.SelectedRepos)
	}

	ApiClient, err := apiclient_util.GetApiClient(nil)
	if err != nil {
		log.Fatal(err)
	}

	var namespaceList []apiclient.GitNamespace

	err = views_util.WithSpinner(func() error {
		namespaceList, _, err = ApiClient.GitProviderAPI.GetNamespaces(ctx, providerId).Execute()
		return err
	})
	if err != nil {
		return nil, err
	}

	if len(namespaceList) == 1 {
		namespaceId = *namespaceList[0].Id
	} else {
		namespaceId = selection.GetNamespaceIdFromPrompt(namespaceList, config.ProjectOrder, providerId, config.DisabledGitProviders, config.DisabledNamespaces)
		if namespaceId == "" {
			return nil, errors.New("namespace not found")
		}
	}

	var providerRepos []apiclient.GitRepository
	err = views_util.WithSpinner(func() error {
		providerRepos, _, err = ApiClient.GitProviderAPI.GetRepositories(ctx, providerId, namespaceId).Execute()
		return err
	})

	if err != nil {
		return nil, err
	}

	var chosenRepo *apiclient.GitRepository
	if len(namespaceList) > 1 {
		chosenRepo = selection.GetRepositoryFromPrompt(providerRepos, config.ProjectOrder, namespaceId, config.DisabledNamespaces, config.SelectedRepos)
	} else {
		chosenRepo = selection.GetRepositoryFromPrompt(providerRepos, config.ProjectOrder, providerId, config.DisabledGitProviders, config.SelectedRepos)
	}
	if chosenRepo == nil {
		return nil, errors.New("must select a repository")
	}

	if config.SkipBranchSelection {
		return chosenRepo, nil
	}

	var branchList []apiclient.GitBranch
	err = views_util.WithSpinner(func() error {
		branchList, _, err = ApiClient.GitProviderAPI.GetRepoBranches(ctx, providerId, namespaceId, url.QueryEscape(*chosenRepo.Id)).Execute()
		return err
	})

	if err != nil {
		return nil, err
	}

	if len(branchList) == 0 {
		return nil, errors.New("no branches found")
	}

	if len(branchList) == 1 {
		chosenRepo.Branch = branchList[0].Name
		chosenRepo.Sha = branchList[0].Sha
		return chosenRepo, nil
	}

	var prList []apiclient.GitPullRequest
	err = views_util.WithSpinner(func() error {
		prList, _, err = ApiClient.GitProviderAPI.GetRepoPRs(ctx, providerId, namespaceId, url.QueryEscape(*chosenRepo.Id)).Execute()
		return err
	})

	if err != nil {
		return nil, err
	}

	var branch *apiclient.GitBranch
	if len(prList) == 0 {
		branch = selection.GetBranchFromPrompt(branchList, config.ProjectOrder)
		if branch == nil {
			return nil, errors.New("must select a branch")
		}

		chosenRepo.Branch = branch.Name
		chosenRepo.Sha = branch.Sha

		return chosenRepo, nil
	}

	checkoutOptions = append(checkoutOptions, selection.CheckoutDefault)
	checkoutOptions = append(checkoutOptions, selection.CheckoutBranch)
	checkoutOptions = append(checkoutOptions, selection.CheckoutPR)

	chosenCheckoutOption := selection.GetCheckoutOptionFromPrompt(config.ProjectOrder, checkoutOptions)
	if chosenCheckoutOption == selection.CheckoutDefault {
		return chosenRepo, nil
	}

	if chosenCheckoutOption == selection.CheckoutBranch {
		branch = selection.GetBranchFromPrompt(branchList, config.ProjectOrder)
		if branch == nil {
			return nil, errors.New("must select a branch")
		}
		chosenRepo.Branch = branch.Name
		chosenRepo.Sha = branch.Sha
	} else if chosenCheckoutOption == selection.CheckoutPR {
		chosenPullRequest := selection.GetPullRequestFromPrompt(prList, config.ProjectOrder)
		if chosenPullRequest == nil {
			return nil, errors.New("must select a pull request")
		}

		chosenRepo.Branch = chosenPullRequest.Branch
		chosenRepo.Sha = chosenPullRequest.Sha
		chosenRepo.Id = chosenPullRequest.SourceRepoId
		chosenRepo.Name = chosenPullRequest.SourceRepoName
		chosenRepo.Owner = chosenPullRequest.SourceRepoOwner
		chosenRepo.Url = chosenPullRequest.SourceRepoUrl
	}

	return chosenRepo, nil
}

func getBranchFromWizard(config RepositoryWizardConfig) (*apiclient.GitRepository, error) {
	return nil, nil
}
