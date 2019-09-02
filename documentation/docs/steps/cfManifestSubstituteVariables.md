# ${docGenStepName}

## ${docGenDescription}

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
cfManifestSubstituteVariables (
  script: this,
  manifestFile: "path/to/manifest.yml",                      //optional, default: manifest.yml
  manifestVariablesFiles: ["path/to/manifest-variables.yml"] //optional, default: ['manifest-variables.yml']
  manifestVariables: [[key : value], [key : value]]          //optional, default: []
)
```

For example, you can refer to the parameters using relative paths (similar to `cf push --vars-file`):

```groovy
cfManifestSubstituteVariables (
  script: this,
  manifestFile: "manifest.yml",
  manifestVariablesFiles: ["manifest-variables.yml"]
)
```

Furthermore, you can also specify variables and their values directly (similar to `cf push --var`):

```groovy
cfManifestSubstituteVariables (
  script: this,
  manifestFile: "manifest.yml",
  manifestVariablesFiles: ["manifest-variables.yml"],
  manifestVariables: [[key1 : value1], [key2 : value2]]
)
```

If you are using the Cloud Foundry [Create-Service-Push](https://github.com/dawu415/CF-CLI-Create-Service-Push-Plugin) CLI plugin you will most likely also have a `services-manifest.yml` file.
Also in this file you can specify variable references, that can be resolved from the same variables file, e.g. like this:

```groovy
// resolve variables in manifest.yml
cfManifestSubstituteVariables (
  script: this,
  manifestFile: "manifest.yml",
  manifestVariablesFiles: ["manifest-variables.yml"]
)

// resolve variables in services-manifest.yml from same file.
cfManifestSubstituteVariables (
  script: this,
  manifestFile: "services-manifest.yml",
  manifestVariablesFiles: ["manifest-variables.yml"]
)
```
