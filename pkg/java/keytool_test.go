package java

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	piperMock "github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGetDefaultTruststorePath(t *testing.T) {
	// prepare
	os.Setenv("JAVA_HOME", mock.Anything)
	require.Equal(t, mock.Anything, os.Getenv("JAVA_HOME"))
	// test
	result := GetDefaultTruststorePath()
	// assert
	assert.Equal(t, "lib/security/cacerts", defaultTruststorePath)
	assert.Equal(t, filepath.Join(mock.Anything, defaultTruststorePath), result)
	// restore
	os.Unsetenv("JAVA_HOME")
}

func TestGetMavenOpts(t *testing.T) {
	// test
	result := GetMavenOpts(mock.Anything)
	// assert
	assert.Equal(t, "changeit", DefaultTruststorePassword)
	assert.Equal(t, "-Djavax.net.ssl.trustStore="+mock.Anything+" -Djavax.net.ssl.trustStorePassword="+DefaultTruststorePassword, result)
}

func TestImportCert(t *testing.T) {
	// prepare
	secretstorePath := filepath.Join(mock.Anything, mock.Anything)
	mockRunner := &piperMock.ExecMockRunner{}
	// test
	err := ImportCert(mockRunner, mock.Anything, secretstorePath)
	// assert
	assert.NoError(t, err)
	assert.Len(t, mockRunner.Calls, 1)
	for _, call := range mockRunner.Calls {
		assert.Equal(t, "keytool", call.Exec)
		assert.Equal(t, strings.Join(call.Params, " "), "-import -noprompt -storepass changeit -keystore mock.Anything -file "+secretstorePath+" -alias mock.Anything")
	}
}

func TestImportTruststore(t *testing.T) {
	// prepare
	srcSecretstorePath := filepath.Join(mock.Anything, mock.Anything)
	destSecretstorePath := filepath.Join(mock.Anything, mock.Anything)
	mockRunner := &piperMock.ExecMockRunner{}
	// test
	err := ImportTruststore(mockRunner, destSecretstorePath, srcSecretstorePath)
	// assert
	assert.NoError(t, err)
	assert.Len(t, mockRunner.Calls, 1)
	for _, call := range mockRunner.Calls {
		assert.Equal(t, "keytool", call.Exec)
		assert.Equal(t, strings.Join(call.Params, " "), "-importkeystore -noprompt -srckeystore "+srcSecretstorePath+" -srcstorepass changeit -destkeystore "+destSecretstorePath+" -deststorepass changeit")
	}
}
