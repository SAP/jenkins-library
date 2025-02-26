# ${docGenStepName}

!!! note
    Please note, that the npmExecuteTests step is in beta state, and there could be breaking changes before we remove the beta notice.

## ${docGenDescription}

## ${docGenParameters}

## ${docGenConfiguration}

## Examples

### Simple example using wdi5

```yaml
stages:
  - name: Test
    steps:
      - name: npmExecuteTests
        type: npmExecuteTests
        params:
          baseUrl: "http://example.com/index.html"
```

This will run your wdi5 tests with the given baseUrl.

### Advanced example using custom test script with credentials using Vault

```yaml
stages:
  - name: Test
    steps:
      - name: npmExecuteTests
        type: npmExecuteTests
        params:
          installCommand: "npm install"
          runCommand: "npm run custom-e2e-test"
          usernameEnvVar: "e2e_username"
          passwordEnvVar: "e2e_password"
          baseUrl: "http://example.com/index.html"
          urlOptionPrefix: "--base-url="
```

and Vault configuration in PIPELINE-GROUP-<id>/PIPELINE-<id>/appMetadata

```json
{
  "vaultURLs": [
    {
      "url": "http://one.example.com/index.html",
      "username": "some-username1",
      "password": "some-password1"
    },
    {
      "url": "http://two.example.com/index.html",
      "username": "some-username2",
      "password": "some-password2"
    }
  ],
  "vaultUsername": "base-url-username",
  "vaultPassword": "base-url-password"
}
```

This will run your custom install and run script for each URL from secrets and use the given URL like so:

```shell
npm run custom-e2e-test --base-url=http://one.example.com/index.html
```

Each test run will have their own environment variables set:

```shell
e2e_username=some-username1
e2e_password=some-password1
```

Environment variables are reset before each test run with their corresponding values from the secrets

### Custom environment variables and $PATH

```yaml
stages:
  - name: Test
    steps:
      - name: npmExecuteTests
        type: npmExecuteTests
        params:
          envs:
            - "MY_ENV_VAR=value"
          paths:
            - "/path/to/add"
```

If you're running uiVeri5 tests, you might need to set additional environment variables or add paths to the $PATH variable. This can be done using the `envs` and `paths` parameters:

```yaml
stages:
  - name: Test
    steps:
      - name: npmExecuteTests
        type: npmExecuteTests
        params:
          runCommand: "/home/node/.npm-global/bin/uiveri5"
          installCommand: "npm install @ui5/uiveri5 --global --quiet"
          runOptions: ["--seleniumAddress=http://localhost:4444/wd/hub"]
          usernameEnvVar: "PIPER_SELENIUM_HUB_USER"
          passwordEnvVar: "PIPER_SELENIUM_HUB_PASSWORD"
          envs:
            - "NPM_CONFIG_PREFIX=~/.npm-global"
          paths:
            - "~/.npm-global/bin"
```
