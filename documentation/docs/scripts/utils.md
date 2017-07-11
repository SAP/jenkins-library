# Utils

## Description
Provides utility functions.

## Constructors

### Utils()

Default no-argument constructor. Instances of the Utils class does not hold any instance specific state.

#### Example

```groovy
new Utils()
```

## Method Details

### getMandatoryParameter(Map map, paramName, defaultValue)

#### Description

Retrieves the parameter value for parameter `paramName` from parameter map `map`. In case there is no parameter with the given key contained in parameter map `map` `defaultValue` is returned. In case there no such parameter contained in `map` and `defaultValue` is `null` an exception is thrown.

#### Parameters

* `map` - A map containing configuration parameters.
* `paramName` - The key of the parameter which should be looked up.
* `defaultValue` - The value which is returned in case there is no parameter with key `paramName` contained in `map`.

#### Return value
The value to the parameter to be retrieved, or the default value if the former is `null`, either since there is no such key or the key is associated with value `null`. In case the parameter is not defined or the value for that parameter is `null`and there is no default value an exception is thrown.

#### Side effects

none

#### Exceptions
* `Exception`: If the value to be retrieved and the default value are both `null`.

#### Example

```groovy
def utils =  new Utils()
def parameters = [DEPLOY_ACCOUNT: 'deploy-account']
assert utils.getMandatoryParameter(parameters, 'DEPLOY_ACCOUNT', null) == 'deploy-account'
assert utils.getMandatoryParameter(parameters, 'DEPLOY_USER', 'john_doe') == 'john_doe'
```

### retrieveGitCoordinates(script)

#### Description
Retrieves the git-remote-url and git-branch. The parameters 'GIT_URL' and 'GIT_BRANCH' are retrieved from Jenkins job configuration. If these are not set, the git-url and git-branch are retrieved from the same repository where the Jenkinsfile resides.


#### Parameters

* `script` The script calling the method. Basically the `Jenkinsfile`. It is assumed that the script provides access to the parameters defined when launching the build, especially `GIT_URL`and `GIT_BRANCH`.

#### Return value

A map containing git-url and git-branch: `[url: gitUrl, branch: gitBranch]`

## Exceptions

* `AbortException`: if only one of `GIT_URL`,  `GIT_BRANCH` is set in the Jenkins job configuration.

#### Example

```groovy
def gitCoordinates = new Utils().retrieveGitCoordinates(this)
def gitUrl = gitCoordinates.url
def gitBranch = gitCoordinates.branch
```
