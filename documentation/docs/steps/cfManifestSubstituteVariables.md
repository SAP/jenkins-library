# ${docGenStepName}

## ${docGenDescription}

## Prerequisites

This step is only relevant if both a `manifest.yml` and a corresponding variables Yaml file are found at the specified paths in the current source tree.
The step will activate itself in this case, and tries to replace any variable references found in `manifest.yml` with the values found in the variables file.

**Note:** It is possible to use one variables file for more than one `manifest.yml`.

## ${docGenParameters}

## ${docGenConfiguration}

## ${docJenkinsPluginDependencies}

## Side effects

Unless configured otherwise, this step will *replace* the input `manifest.yml` with a version that has all variable references replaced. This alters the source tree in your Jenkins workspace.
If you prefer to generate a separate output file, use the step's `outputManifestFile` parameter. Keep in mind, however, that your Cloud Foundry deployment step should then also reference this output file - otherwise CF deployment will fail with unresolved variable reference errors.

## Exceptions

* `org.yaml.snakeyaml.scanner.ScannerException` - in case any of the loaded input files contains malformed Yaml and cannot be parsed.

* `hudson.AbortException` - in case of internal errors and when not all variables could be replaced due to missing replacement values.

## Example

Usage of pipeline step:

```groovy
cfManifestSubstituteVariables manifestFile: "path/to/manifest.yml", variablesFile:"path/to/manifest-variables.yml", script: this
```

For example, you can refer to the parameters using relative paths:

```groovy
cfManifestSubstituteVariables manifestFile: "manifest.yml", variablesFile:"manifest-variables.yml", script: this
```

You can also refer to parameters using absolute paths, like this:

```groovy
cfManifestSubstituteVariables manifestFile: "\$\{WORKSPACE\}/manifest.yml", variablesFile:"\$\{WORKSPACE\}/manifest-variables.yml", script: this
```

If you are using the Cloud Foundry [Create-Service-Push](https://github.com/dawu415/CF-CLI-Create-Service-Push-Plugin) CLI plugin you will most likely also have a `services-manifest.yml` file.
Also in this file you can specify variable references, that can be resolved from the same variables file, e.g. like this:

```groovy
cfManifestSubstituteVariables manifestFile: "manifest.yml", variablesFile:"manifest-variables.yml", script: this // resolve variables in manifest.yml
cfManifestSubstituteVariables manifestFile: "services-manifest.yml", variablesFile:"manifest-variables.yml", script: this // resolve variables in services-manifest.yml from same file.
```
