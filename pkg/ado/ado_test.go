package ado

import (
	"context"
	"errors"
	"testing"

	"github.com/SAP/jenkins-library/pkg/ado/mocks"
	"github.com/microsoft/azure-devops-go-api/azuredevops/build"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUpdateVariables(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	const secretName = "test-secret"
	const secretValue = "secret-value"
	const projectID = "some-id"
	const pipelineID = 1
	testErr := errors.New("error")

	tests := []struct {
		name                  string
		variables             []Variable
		getDefinitionError    error
		updateDefinitionError error
		isErrorExpected       bool
		errorStr              string
	}{
		{
			name:                  "Test update secret - successful",
			variables:             []Variable{{Name: secretName, Value: secretValue, IsSecret: true}},
			getDefinitionError:    nil,
			updateDefinitionError: nil,
			isErrorExpected:       false,
		},
		{
			name:                  "Failed get definition",
			variables:             []Variable{{Name: secretName, Value: secretValue, IsSecret: true}},
			getDefinitionError:    testErr,
			updateDefinitionError: nil,
			isErrorExpected:       true,
			errorStr:              "get definition failed",
		},
		{
			name:                  "Failed update definition",
			variables:             []Variable{{Name: secretName, Value: secretValue, IsSecret: true}},
			getDefinitionError:    nil,
			updateDefinitionError: testErr,
			isErrorExpected:       true,
			errorStr:              "update definition failed",
		},
		{
			name:                  "Slice variables is empty",
			variables:             []Variable{},
			getDefinitionError:    nil,
			updateDefinitionError: nil,
			isErrorExpected:       true,
			errorStr:              "slice variables must not be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buildClientMock := &mocks.Client{}
			buildClientMock.On("GetDefinition", ctx, mock.Anything).Return(
				func(ctx context.Context, getDefinitionArgs build.GetDefinitionArgs) *build.BuildDefinition {
					return &build.BuildDefinition{}
				},
				func(ctx context.Context, getDefinitionArgs build.GetDefinitionArgs) error {
					return tt.getDefinitionError
				},
			)

			buildClientMock.On("UpdateDefinition", ctx, mock.Anything).Return(
				func(ctx context.Context, updateDefinitionArgs build.UpdateDefinitionArgs) *build.BuildDefinition {
					return &build.BuildDefinition{}
				},
				func(ctx context.Context, updateDefinitionArgs build.UpdateDefinitionArgs) error {
					return tt.updateDefinitionError
				},
			)

			buildClientImpl := BuildClientImpl{
				ctx:         ctx,
				buildClient: buildClientMock,
				project:     projectID,
				pipelineID:  pipelineID,
			}

			err := buildClientImpl.UpdateVariables(tt.variables)
			if tt.isErrorExpected {
				assert.Contains(t, err.Error(), tt.errorStr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
