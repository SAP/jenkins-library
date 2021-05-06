# ${docGenStepName}

## ${docGenDescription}

## Prerequisites

Your company has registered an account with WhiteSource and you have enabled the use of so called `User Keys` to manage
access to your organization in WhiteSource via dedicated privileges. Scanning your products without adequate user level
access protection imposed on the WhiteSource backend would simply allow access based on the organization token.

## ${docGenParameters}

## ${docGenConfiguration}

## ${docJenkinsPluginDependencies}

## Exceptions

None

## Examples

```groovy
whitesourceExecuteScan script: this, buildTool: 'pip', productName: 'My Whitesource Product', userTokenCredentialsId: 'companyAdminToken', orgAdminUserTokenCredentialsId: 'orgAdminToken', orgToken: 'myWhitesourceOrganizationToken'
```
