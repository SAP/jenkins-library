# ${docGenStepName}

## Prerequisites

You need a GitHub access token with permission to create releases. You can provide a personal access token through Jenkins credentials or Vault, or retrieve a GitHub App installation token from System Trust.
When both System Trust and Vault are configured, System Trust is preferred and Vault is used only as a fallback. Set `skipSystemTrust: true` for the step to bypass System Trust and use the Vault token.

Please see [GitHub documentation for details about authenticating to the REST API](https://docs.github.com/en/rest/authentication/authenticating-to-the-rest-api).

## ${docJenkinsPluginDependencies}

## ${docGenParameters}

## ${docGenConfiguration}

## ${docGenDescription}

## Example

Usage of pipeline step:

```groovy
githubPublishRelease script: this, releaseBodyHeader: "**This is the latest success!**<br />"
```
