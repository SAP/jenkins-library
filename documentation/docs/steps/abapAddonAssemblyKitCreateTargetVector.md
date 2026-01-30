# ${docGenStepName}

!!! warning "Deprecation notice"
This step will soon be deprecated!

## ${docGenDescription}

## Prerequisites

* The credentials to access the AAKaaS (Technical Communication User) must be stored in the Jenkins Credential Store
* Product Version name and the resolved version(version, spslevel and patchlevel) must be part of the addonDescriptor structure in Piper commonPipelineEnvironment. This is the case if the step [abapAddonAssemblyKitCheckPV](https://sap.github.io/jenkins-library/steps/abapAddonAssemblyKitCheckPV) has been executed before.
* For each Software Component Version which should be part of the Target Vector, the name and the resolved version(version, splevel and patchlevel) as well as the Delivery Package must be part of the addonDescriptor structure in Piper commonPipelineEnvironment. This is the case if the step [abapAddonAssemblyKitCheckCVs](https://sap.github.io/jenkins-library/steps/abapAddonAssemblyKitCheckCVs) has been executed before.
* The Delivery Packages must exist in the package registry (status "P" = planned) which is the case if step [abapAddonAssemblyKitReserveNextPackages](https://sap.github.io/jenkins-library/steps/abapAddonAssemblyKitReserveNextPackages) has been executed before. Alternatively the package can already exist as physical packages (status "L" = locked or "R" = released).

A detailed description of all prerequisites of the scenario and how to configure them can be found in the [Scenario Description](https://www.project-piper.io/scenarios/abapEnvironmentAddons/).

## ${docGenParameters}

## ${docGenConfiguration}

## ${docJenkinsPluginDependencies}

## Examples

### Configuration in the config.yml

The recommended way to configure your pipeline is via the config.yml file. In this case, calling the step in the Jenkinsfile is reduced to one line:

```groovy
abapAddonAssemblyKitCreateTargetVector script: this
```

If the step is to be configured individually the config.yml should look like this:

```yaml
steps:
  abapAddonAssemblyKitCreateTargetVector:
    abapAddonAssemblyKitCredentialsId: 'abapAddonAssemblyKitCredentialsId'
```

More convenient ways of configuration (e.g. on stage level) are described in the respective scenario/pipeline documentation.
