# ${docGenStepName}

## ${docGenDescription}

## Prerequisites

You need to store the API token for the Detect service as _'Secret text'_ credential in your Jenkins system.

## ${docJenkinsPluginDependencies}

## ${docGenParameters}

## ${docGenConfiguration}

## Rapid scan

In addition to the full scan, Black Duck also offers a faster and easier scan option, called <a href="https://community.synopsys.com/s/document-item?bundleId=integrations-detect&topicId=downloadingandrunning%2Frapidscan.html&_LANG=enus" target="_blank">Rapid Scan</a>.
Its main advantage is speed. In most cases, the scan is completed in less than 30 seconds. It doesn't save any information in Black Duck side.
The result can be found in console on pipeline. By default, black duck scans in 'FULL' mode.

### Rapid scan on pull requests

If the orchestrator is configured to detect pull requests, then piper pipeline in detecExecuationScan step can recognize the pull request and change the Black Duck scan mode from 'FULL' to 'RAPID'. This does not affect to usual branch scans.

- **Note**
  1. This functionality is not applicable to GPP (General Purpose Pipeline)
  2. This can be used only for custom pipelines based on Jenkins piper library

#### Result of scan on pull request comment

If `githubApi` and `githubToken` are provided, then pipeline adds the scan result to the comment of the opened pull request.

![blackDuckPullRequestComment](../images/BDRapidScanPrs.png)

#### Steps to achieve this

1. Specify all required parameters of the DetectExecution step in .pipeline/config.yaml (`githubApi`, `githubToken` optional)
2. Enable detecExecuationScan in the orchestrator
3. Specify `githubApi` and `githubToken` in the DetectExecution step to get the result in the pull request comment. (optional)
4. Open a pull request with some changes to main branch

#### Example for jenkins orchestrator

In Jenkinsfile

```
@Library('piper-lib') _
@Library('piper-lib-os') __

node {
  stage('Init') {
    checkout scm
    setupPipelineEnvironment script: this
  }
  stage('detectExecuteScan') {
     detectExecuteScan script: this
  }
  ...
}
```

In config.yml

```
...
steps:
  ...
  detectExecuteScan:
    serverUrl: 'https://sap-staging.app.blackduck.com/'
    detectTokenCredentialsId: 'JenkinsCredentialsIdForBlackDuckToken'
    projectName: 'projectNameInBlackDuckUI'
    version: 'v1.0'
    githubApiUrl: 'https://github.wdf.sap.corp/api/v3'
    githubToken: 'JenkinsCredentialsIdForGithub'
  ...
...
```

**Note**: Despite rapid scans doing necessary security checks for daily development, it is not sufficient for production deployment and releases.
Only use full scans for production deployment and releases.
