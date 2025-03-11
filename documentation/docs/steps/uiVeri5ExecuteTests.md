# ${docGenStepName}

!!! warning "Deprecation notice"
    Details of changes after the step migrated to a golang based step can be found [below](#exceptions).

## ${docGenDescription}

## Prerequisites

## ${docGenParameters}

## ${docGenConfiguration}

## ${docJenkinsPluginDependencies}

## Exceptions

The parameter `testOptions` is deprecated and is replaced by array type parameter `runOptions`. Groovy templating for this parameter is deprecated and no longer supported.

Using the `runOptions` parameter the 'seleniumAddress' for UIVeri5 can be set.
The former groovy implementation included a default for seleniumAddress in the runCommand. Since this is not possible with the golang-based implementation, the seleniumAddress has to be added to the runOptions. For jenkins on kubernetes the host is 'localhost', in other environments, e.g. native jenkins installations, the host can be set to 'selenium'.

```yaml
runOptions: ["--seleniumAddress=http://localhost:4444/wd/hub", ..... ]
```

The parameter `failOnError` is no longer supported on the step due to strategic reasons of pipeline resilience. To achieve the former behaviour with `failOnError: false` configured, the step can be wrapped using try/catch in your custom pipeline script.

The `installCommand` does not support queueing shell commands using `&&` and `|` operator any longer.

If you see an error like `fatal: Not a git repository (or any parent up to mount point /home/jenkins)` it is likely that your test description cannot be found.<br />
Please make sure to point parameter `runOptions` to your `conf.js` file like `runOptions: [...., './path/to/my/tests/conf.js']`

## Examples

### Passing credentials from Jenkins

When running acceptance tests in a real environment, authentication will be enabled in most cases. UIVeri5 includes [features to automatically perform the login](https://github.com/SAP/ui5-uiveri5/blob/master/docs/config/authentication.md) with credentials in the `conf.js`. However, having credentials to the acceptance system stored in plain text is not an optimal solution.

Therefore, UIVeri5 allows templating to set parameters at runtime, as shown in the following example `conf.js`:

```js
// Read environment variables
const defaultParams = {
    url: process.env.TARGET_SERVER_URL,
    user: process.env.TEST_USER,
    pass: process.env.TEST_PASS
};

// Resolve path to specs relative to the working directory
const path = require('path');
const specs = path.relative(process.cwd(), path.join(__dirname, '*.spec.js'));

// export UIVeri5 config
exports.config = {
    profile: 'integration',
    baseUrl: '\${params.url}',
    specs: specs,
    params: defaultParams, // can be overridden via cli `--params.<key>=<value>`
    auth: {
        // set up authorization for CF XSUAA
        'sapcloud-form': {
            user: '\${params.user}',
            pass: '\${params.pass}',
            userFieldSelector: 'input[id="j_username"]',
            passFieldSelector: 'input[id="j_password"]',
            logonButtonSelector: 'button[type="submit"]',
            redirectUrl: /cp.portal\/site/
        }
    }
};
```

While default values for `baseUrl`, `user` and `pass` are read from the environment, they can also be overridden when calling the CLI.

In a custom Pipeline, this is very simple: Just wrap the call to `uiVeri5ExecuteTests` in `withCredentials`:

```groovy
withCredentials([usernamePassword(
    credentialsId: 'MY_ACCEPTANCE_CREDENTIALS',
    passwordVariable: 'password',
    usernameVariable: 'username'
)]) {
    uiVeri5ExecuteTests script: this, runOptions: ["--baseURL=NEW_BASE_URL", "--params.user=${username}", "--params.pass=${password}", "--seleniumAddress=http://localhost:4444/wd/hub", "./uiveri5/conf.js"]
}
```

**Please note:** It is not recommended to override any secrets with the runOptions, because they may be seen in the Jenkins pipeline run console output. During the `withCredentials` call, the credentials are written to the environment and can be accessed by the test code.

The following example shows the recommended way to handle the username and password for a uiVeri5ExecuteTests call that needs authentication. The `passwordVariable` and `usernameVariable` need to match the environment variables in the test code.

```groovy
withCredentials([usernamePassword(
    credentialsId: 'MY_ACCEPTANCE_CREDENTIALS',
    passwordVariable: 'TEST_PASS',
    usernameVariable: 'TEST_USER'
)]) {
    uiVeri5ExecuteTests script: this, runOptions: ["--seleniumAddress=http://localhost:4444/wd/hub", "./uiveri5/conf.js"]
}
```

There is also the option to use [vault for test credentials](https://www.project-piper.io/infrastructure/vault/#using-vault-for-test-credentials).

In a Pipeline Template, a [Stage Exit](../extensibility.md#1-extend-individual-stages) can be used to fetch the credentials and store them in the environment. As the environment is passed down to uiVeri5ExecuteTests, the variables will be present there. This is an example for the stage exit `.pipeline/extensions/Acceptance.groovy` where the `credentialsId` is read from the `config.yml`:

```groovy
void call(Map params) {
    // read username and password from the credential store
    withCredentials([usernamePassword(
        credentialsId: params.config.acceptanceCredentialsId,
        passwordVariable: 'password',
        usernameVariable: 'username'
    )]) {
        // store the result in the environment variables for executeUIVeri5Test
        withEnv(["TEST_USER=\${username}", "TEST_PASS=\${password}"]) {
            //execute original stage as defined in the template
            params.originalStage()
        }
    }
}
return this
```
