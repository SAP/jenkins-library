# ${docGenStepName}

## ${docGenDescription}

## Prerequisites

* **running Karma tests** - have a NPM module with running tests executed with Karma
* **configured WebDriver** - have the [`karma-webdriver-launcher`](https://github.com/karma-runner/karma-webdriver-launcher) package installed and a custom, WebDriver-based browser configured in Karma

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
