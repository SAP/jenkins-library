# ${docGenStepName}

## ${docGenDescription}

## Prerequisites

## ${docGenParameters}

## ${docGenConfiguration}

## Exceptions

If you see an error like `fatal: Not a git repository (or any parent up to mount point /home/jenkins)` it is likely that your test description cannot be found.<br />
Please make sure to point parameter `testOptions` to your `conf.js` file like `testOptions: './path/to/my/tests/conf.js'`

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
    baseUrl: '${params.url}',
    specs: specs,
    params: defaultParams, // can be overridden via cli `--params.<key>=<value>`
    auth: {
        // set up authorization for CF XSUAA
        'sapcloud-form': {
            user: '${params.user}',
            pass: '${params.pass}',
            userFieldSelector: 'input[name="username"]',
            passFieldSelector: 'input[name="password"]',
            logonButtonSelector: 'input[type="submit"]',
            redirectUrl: /cp.portal\/site/
        }
    }
};
```

While default values for `baseUrl`, `user` and `pass` are read from the environment, they can also be overridden when calling the CLI.

In a custom Pipeline, this is very simple: Just wrap the call to `uiVeri5ExecuteTests` in `withCredentials` (`TARGET_SERVER_URL` is read from `config.yml`):

```groovy
withCredentials([usernamePassword(
    credentialsId: 'MY_ACCEPTANCE_CREDENTIALS',
    passwordVariable: 'password',
    usernameVariable: 'username'
)]) {
    uiVeri5ExecuteTests script: this, testOptions: "./uiveri5/conf.js --params.user=${username} --params.pass=${password}"
}
```

In a Pipeline Template, a [Stage Exit](#) can be used to fetch the credentials and store them in the environment. As the environment is passed down to uiVeri5ExecuteTests, the variables will be present there. This is an example for the stage exit `.pipeline/extensions/Acceptance.groovy` where the `credentialsId` is read from the `config.yml`:

```groovy
void call(Map params) {
    // read username and password from the credential store
    withCredentials([usernamePassword(
        credentialsId: params.config.acceptanceCredentialsId,
        passwordVariable: 'password',
        usernameVariable: 'username'
    )]) {
        // store the result in the environment variables for executeUIVeri5Test
        withEnv(["TEST_USER=${username}", "TEST_PASS=${password}"]) {
            //execute original stage as defined in the template
            params.originalStage()
        }
    }
}
return this
```
