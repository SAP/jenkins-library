# ${docGenStepName}

!!! warning "Deprecation notice"
    Details of changes after the step migration to a golang can be found [below](#exceptions).

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

- **newmanRunCommand**:

The parameter `newmanRunCommand` is deprecated by now and is replaced by list parameter `runOptions`. For backward compatibility, the `newmanRunCommand` parameter will still be used if configured. Nevertheless, using this parameter can break the step in some cases, e.g. when spaces are used in single quoted strings like spaces in file names. Also Groovy Templating is deprecated and now replaced by Go Templating. The example show the required changes:

```yaml
# deprecated groovy default
newmanRunCommand: "run '${config.newmanCollection}' --environment '${config.newmanEnvironment}' --globals '${config.newmanGlobals}' --reporters junit,html --reporter-junit-export 'target/newman/TEST-${collectionDisplayName}.xml' --reporter-html-export 'target/newman/TEST-${collectionDisplayName}.html'"
```

```yaml
# new run options using golang templating
{{`runOptions: ["run", "{{.NewmanCollection}}", "--environment", "{{.Config.NewmanEnvironment}}", "--globals", "{{.Config.NewmanGlobals}}", "--reporters", "junit,html", "--reporter-junit-export", "target/newman/TEST-{{.CollectionDisplayName}}.xml", "--reporter-html-export", "target/newman/TEST-{{.CollectionDisplayName}}.html"]`}}
```

If the following error occurs during the pipeline run, the `newmanRunCommand` is probably still configured with the deprecated groovy template syntax:
> info  newmanExecute - error: collection could not be loaded
> info  newmanExecute -   unable to read data from file "${config.newmanCollection}"
> info  newmanExecute -   ENOENT: no such file or directory, open '${config.newmanCollection}'

- **newmanEnvironment and newmanGlobals**:

Referencing `newmanEnvironment` and `newmanGlobals` in the runOptions is redundant now. Both parameters are added to runCommand using `newmanEnvironment` and `newmanGlobals` from config  when configured and not referenced by go templating using `"--environment", "{{`{{.Config.NewmanEnvironment}}`}}"` and `"--globals", "{{`{{.Config.NewmanGlobals}}`}}"` as shown above.

## Passing Credentials

If you need to pass additional credentials you can do so via environment
variables. This is done via templating in the `runOptions`, as per this example:

```yaml
runOptions: [
    {{`"run", "{{.NewmanCollection}}",`}}
    {{`"--environment", "{{.Config.NewmanEnvironment}}",`}}
    {{`"--env-var", "username={{getenv \"PIPER_TESTCREDENTIAL_USERNAME\"}}",`}}
    {{`"--env-var", "password={{getenv \"PIPER_TESTCREDENTIAL_PASSWORD\"}}"`}}
]
```

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
