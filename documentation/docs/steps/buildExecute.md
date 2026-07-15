# ${docGenStepName}

!!! warning "Jenkins / Groovy step"
    This step is implemented as a Groovy DSL step and is available for **Jenkins pipelines only**.
    It is not available in GitHub Actions (GPP) pipelines.

## ${docGenDescription}

## Prerequisites

When performing a Docker build you need to maintain the respective credentials in your Jenkins credentials store.<br />
Further details

* for builds when a Docker deamon: see step [containerPushToRegistry](containerPushToRegistry.md)
* for builds using Kaniko: see step [kanikoExecute](kanikoExecute.md)

## Example

```groovy
buildExecute script:this, buildTool: 'maven'
```

## ${docGenParameters}

## ${docGenConfiguration}

## ${docJenkinsPluginDependencies}
