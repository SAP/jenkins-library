# Utils

## Description
Provides git related utility functions.

## Constructors

### Utils()

Default no-argument constructor. Instances of the Utils class does not hold any instance specific state.

#### Example

```groovy
new GitUtils()
```

## Method Details

### getCredentialsId(url)

#### Description

Resolves a credentialsId based on the repositories configured in the Jenkins job definition.
In case there are several entries with the same repository url the credentialId of the first match is returned.

#### Parameters

* `repoUrl` - The url of the repository for that the credentialId should be resolved.

#### Return value
The credentialsId or `null` in case there is no repository with the given url configured or there are no credentials
maintained for that repository.

#### Side effects

none

#### Exceptions

none
#### Example

```groovy
def gitUtils = new GitUtils()
assert gitUtils.getCredentialsId('https://example.com/myGitRepo.git') == 'credentials-key'
```
