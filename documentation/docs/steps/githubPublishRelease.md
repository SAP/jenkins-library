# ${docGenStepName}

## Prerequisites

You need to create a personal access token within GitHub and add this to the Jenkins credentials store.

Please see [GitHub documentation for details about creating the personal access token](https://help.github.com/articles/creating-a-personal-access-token-for-the-command-line/).

## ${docJenkinsPluginDependencies}

## ${docGenParameters}

## ${docGenConfiguration}

## ${docGenDescription}

## Example

Usage of pipeline step:

```groovy
githubPublishRelease script: this, releaseBodyHeader: "**This is the latest success!**<br />"
```
