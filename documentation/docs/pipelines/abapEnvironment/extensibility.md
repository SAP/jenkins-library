# Extensibility

In general, the SAP BTP, ABAP environment pipeline supports different scenarios. The idea is that only configured stages are executed and the user is able to choose the appropriate stages.
In this section, you can learn how to extend the SAP BTP, ABAP environment pipeline with our recommended and best-practice approaches.

## 1. Extend the ATC stage via the Checkstyle/Warnings Next Generation Plugin

The `ATC` stage will execute ATC checks on a SAP BTP ABAP environment system via the step [abapEnvironmentRunATCCheck](https://sap.github.io/jenkins-library/steps/abapEnvironmentRunATCCheck/).
These results will be pinned to the respective Jenkins Jobs as an XML file in Checkstyle format. Per default this file will be named `ATCResults.xml`. You can change the file name via the step parameter `atcResultsFileName`.
Jenkins offers the possibility to  display the ATC results utilizing the checkstyle format with the [Warnings Next Generation Plugin](https://www.jenkins.io/doc/pipeline/steps/warnings-ng/#warnings-next-generation-plugin) ([GitHub Project](https://github.com/jenkinsci/warnings-ng-plugin)).

To achieve this, create a file `.pipeline/extensions/ATC.groovy` with the following content:

```groovy
void call(Map params) {
  //access stage name
  echo "Start - Extension for stage: ${params.stageName}"

  //access config
  echo "Current stage config: ${params.config}"

  //execute original stage as defined in the template
  params.originalStage()

  recordIssues tools: [checkStyle(pattern: '**/ATCResults.xml')], qualityGates: [[threshold: 1, type: 'TOTAL', unstable: true]]

  echo "End - Extension for stage: ${params.stageName}"
}
return this
```

The Jenkins pipeline step [recordIssues](https://www.jenkins.io/doc/pipeline/steps/warnings-ng/#recordissues-record-compiler-warnings-and-static-analysis-results) captures the results:
While `tools: [checkStyle(pattern: '**/**/ATCResults.xml')]` will display the ATC findings using the checkstyle format, `qualityGates: [[threshold: 1, type: 'TOTAL', unstable: true]]` will set the build result to UNSTABLE in case the ATC results contain at least one warning or error in total.

You can define several quality gates that will be checked after the issues have been reported. For example by providing a `qualityGates` configuration with option `unstable: false` it would be possible to end the pipeline execution in case of findings. See [Quality gate configuration](https://github.com/jenkinsci/warnings-ng-plugin/blob/master/doc/Documentation.md#quality-gate-configuration) for details.

If the pipeline execution should be aborted in case of ATC findings, to not continue with execution of following pipeline stages, use the [error](https://www.jenkins.io/doc/pipeline/steps/workflow-basic-steps/#error-error-signal) step in the stage extension to cause the build to stop:

```groovy
if (currentBuild.result == 'FAILURE') {
  error('Stopping build due to ATC Check quality gate')
}
```

!!! caution "Local Jenkins"
    If you are using a local Jenkins you may have to [adapt the Jenkins URL](https://stackoverflow.com/a/39543223) in the configuration if the CheckStyle Plugin shows this error: "Can't create fingerprints for some files".

## 2. Extend the ATC stage to send ATC results via E-Mail

In general when executing the `ATC` stage, the respective ATC results will normally be pinned to the Jenkins Job in a checkStyle XML format.
Additionally, you can set the `generateHTML` flag to `true` for the `abapEnvironmentRunATCCheck` step. This includes the generation of an HTML document containing the ATC results for the `abapEnvironmentRunATCCheck` step that will also be pinned to the respective Jenkins Job.
The ATC results can be attached to an E-Mail or being sent as the E-Mail body with the [Email Extension Plugin](https://www.jenkins.io/doc/pipeline/steps/email-ext/) ([GitHub Project](https://github.com/jenkinsci/email-ext-plugin)) using the `emailext()` method. Make sure that you have configured the Email Extension Plugin correctly before using it.

In the following example we only provide a sample configuration using the Jenkins [Email Extension Plugin](https://www.jenkins.io/doc/pipeline/steps/email-ext/). The E-Mail can be fully customized to your needs. Please refer to the Email Extension Plugin Documentation to see the full list of parameter that are supported.
If you haven't created it already, create/extend the file `.pipeline/extensions/ATC.groovy` with the following content:

```groovy
void call(Map params) {
  //access stage name
  echo "Start - Extension for stage: ${params.stageName}"

  //access config
  echo "Current stage config: ${params.config}"

  //execute original stage as defined in the template
  params.originalStage()

  emailext (
    attachmentsPattern: 'ATCResults.html', //This will attach the ATC results to the E-Mail
    to: 'user@example.com, admin@example.com',
    subject: "ATC results Mail from latest Run in System H01",
    body: 'Dear User, here are the results from the latest ATC run ${env.BUILD_ID}.' + readFile('ATCResults.html') //This will parse the ATC results and send it as the E-Mail body
 )

  echo "End - Extension for stage: ${params.stageName}"
}
return this
```

Note that in above example the ATC results, stored in the `ATCResults.html` file that is pinned to the Jenkins Job, will be sent as an attachmend using the `attachmentsPattern` parameter as well as being parsed and attached to the E-Mail body using the `body` parameter. Both methods are possible. If you chose to include the ATC results in the E-Mail body make sure to read the file content properly, e.g. using the `readFile()` method.
The `subject` parameter defines the subject of the E-Mail that will be sent. The `to` parameter specifies a list of recipients separated by a comma. You can also set a Distribution Lists as a recipient.
For all parameters it is also possible to use Jenkins environment variables like `${env.BUILD_ID}` or `${env.JENKINS_URL}`.

## 3. Extend the AUnit stage via the JUnit Plugin

The `AUnit` stage will execute AUnit test runs on a SAP BTP ABAP environment system via the step [abapEnvironmentRunAUnitTest](https://sap.github.io/jenkins-library/steps/abapEnvironmentRunAUnitTest/).
These results will be pinned to the respective Jenkins Jobs as an XML file in the JUnit format. Per default this file will be named `AUnitResults.xml`. You can change the file name via the step parameter `aUnitResultsFileName`.
Jenkins offers the possibility to  display the AUnit results utilizing the JUnit format with the [JUnit Plugin](https://plugins.jenkins.io/junit/) ([GitHub Project](https://github.com/jenkinsci/junit-plugin)).

To achieve this, create a file `.pipeline/extensions/AUnit.groovy` with the following content:

```groovy
void call(Map params) {
  //access stage name
  echo "Start - Extension for stage: ${params.stageName}"

  //access config
  echo "Current stage config: ${params.config}"

  //execute original stage as defined in the template
  params.originalStage()

  junit skipPublishingChecks: true, allowEmptyResults: true, testResults: '**/AUnitResults.xml'

  echo "End - Extension for stage: ${params.stageName}"
}
return this
```

You can simply use the JUnit Plugin for Jenkins in the AUnit stage within the `.pipeline/extensions/AUnit.groovy` file by using the `junit` command. You can set optional parameters like `skipPublishingChecks: true` in order to disable an integration to the GitHub Checks API. `allowEmptyResults: true` allows the build status of the Jenkins run to be `SUCCESS` even if there have been no results from the respective AUnit test run in the test results file. Vice versa, `allowEmptyResults: false` will set the build status to `FAILURE` if the test results file contains no results.
The `testResults` parameter specifies the path to the AUnit test results file which has been saved and pinned to the Jenkins job in the `abapEnvironmentRunAUnitTest` step. Please refer to the documentation of the ([JUnit Plugin](https://plugins.jenkins.io/junit/#documentation)) for more detailled information on the usage and configuration of the JUnit plugin parameters.

## 4. Extend the AUnit stage to send AUnit results via E-Mail

In general when executing the `AUnit` stage, the respective AUnit results will normally be pinned to the Jenkins Job in a JUnit XML format.
Additionally, you can set the `generateHTML` flag to `true` for the `abapEnvironmentRunAUnitTest` step. This includes the generation of an HTML document containing the AUnit results for the `abapEnvironmentRunAUnitTest` step that will also be pinned to the respective Jenkins Job.
The AUnit results can be attached to an E-Mail or being sent as the E-Mail body with the [Email Extension Plugin](https://www.jenkins.io/doc/pipeline/steps/email-ext/) ([GitHub Project](https://github.com/jenkinsci/email-ext-plugin)) using the `emailext()` method. Make sure that you have configured the Email Extension Plugin correctly before using it.

In the following example we only provide a sample configuration using the Jenkins [Email Extension Plugin](https://www.jenkins.io/doc/pipeline/steps/email-ext/). The E-Mail can be fully customized to your needs. Please refer to the Email Extension Plugin Documentation to see the full list of parameter that are supported.
If you haven't created it already, create/extend the file `.pipeline/extensions/AUnit.groovy` with the following content:

```groovy
void call(Map params) {
  //access stage name
  echo "Start - Extension for stage: ${params.stageName}"

  //access config
  echo "Current stage config: ${params.config}"

  //execute original stage as defined in the template
  params.originalStage()

  emailext (
    attachmentsPattern: 'AUnitResults.html', //This will attach the AUnit results to the E-Mail
    to: 'user@example.com, admin@example.com',
    subject: "AUnit results Mail from latest Run in System H01",
    body: 'Dear User, here are the results from the latest AUnit test run ${env.BUILD_ID}.' + readFile('AUnitResults.html') //This will parse the AUnit results and send it as the E-Mail body
 )

  echo "End - Extension for stage: ${params.stageName}"
}
return this
```

Note that in above example the AUnit test run results, stored in the `AUnitResults.html` file that is pinned to the Jenkins job, will be sent as an attachment using the `attachmentsPattern` parameter as well as being parsed and attached to the E-Mail body using the `body` parameter. Both methods are possible. If you chose to include the AUnit test run results in the E-Mail body make sure to read the file content properly, e.g. using the `readFile()` method.
The `subject` parameter defines the subject of the E-Mail that will be sent. The `to` parameter specifies a list of recipients separated by a comma. You can also set a distribution list as a recipient.
For all parameters it is also possible to use Jenkins environment variables like `${env.BUILD_ID}` or `${env.JENKINS_URL}`.
