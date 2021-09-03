package ado

import (
	"context"
	"errors"
	"fmt"

	"github.com/microsoft/azure-devops-go-api/azuredevops"
	"github.com/microsoft/azure-devops-go-api/azuredevops/build"
)

const azureUrl = "https://dev.azure.com"

type BuildClient interface {
	UpdateVariables(variables []Variable) error
}

type BuildClientImpl struct {
	ctx         context.Context
	buildClient build.Client
	project     string
	pipelineID  int
}

type Variable struct {
	Name          string
	Value         string
	IsSecret      bool
	AllowOverride bool
}

//UpdateVariables updates variables in build definition or creates them if they are missing
func (bc *BuildClientImpl) UpdateVariables(variables []Variable) error {
	if len(variables) == 0 {
		return errors.New("error: variables must not be empty")
	}
	getDefinitionArgs := build.GetDefinitionArgs{
		Project:      &bc.project,
		DefinitionId: &bc.pipelineID,
	}

	// Get a build definition
	buildDefinition, err := bc.buildClient.GetDefinition(bc.ctx, getDefinitionArgs)
	if err != nil {
		return err
	}

	buildDefinitionVars := *buildDefinition.Variables
	if buildDefinition.Variables == nil {
		buildDefinitionVars = map[string]build.BuildDefinitionVariable{}
	}

	for _, variable := range variables {
		buildDefinitionVars[variable.Name] = build.BuildDefinitionVariable{
			Value:         &variable.Value,
			IsSecret:      &variable.IsSecret,
			AllowOverride: &variable.AllowOverride,
		}
	}

	buildDefinition.Variables = &buildDefinitionVars

	updateDefinitionArgs := build.UpdateDefinitionArgs{
		Definition:   buildDefinition,
		Project:      &bc.project,
		DefinitionId: &bc.pipelineID,
	}

	_, err = bc.buildClient.UpdateDefinition(bc.ctx, updateDefinitionArgs)
	if err != nil {
		return err
	}

	return nil
}

//NewBuildClient Create a client to interact with the Build area
func NewBuildClient(organization string, personalAccessToken string, project string, pipelineID int) (BuildClient, error) {
	if err := validateArgs(organization, personalAccessToken, project); err != nil {
		return nil, err
	}

	organizationUrl := fmt.Sprintf("%s/%s", azureUrl, organization)
	// Create a connection to your organization
	connection := azuredevops.NewPatConnection(organizationUrl, personalAccessToken)

	ctx := context.Background()

	// Create a client to interact with the Core area
	buildClient, err := build.NewClient(ctx, connection)
	if err != nil {
		return nil, err
	}

	buildClientImpl := &BuildClientImpl{
		ctx:         ctx,
		buildClient: buildClient,
		project:     project,
		pipelineID:  pipelineID,
	}

	return buildClientImpl, nil
}

func validateArgs(org string, pat string, project string) error {
	if org == "" {
		return errors.New("error: organization must not be empty")
	}
	if pat == "" {
		return errors.New("error: personal access token must not be empty")
	}
	if project == "" {
		return errors.New("error: project must not be empty")
	}
	return nil
}
