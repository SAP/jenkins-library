# ${docGenStepName}

!!! warning "Deprecation notice"
    Details of changes after the step migrated to a golang based step can be found [below](#exceptions).

## ${docGenDescription}

## Prerequisites

* prepared Postman with a test collection

## ${docGenParameters}

## ${docGenConfiguration}

## Side effects

Step uses `dockerExecute` inside.

## ${docJenkinsPluginDependencies}

## Exceptions

The step has been migrated into a golang-based step. The following release notes belong to the new implementation:

- Groovy Templating is deprecated and now replaced by Go Templating. The example show the required changes:

```yaml
# deprecated groovy default
newmanRunCommand: "run '${config.newmanCollection}' --environment '${config.newmanEnvironment}' --globals '${config.newmanGlobals}' --reporters junit,html --reporter-junit-export 'target/newman/TEST-${collectionDisplayName}.xml' --reporter-html-export 'target/newman/TEST-${collectionDisplayName}.html'"
```

```yaml
# current run command using golang templating
newmanRunCommand: "run \{\{.NewmanCollection\}\} --environment \{\{.Config.NewmanEnvironment\}\} --globals \{\{.Config.NewmanGlobals\}\} --reporters junit,html --reporter-junit-export target/newman/TEST-\{\{.CollectionDisplayName\}\}.xml --reporter-html-export target/newman/TEST-\{\{.CollectionDisplayName\}\}.html"
```

Including `--environment \{\{.Config.NewmanEnvironment\}\}` and `--globals \{\{.Config.NewmanGlobals\}\}` in the runCommand is rendundant since both parameters are also added to runCommand using `newmanEnvironment` and `newmanGlobals` from config.

## Example

Pipeline step:

```groovy
newmanExecute script: this
```

This step should be used in combination with `testsPublishResults`:

```groovy
newmanExecute script: this, failOnError: false
testsPublishResults script: this, junit: [pattern: '**/newman/TEST-*.xml']
```
