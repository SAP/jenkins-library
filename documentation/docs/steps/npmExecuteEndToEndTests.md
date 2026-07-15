# ${docGenStepName}

## ${docGenDescription}

## ${docGenParameters}

## ${docGenConfiguration}

## Examples

### Passing credentials

When running acceptance tests in a real environment, authentication will be enabled in most cases. WDI5 includes [features to automatically perform the login](https://ui5-community.github.io/wdi5/#/authentication). For this, if the step parameter `wdi5` is set to `true`, the provided basic auth credential (`credentialsId`) are mapped to the environment variables `wdi5_username` and `wdi5_password`.
