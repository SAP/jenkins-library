# ${docGenStepName}

!!! warning "Deprecation notice"
This step will soon be deprecated!

Please note that Karma is [marked as **DEPRECATED**](https://github.com/karma-runner/karma#karma-is-deprecated-and-is-not-accepting-new-features-or-general-bug-fixes) as of 04/2023. There is no migration path defined yet.

## ${docGenDescription}

## Prerequisites

* **running Karma tests** - have a NPM module with running tests executed with Karma
* **configured WebDriver** - have the [`karma-webdriver-launcher`](https://github.com/karma-runner/karma-webdriver-launcher) package installed and a custom, WebDriver-based browser configured in Karma

## ${docJenkinsPluginDependencies}

## ${docGenParameters}

## ${docGenConfiguration}

## Side effects

Step uses `seleniumExecuteTest` & `dockerExecute` inside.

## Exceptions

none

## Example

```groovy
karmaExecuteTests script: this, modules: ['./shoppinglist', './catalog']
```
