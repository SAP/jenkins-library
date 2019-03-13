# karmaExecuteTests

## Description

Content here is generated from corresponnding step, see `vars`.

## Prerequisites

* **running Karma tests** - have a NPM module with running tests executed with Karma
* **configured WebDriver** - have the [`karma-webdriver-launcher`](https://github.com/karma-runner/karma-webdriver-launcher) package installed and a custom, WebDriver-based browser configured in Karma

## Parameters

Content here is generated from corresponnding step, see `vars`.

## Step configuration

Content here is generated from corresponnding step, see `vars`.

## Side effects

Step uses `seleniumExecuteTest` & `dockerExecute` inside.

## Exceptions

none

## Example

```groovy
karmaExecuteTests script: this, modules: ['./shoppinglist', './catalog']
```
