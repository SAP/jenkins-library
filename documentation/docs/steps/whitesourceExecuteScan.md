# whitesourceExecuteScan

## Description

With this step [WhiteSource](https://www.whitesourcesoftware.com) security and license compliance scans can be executed and assessed.

WhiteSource is a Software as a Service offering based on a so called unified agent that locally determines the dependency
tree of a node.js, Java, Python, Ruby, or Scala based solution and sends it to the WhiteSource server for a policy based license compliance
check and additional Free and Open Source Software Publicly Known Vulnerabilities detection.

!!! note "Docker Images"
    The underlying Docker images are public and specific to the solution's programming language(s) and may therefore be exchanged
    to fit and suite the relevant scenario. The default Python environment used is i.e. Python 3 based.

## Prerequisites

## Parameters

| name | mandatory | default | possible values |
|------|-----------|---------|-----------------|
| `agentDownloadUrl` | no | `https://github.com/whitesource/unified-agent-distribution/raw/master/standAlone/${config.agentFileName}` |  |
| `agentFileName` | no | `wss-unified-agent.jar` |  |
| `agentParameters` | no |  |  |
| `buildDescriptorExcludeList` | no |  |  |
| `buildDescriptorFile` | no |  |  |
| `configFilePath` | no | `./wss-unified-agent.config` |  |
| `createProductFromPipeline` | no | `true` |  |
| `cvssSeverityLimit` | no | `-1` | `-1` to switch failing off, any `positive integer between 0 and 10` to fail on issues with the specified limit or above |
| `dockerImage` | no |  |  |
| `dockerWorkspace` | no |  |  |
| `emailAddressesOfInitialProductAdmins` | no |  |  |
| `jreDownloadUrl` | no | `https://github.com/SAP/SapMachine/releases/download/sapmachine-11.0.2/sapmachine-jre-11.0.2_linux-x64_bin.tar.gz` |  |
| `licensingVulnerabilities` | no | `true` | `true`, `false` |
| `orgAdminUserTokenCredentialsId` | no |  |  |
| `orgToken` | yes |  |  |
| `parallelLimit` | no | `15` |  |
| `productName` | yes |  |  |
| `productToken` | no |  |  |
| `productVersion` | no |  |  |
| `projectNames` | no |  |  |
| `reporting` | no | `true` | `true`, `false` |
| `scanType` | no |  | `maven`, `mta`, `npm`, `pip`, `sbt` |
| `script` | yes |  |  |
| `securityVulnerabilities` | no | `true` | `true`, `false` |
| `serviceUrl` | yes |  |  |
| `stashContent` | no |  |  |
| `timeout` | no |  |  |
| `userTokenCredentialsId` | yes |  |  |
| `verbose` | no |  | `true`, `false` |
| `vulnerabilityReportFileName` | no | `piper_whitesource_vulnerability_report` |  |
| `vulnerabilityReportTitle` | no | `WhiteSource Security Vulnerability Report` |  |

* `agentDownloadUrl` - URL used to download the latest version of the WhiteSource Unified Agent.
* `agentFileName` - Locally used name for the Unified Agent jar file after download.
* `agentParameters` - Additional parameters passed to the Unified Agent command line.
* `buildDescriptorExcludeList` - List of build descriptors and therefore modules to exclude from the scan and assessment activities.
* `buildDescriptorFile` - Explicit path to the build descriptor file.
* `configFilePath` - Explicit path to the WhiteSource Unified Agent configuration file.
* `createProductFromPipeline` - Whether to create the related WhiteSource product on the fly based on the supplied pipeline configuration.
* `cvssSeverityLimit` - Limit of tollerable CVSS v3 score upon assessment and in consequence fails the build, defaults to  `-1`.
* `dockerImage` - Docker image to be used for scanning.
* `dockerWorkspace` - Docker workspace to be used for scanning.
* `emailAddressesOfInitialProductAdmins` - The list of email addresses to assign as product admins for newly created WhiteSource products.
* `jreDownloadUrl` - URL used for downloading the Java Runtime Environment (JRE) required to run the WhiteSource Unified Agent.
* `licensingVulnerabilities` - Whether license compliance is considered and reported as part of the assessment.
* `orgAdminUserTokenCredentialsId` - Jenkins credentials ID referring to the organization admin's token.
* `orgToken` - WhiteSource token identifying your organization.
* `parallelLimit` - Limit of parallel jobs being run at once in case of `scanType: 'mta'` based scenarios, defaults to `15`.
* `productName` - Name of the WhiteSource product to be created and used for results aggregation.
* `productToken` - Token of the WhiteSource product to be created and used for results aggregation, usually determined automatically.
* `productVersion` - Version of the WhiteSource product to be created and used for results aggregation, usually determined automatically.
* `projectNames` - List of WhiteSource projects to be included in the assessment part of the step, usually determined automatically.
* `reporting` - Whether assessment is being done at all, defaults to `true`.
* `scanType` - Type of development stack used to implement the solution.
* `script` - The common script environment of the Jenkinsfile running. Typically the reference to the script calling the pipeline step is provided with the this parameter, as in `script: this`. This allows the function to access the commonPipelineEnvironment for retrieving, for example, configuration parameters.
* `securityVulnerabilities` - Whether security compliance is considered and reported as part of the assessment.
* `serviceUrl` - URL to the WhiteSource server API used for communication, defaults to `https://saas.whitesourcesoftware.com/api`.
* `stashContent` - List of stashes to be unstashed into the workspace before performing the scan.
* `timeout` - Timeout in seconds until a HTTP call is forcefully terminated.
* `userTokenCredentialsId` - Jenkins credentials ID referring to the product admin's token.
* `verbose` - Whether verbose output should be produced.
* `vulnerabilityReportFileName` - Name of the file the vulnerability report is written to.
* `vulnerabilityReportTitle` - Title of vulnerability report written during the assessment phase.

## Step configuration

We recommend to define values of step parameters via [config.yml file](../configuration.md).

In following sections of the config.yml the configuration is possible:

| parameter | general | step | stage |
|-----------|---------|------|-------|
| `agentDownloadUrl` |  | X | X |
| `agentFileName` |  | X | X |
| `agentParameters` |  | X | X |
| `buildDescriptorExcludeList` |  | X | X |
| `buildDescriptorFile` |  | X | X |
| `configFilePath` |  | X | X |
| `createProductFromPipeline` |  | X | X |
| `cvssSeverityLimit` |  | X | X |
| `dockerImage` |  | X | X |
| `dockerWorkspace` |  | X | X |
| `emailAddressesOfInitialProductAdmins` |  | X | X |
| `jreDownloadUrl` |  | X | X |
| `licensingVulnerabilities` |  | X | X |
| `orgAdminUserTokenCredentialsId` | X | X | X |
| `orgToken` | X | X | X |
| `parallelLimit` |  | X | X |
| `productName` | X | X | X |
| `productToken` | X | X | X |
| `productVersion` | X | X | X |
| `projectNames` | X | X | X |
| `reporting` |  | X | X |
| `scanType` | X | X | X |
| `script` |  |  |  |
| `securityVulnerabilities` |  | X | X |
| `serviceUrl` | X | X | X |
| `stashContent` |  | X | X |
| `timeout` |  | X | X |
| `userTokenCredentialsId` | X | X | X |
| `verbose` | X | X | X |
| `vulnerabilityReportFileName` |  | X | X |
| `vulnerabilityReportTitle` |  | X | X |

## Exceptions

## Examples
