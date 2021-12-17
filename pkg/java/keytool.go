package java

import (
	"os"
	"path/filepath"

	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
)

const defaultTruststorePath = "lib/security/cacerts"
const DefaultTruststorePassword = "changeit"

func GetDefaultTruststorePath() string {
	return filepath.Join(os.Getenv("JAVA_HOME"), filepath.FromSlash(defaultTruststorePath))
}

func GetMavenOpts(truststoreFile string) string {
	return "-Djavax.net.ssl.trustStore=" + truststoreFile + " -Djavax.net.ssl.trustStorePassword=" + DefaultTruststorePassword
}

func ImportCert(runner command.ExecRunner, destTruststore, certfile string) error {
	options := []string{
		"-import",
		"-noprompt",
		"-storepass", DefaultTruststorePassword,
		"-keystore", destTruststore,
		"-file", certfile,
		"-alias", filepath.Base(certfile),
	}
	log.Entry().Infof("Importing certificate: %s", certfile)
	return runner.RunExecutable("keytool", options...)
}

func ImportTruststore(runner command.ExecRunner, destTruststore, srcTruststore string) error {
	options := []string{
		"-importkeystore",
		"-noprompt",
		"-srckeystore", srcTruststore,
		"-srcstorepass", DefaultTruststorePassword,
		"-destkeystore", destTruststore,
		"-deststorepass", DefaultTruststorePassword,
	}
	log.Entry().Debugf("Copying existing trust store: %s", srcTruststore)
	return runner.RunExecutable("keytool", options...)
}
