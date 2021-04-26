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

While `tools: [checkStyle(pattern: '**/**/ATCResults.xml')]` will display the ATC findings using the checkstyle format, `qualityGates: [[threshold: 1, type: 'TOTAL', unstable: true]]` will set the build result to UNSTABLE in case the ATC results contain at least one warning or error in total.

!!! caution "Local Jenkins"
    If you are using a local Jenkins you may have to [adapt the Jenkins URL](https://stackoverflow.com/a/39543223) in the configuration if the CheckStyle Plugin shows this error: "Can't create fingerprints for some files".

## 2. Extend the ATC stage to send ATC Results via E-Mail

In general when executing the `ATC` stage, the respective ATC results will normally be pinned to the Jenkins Job in a checkStyle XML format. Additionally, you can set the `generateHTML` flag to `true` for the `abapEnvironmentRunATCCheck` step. It includes the generation of an HTML document containing the ATC results for the `abapEnvironmentRunATCCheck` step that will also be pinned to the respective Jenkins Job.
The ATC Results can be attached to an E-Mail or being sent as the E-Mail body with the [Email Extension Plugin](https://www.jenkins.io/doc/pipeline/steps/email-ext/) ([GitHub Project](https://github.com/jenkinsci/email-ext-plugin)). Make sure that you have configured the Plugin correctly before using it.
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
 
  emailext(
    attachmentsPattern: 'ATCResults.html', //This will attach the ATC Results to the E-Mail
    to: 'user@example.com, admin@example.com',
    subject: "ATC Results Mail from latest Run in System H01",
    body: 'Dear User, here are the results from the latest ATC run ${env.BUILD_ID}.' + readFile('ATCResults.html') //This will parse the ATC Results and send it as the E-Mail body
 )

  echo "End - Extension for stage: ${params.stageName}"
}
return this
```

Note that in above example the ATC Results, stored in the `ATCResults.html` file that is pinned to the Jenkins Job, will be sent as an attachmend using the `attachmentsPattern` parameter as well as being parsed and attached to the E-Mail body using the `body` parameter. Both methods are possible. If you chose to include the ATC Results in the E-Mail body make sure to read the file content properly, e.g. using the `readFile()` method.
The `subject` parameter defines the subject of the E-Mail that will be sent. The `to` parameter specifies a list of recipients separated by a comma. You can also use Distribution Lists.
For all parameters it is also possible to use Jenkins environment variables, e.g. `${env.BUILD_ID}` or `${env.JENKINS_URL}`.
