# ${docGenStepName}

## ${docGenDescription}

## Prerequisites

You need to store the API token for the Detect service as _'Secret text'_ credential in your Jenkins system.

## ${docJenkinsPluginDependencies}

## ${docGenParameters}

## ${docGenConfiguration}

## Rapid scan - Pending: Feb 6th, 2023

In addition to the full scan, Black Duck also offers a faster and easier scan option, called "Rapid Scan". Its main advantage is speed. In most cases, the scan is completed in less than 30 seconds.



If the orchestrator (Jenkins, Azure, Github Actions) is configured to detect pull requests, the pipeline can recognize this and change the Black Duck scan mode from "Full" to "Rapid".

If `githubApi` and `githubToken` are provided, the pipeline adds the scan result to the comment of the opened pull request.



There is also a parameter, `scanMode`, for the DetectExecution step, which you can use to change the scan mode to "Rapid".

Note: Despite rapid scans doing necessary security checks for daily development, it is not sufficient for production deployment and releases. Only use full scans for production deployment and releases.
