# ${docGenStepName}

## ${docGenDescription}

## Prerequisites

none

## Example

```groovy
seleniumExecuteTests (script: this) {
    git url: 'https://github.com/xxxxx/WebDriverIOTest.git'
    sh '''npm install
        node index.js'''
}
```

### Example test using WebdriverIO

Example based on <http://webdriver.io/guide/getstarted/modes.html> and <http://webdriver.io/guide.html>

#### Configuration for Local Docker Environment

```js
var webdriverio = require('webdriverio');
var options = {
    host: 'selenium',
    port: 4444,
    desiredCapabilities: {
        browserName: 'chrome'
    }
};
```

#### Configuration for Kubernetes Environment

```js
var webdriverio = require('webdriverio');
var options = {
    host: 'localhost',
    port: 4444,
    desiredCapabilities: {
        browserName: 'chrome'
    }
};
```

#### Test Code (index.js)

```js
// ToDo: add configuration from above

webdriverio
    .remote(options)
    .init()
    .url('http://www.google.com')
    .getTitle().then(function(title) {
        console.log('Title was: ' + title);
    })
    .end()
    .catch(function(err) {
        console.log(err);
    });
```

## ${docGenParameters}

## ${docGenConfiguration}

## ${docJenkinsPluginDependencies}

## Side effects

none

## Exceptions

none
