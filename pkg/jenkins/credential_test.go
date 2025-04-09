//go:build unit

package jenkins

import (
	"context"
	"testing"

	"github.com/SAP/jenkins-library/pkg/jenkins/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUpdateCredential(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	const ID = "testID"
	const testSecret = "testSecret"
	const domain = "_"

	t.Run("That secret is updated", func(t *testing.T) {
		credManagerMock := mocks.CredentialsManager{}
		testCredential := StringCredentials{ID: ID, Secret: testSecret}

		credManagerMock.On("Update", ctx, domain, ID, mock.Anything).Return(nil)
		err := UpdateCredential(ctx, &credManagerMock, domain, testCredential)
		credManagerMock.AssertCalled(t, "Update", ctx, domain, ID, testCredential)

		assert.NoError(t, err)
	})

	t.Run("Test that wrong credential type fails ", func(t *testing.T) {
		credManagerMock := mocks.CredentialsManager{}

		credManagerMock.On("Update", ctx, domain, ID, mock.Anything).Return(nil)
		err := UpdateCredential(ctx, &credManagerMock, domain, 5)
		credManagerMock.AssertNotCalled(t, "Update", ctx, domain, ID, mock.Anything)
		assert.EqualError(t, err, "'credential' parameter is supposed to be a Credentials struct not 'int'")
	})

	t.Run("Test that wrong credential type fails ", func(t *testing.T) {
		credManagerMock := mocks.CredentialsManager{}
		testCredential := struct{ Secret string }{
			Secret: "Test",
		}

		credManagerMock.On("Update", ctx, domain, ID, mock.Anything).Return(nil)
		err := UpdateCredential(ctx, &credManagerMock, domain, testCredential)
		credManagerMock.AssertNotCalled(t, "Update", ctx, domain, ID, mock.Anything)
		assert.EqualError(t, err, "'credential' parameter is supposed to be a Credentials struct not 'struct { Secret string }'")
	})

	t.Run("Test that empty secret id fails ", func(t *testing.T) {
		credManagerMock := mocks.CredentialsManager{}
		testCredential := StringCredentials{ID: "", Secret: testSecret}

		credManagerMock.On("Update", ctx, domain, ID, mock.Anything).Return(nil)
		err := UpdateCredential(ctx, &credManagerMock, domain, testCredential)
		credManagerMock.AssertNotCalled(t, "Update", ctx, domain, ID, mock.Anything)
		assert.EqualError(t, err, "secret ID should not be empty")
	})

}
