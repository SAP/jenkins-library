# ${docGenStepName}

## ${docGenDescription}

## Prerequisites

You need to store the API token for the Detect service as _'Secret text'_ credential in your Jenkins system.

## ${docJenkinsPluginDependencies}

## ${docGenParameters}

## ${docGenConfiguration}

## Rapid scan

In addition to the full scan, Black Duck also offers a faster and easier scan option, called <a href="https://community.synopsys.com/s/document-item?bundleId=integrations-detect&topicId=downloadingandrunning%2Frapidscan.html&_LANG=enus" target="_blank">Rapid Scan</a>. Its main advantage is speed. In most cases, the scan is completed in less than 30 seconds.
It doesn't save any inormation in Black Duck side. The result can be found in console on pipeline. By default blackduck scans in 'FULL' mode, but you are able to change scan mode by parameter `scanMode`.
When you set `scanMode='RAPID'` in the DetectExecution step, then Black Duck scans in 'Rapid' mode.

### Rapid scan  on pull requests
If the Jenkins orchestrator is configured to detect pull requests, then piper pipeline can recognize this
and change the Black Duck scan mode from 'FULL' to 'RAPID'. This does not effect to usual branch scans.

#### Result of scan on pull request comment
If `githubApi` and `githubToken` are provided, then pipeline adds the scan result to the comment of the opened pull request.
As shown here:
![Pull request commnet](../images/BDrapidScanPrs.png)


#### Steps to achive this:
1. Insert piper-lib in your Jenkins orchestrator
2. Configure your Jenkins orchestrator to detect pull requests
3. Provide all needed parameters of DetectExecution step (inc.`githubApi`, `githubToken`)
4. Open a pull request with some changes to main branch


Note: Despite rapid scans doing necessary security checks for daily development, it is not sufficient for production deployment and releases. Only use full scans for production deployment and releases.

