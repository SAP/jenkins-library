# ABAP Environment Pipeline

The goal of the ABAP Environment Pipeline is to enable Continuous Integration for the SAP Cloud Platform ABAP Environment, als known as Steampunk.
In the current state, the pipeline enables you to pull your Software Components to specifc systems and perform ATC checks.

## Configuration

1. Configure your Jenkins Server according to the [documentation](https://sap.github.io/jenkins-library/guidedtour/)
2. Create a file named `Jenkinsfile` in your repository with the following content:

```
@Library('piper-lib-os') _

abapEnvironmentPipeline script: this
```

The annotation `@Library('piper-lib-os')` is a reference to the Jenkins Configuration, where you configured the Piper Library as a "Global Pipeline Library". If you want to **avoid breaking changes** we advise you to use a specific release of the Piper Library instead of the default master branch (see https://sap.github.io/jenkins-library/customjenkins/#shared-library)

3. Create a file `manifest.yml`. The pipeline will

4. Create a file `.pipeline/config.yml` where you store the configuration for the pipeline, e.g. apiEndpoints and credentialIds. The steps make use of the Credentials Store of the Jenkins Server. Here is an example of the configuration file:
```

```
