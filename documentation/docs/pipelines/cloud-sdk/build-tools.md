# Build Tools

The SAP Cloud SDK supports multiple programming languages (Java and JavaScript) and can be used in the SAP Cloud Application Programming Model.
For each of these variants project templates exists (as referenced in the project's main [Readme](https://github.com/SAP/cloud-s4-sdk-pipeline/blob/master/README.md) file).
These templates introduce standard tooling, such as build tools, and a standard structure.

The SAP Cloud SDK Continuous Delivery Toolkit expects that the project follows this structure and depends on the build tools introduced by these templates.

The supported build tools are:

* [Maven](https://maven.apache.org/) for Java projects
* [npm](https://www.npmjs.com/) for JavaScript projects
* [MTA](https://sap.github.io/cloud-mta-build-tool) for Multi-Target Application Model projects

MTA itself makes use of other build tools, such as Maven and npm depending on what types of modules your application has.

*Note: The npm pipeline variant is in an early state. Some interfaces might change. We recommend consuming a fixed released version as described in the project [Readme](https://github.com/SAP/cloud-s4-sdk-pipeline/blob/master/README.md#versioning).*

## Feature Matrix

Support for the different features of the pipeline may vary in each variant of the SDK pipeline build tool.
The following table gives an overview over the features available per build tool.

| Feature                    | Maven | npm | MTA Maven | MTA npm |
|----------------------------|-------|-----|-----------|---------|
| Automatic Versioning       | x     | x   | x         | x       |
| Build                      | x     | x   | x         | x       |
| Backend Integration Tests  | x     | x   | x         | x       |
| Frontend Integration Tests | x     | x   | x         | x       |
| Backend Unit Tests         | x     | x   | x         | x       |
| Frontend Unit Tests        | x     | x   | x         | x       |
| NPM Dependency Audit       | x     | x   | x         | x       |
| Linting                    | x     |     | x         | x       |
| Static Code Checks         | x     |     | x         |         |
| End-To-End Tests           | x     |     | x         | x       |
| Performance Tests          | x     |     | x         |         |
| Resilience Checks          | x     |     | x         |         |
| S4HANA Public APIs         | x     |     | x         |         |
| Code Coverage Checks       | x     | x   | x         | x       |
| Checkmarx Integration      | x     |     | x         |         |
| Fortify Integration        | x     |     | x         |         |
| SourceClear Integration    | x     |     |           |         |
| Whitesource Integration    | x     | x   | x         | x       |
| Deployment to Nexus        | x     |     | x         | x       |
| Zero Downtime Deployment   | x     | x   | x¹        | x¹      |
| Download Cache             | x     | x   | x         | x       |

¹ MTA projects can only be deployed to the Cloud Foundry Environment

## Java/Node.js runtime versions

Runtime versions used in builds are determined by Docker images.

For Java, the default is still (as of August 2020) version 8.
For more details, please check the [documentation of the SAP Cloud SDK for Java](https://sap.github.io/cloud-sdk/docs/java/getting-started/).

In case you need to use a specific Java version to build your application, you may do so by setting another Docker image in your `.pipeline/config.yml` file.
See [documentation of the pipeline configuration](../configuration/) and look for the `dockerImage` key on where this option applies.

In most cases, it should be suffcient to configure an image for the `mavenExecute` step like so:

```yaml
steps:
  mavenExecute:
    dockerImage: 'maven:3.6.3-jdk-11'
```

## Projects Requirements

Each variant of the pipeline has different requirements regarding the project structure, location of reports and tooling.

Stages not listed here do not have a special requirement.
In any case, please also consult the [documentation of the pipeline configuration](../configuration/), as some stages have to be activated by providing configuration values.

### Build Tool Independent Requirements

In order to run in the pipeline your project has to include the following two files in the root folder: `Jenkinsfile` and `.pipeline/config.yml`.
You can copy both files from this [github repository](https://github.com/SAP/cloud-s4-sdk-pipeline/blob/master/archetype-resources).
There are two variants of the configuration file.
Please pick the corresponding version for your deployment target and rename it properly.

#### Frontend Unit Tests

For each `package.json` where the script `ci-frontend-unit-test` is defined the command `npm run ci-frontend-unit-test` will be executed in this stage.
Furthermore, the test results have to be stored in the folder `./s4hana_pipeline/reports/frontend-unit` in the root directory.
The required format of the test result report is the JUnit format as an `.xml` file.
The code coverage report can be published as html report and in the cobertura format.
The cobertura report as html report has to be stored in the directory `./s4hana_pipeline/reports/coverage-reports/frontend-unit/report-html/ut/` as an `index.html` file.
These coverage reports will then be published in Jenkins.
Furthermore, if configured in the `.pipeline/config.yml`, the pipeline ensures the configured level of code coverage.

#### Frontend Integration Tests

The command `npm run ci-it-frontend` will be executed in this stage and has to be defined in the `package.json` in the root.
In this stage, the frontend should be tested end-to-end without the backend.
Therefore, even a browser is started to simulate user interactions.
Furthermore, the test results have to be stored in the folder `./s4hana_pipeline/reports/frontend-integration` in the root directory of the project.
The required format of the test result report is the JUnit format as an `.xml` file.
The user is responsible to use a proper reporter for generating the results.
It is recommended to use the same tools as in the `package.json` of this [example project](https://github.com/SAP/cloud-s4-sdk-examples/blob/scaffolding-js/package.json).

#### Backend Unit Tests

##### Maven

Maven unit-tests are executed as part of the [`mavenBuild`](../../../steps/mavenBuild/) step.
They are supposed to be placed inside of `application/src/test`.

##### Java MTA modules

We run the command `mvn test` in each Java MTA module.

##### Npm and Nodejs MTA modules

For each `package.json` where the script `ci-backend-unit-test` is defined the command `npm run ci-backend-unit-test` will be executed in this stage.
Furthermore, the test results have to be stored in the folder `./s4hana_pipeline/reports/backend-unit/` in the root directory of the project.
The required format of the test result report is the JUnit format as an `.xml` file.
For the code coverage the results have to be stored in the folder `./s4hana_pipeline/reports/coverage-reports/backend-unit/` in the cobertura format as an `xml` file.
The user is responsible to use a proper reporter for generating the results.
We recommend the tools used in the `package.json` of this [example project](https://github.com/SAP/cloud-s4-sdk-examples/blob/scaffolding-js/package.json).
If you have multiple npm packages with unit tests the names of the report files must have unique names.

#### Backend Integration Tests

##### Maven and Java MTA modules

If there is a maven module called `integration-tests` we run `maven test` in this module.

##### Npm and Nodejs MTA modules

For each `package.json` where the script `ci-it-backend` is defined the command `npm run ci-it-backend` will be executed in this stage.
Furthermore, the test results have to be stored in the folder `./s4hana_pipeline/reports/backend-integration` in the root directory of the project.
The required format of the test result report is the JUnit format as an `.xml` file.
For the code coverage the results have to be stored in the folder `./s4hana_pipeline/reports/coverage-reports/backend-integration/` in the cobertura format as an `xml` file.
The user is responsible to use a proper reporter for generating the results.
We recommend the tools used in the `package.json` of this [example project](https://github.com/SAP/cloud-s4-sdk-examples/blob/scaffolding-js/package.json).
If you have multiple npm packages with unit tests the names of the report files must have unique names.

#### Lint

For each `package.json` where the script `ci-lint` is defined the command `npm run ci-lint` will be executed as part of the `build` stage.
The required format of the linting results is the checkstyle format as an `xml` file.
The linting results have to be stored in a file named `*cilint.xml`, which may reside in any directory of the project.
The linting results will then be published in Jenkins.

If no script `ci-lint` is defined and Javascript or Typescript files are present in the project, the pipeline will automatically execute ESLint.

If no ESLint configuration files are present in the project directory, a general purpose configuration is used to lint all Javascript and/or Typescript files of the project.
If, on the other hand, ESLint configuration files exist in the project, they will be used to lint Javascript files in the project.
The execution happens according to ESLint's default execution behavior, i.e., for each JS file the ESLint config in that directory or one of the parent directories will be used to lint the file.
Note, in this case only those files will be linted, for which an ESLint config exists.
More details on the execution behavior of ESLint and the usage of configuration files can be found in the [related documentation](https://eslint.org/docs/user-guide/configuring#configuration-cascading-and-hierarchy).
Note, if it is necessary to disable the default linting behavior, it is possible to, e.g., define a script `"ci-lint" : "exit 0"` in your `package.json`.

We recommend the use of a custom defined `ci-lint` script in your `package.json` to address project specific linting requirements.

#### End-to-End Tests

This stage is only executed if you configured it in the file `.pipeline/config.yml`.

The command `npm run ci-e2e` will be executed in this stage.
The url which is defined as `appUrl` in the file `.pipeline/config.yml` will be passed as argument named `launchUrl` to the tests.
This can be reproduced locally by executing:

```
npm run ci-e2e -- --launchUrl=https://path/to/your/running/application
```

The credentials also defined in the file `.pipeline/config.yml` will be available during the test execution as environment variables named `e2e_username` and `e2e_password`.

The test results have to be stored in the folder `./s4hana_pipeline/reports/e2e` in the root directory.
The required format of the test result report is the Cucumber format as an `.json` file, or the JUnit format as an xml file.
Also, screenshots can be stored in this folder.
The screenshots and reports will then be published in Jenkins.
The user is responsible to use a proper reporter for generating the results.

#### Performance Tests

This stage is only executed if you configured it in the file `.pipeline/config.yml`.

Performance tests can be executed using [JMeter](https://jmeter.apache.org/) or [Gatling](https://gatling.io/).

If only JMeter is used as a performance tests tool then test plans can be placed in a default location, which is the directory `{project_root}/performance-tests`. However, if JMeter is used along with Gatling, then JMeter test plans should be kept in a subdirectory under a directory `performance-tests` for example`./performance-tests/JMeter/`.

The gatling test project including the `pom.xml` should be placed in the directory `{project_root}/performance-tests`.
Afterwards, Gatling has to be enable in the configuration.

#### Deployments

For all deployments to Cloud Foundry (excluding MTA) there has to be a file called `manifest.yml`.
This file may only contain exactly one application.
*Note: For JavaScript projects the path of the application should point to the folder `deployment`.*

### Java / Maven

For Maven the pipeline expects the following structure.
The project should have three maven modules named:

- `application`
- `integration-tests`

The module `application` should contain the application code and unit tests.
The module `integration-tests` should contain integration tests.

Furthermore, the test modules have to include the following dependency:

```xml
<dependency>
    <groupId>com.sap.cloud.s4hana.quality</groupId>
    <artifactId>listeners-all</artifactId>
    <scope>test</scope>
</dependency>
```

### JavaScript / npm

The project has to use npm and include a `package.json` in the root directory.
In the pipeline stages, specific scripts in the `package.json` are called to build the project or run tests.
Furthermore, the pipeline expects reports, such as test results, to be written into certain folders.
These stage specific requirements are documented below.

#### Build

By default `npm ci` will be executed.
After `npm ci` the command  `npm  run ci-build` will be executed.
This script can be used to, for example, compile Typescript resources or webpack the frontend.
In the build stage, also development dependencies are installed and tests should also be compiled.

Afterwards the command `npm run ci-package` will be executed.
This step should prepare the deployment by copying all deployment relevant files into the folder `deployment` located in the root of the project.
This folder should not contain any non-production-related resources, such as tests or development dependencies.
This directory has to be defined as path in the `manifest.yml`.

*Note: This steps runs isolated from the steps before. Thus, e.g. modifying node_modules with `npm prune --production` will not have an effect for later stages, such as the test execution.*

### SAP Cloud Application Programming Model / MTA

The project structure follows the standard structure for projects created via the _SAP Cloud Platform Business Application_ SAP Web IDE Template with some constraints.
Please leave the basic structure of the generated project intact.

Make sure to check the _Include support for continuous delivery pipeline of SAP Cloud SDK_ checkbox, which will automatically add the required files for continuous delivery in your project.

If you already created your project without this option, you'll need to copy and paste two files into the root directory of your project, and commit them to your git repository:

* [`Jenkinsfile`](https://github.com/SAP/cloud-s4-sdk-pipeline/blob/master/archetype-resources/Jenkinsfile)
* [`.pipeline/config.yml`](https://github.com/SAP/cloud-s4-sdk-pipeline/blob/master/archetype-resources/cf-pipeline_config.yml)
    * Note: The file must be named `.pipeline/config.yml`, despite the different name of the file template

Further constrains on the project structure (this is all correct in projects generated from the _SAP Cloud Platform Business Application_ SAP Web IDE Template):

On the project root level, a `pom.xml` file is required.

Java services are Maven projects which include the application- and the unit-test code.
A service is typically called `srv`, but the name can be chosen freely.

An `integration-test` module must exist on the root level.
This module is where integration between the services can be tested.

In summary, the project structure should look like this:

```
.
├── Jenkinsfile
├── .pipeline
│   └── config.yml
├── app  // web application, not required
├── db   // only if database module exists
├── integration-tests
│   ├── pom.xml
│   └── src
│       └── test
├── mta.yaml
├── package.json
├── pom.xml
└── srv
    ├── pom.xml
    └── src
        ├── main
        └── test  // Unit-Tests for this service
```
