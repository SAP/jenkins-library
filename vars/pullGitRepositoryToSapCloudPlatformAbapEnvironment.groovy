import static com.sap.piper.Prerequisites.checkScript
import com.sap.piper.ConfigurationHelper
import com.sap.piper.GenerateDocumentation
import com.sap.piper.JenkinsUtils
import com.sap.piper.Utils
import groovy.json.JsonSlurper
import hudson.AbortException
import groovy.transform.Field
import org.jenkinsci.plugins.workflow.steps.FlowInterruptedException

@Field def STEP_NAME = getClass().getName()
@Field Set GENERAL_CONFIG_KEYS = [
    /**
     * Specifies the host address of the SAP Cloud Platform ABAP Environment system
     */
    'host',
    /**
     * Specifies the name of the Repository (Software Component) on the SAP Cloud Platform ABAP Environment system
     */
    'repositoryName'
]
@Field Set STEP_CONFIG_KEYS = GENERAL_CONFIG_KEYS.plus([
    /**
     * Specifies the communication user of the communication scenario SAP_COM_0510
     */
    'username',
    /**
     * Specifies the password of the communication user
     */
    'password'
])
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS
/**
 * Pulls a Repository (Software Component) to a SAP Cloud Platform ABAP Environment system.
 *
 * !!! note "Git Repository and Software Component"
 *       In SAP Cloud Platform ABAP Environment Git repositories are wrapped in Software Components (which are managed in the App "Manage Software Components")
 *       Currently, those two names are used synonymous.
 */
@GenerateDocumentation
void call(Map parameters = [:]) {

    handlePipelineStepErrors(stepName: STEP_NAME, stepParameters: parameters, failOnError: true) {

        def script = checkScript(this, parameters) ?: this

        Map configuration = ConfigurationHelper.newInstance(this)
            .mixinGeneralConfig(script.commonPipelineEnvironment, GENERAL_CONFIG_KEYS)
            .mixinStepConfig(script.commonPipelineEnvironment, STEP_CONFIG_KEYS)
            .mixin(parameters, PARAMETER_KEYS)
            .withMandatoryProperty('host', 'Host not provided')
            .withMandatoryProperty('repositoryName', 'Repository / Software Component not provided')
            .withMandatoryProperty('username')
            .withMandatoryProperty('password')
            .collectValidationFailures()
            .use()

        String usernameColonPassword = configuration.username + ":" + configuration.password
        String authToken = usernameColonPassword.bytes.encodeBase64().toString()
        String urlString = configuration.host + '/sap/opu/odata/sap/MANAGE_GIT_REPOSITORY/Pull'
        echo "[${STEP_NAME}] General Parameters: URL = \"${urlString}\", repositoryName = \"${configuration.repositoryName}\""


        try {
            String urlPullEntity = triggerPull(configuration, urlString, authToken)
            if (urlPullEntity != null) {
                String finalStatus = pollPullStatus(urlPullEntity, authToken)
                if (finalStatus != 'S') {
                    error "[${STEP_NAME}] Pull Failed"
                }
            }
            workspaceCleanup()
        } catch (err) {
            workspaceCleanup()
            throw err
        }
    }
}

private String triggerPull(Map configuration, String url, String authToken) {

    String entityUri = null

    def xCsrfTokenScript = """#!/bin/bash
        curl -I -X GET ${url} \
        -H 'Authorization: Basic ${authToken}' \
        -H 'Accept: application/json' \
        -H 'x-csrf-token: fetch' \
        -D ${HeaderFiles.authFile} \
    """

    sh ( script : xCsrfTokenScript, returnStdout: true )

    HttpHeaderProperties headerProperties = new HttpHeaderProperties(readFile(HeaderFiles.authFile))
    checkRequestStatus(headerProperties)

    def scriptPull = """#!/bin/bash
        curl -X POST \"${url}\" \
        -H 'Authorization: Basic ${authToken}' \
        -H 'Accept: application/json' \
        -H 'Content-Type: application/json' \
        -H 'x-csrf-token: ${headerProperties.xCsrfToken}' \
        --cookie ${HeaderFiles.authFile} \
        -D ${HeaderFiles.postFile} \
        -d '{ \"sc_name\": \"${configuration.repositoryName}\" }'
    """
    def response = sh (
        script : scriptPull,
        returnStdout: true )

    checkRequestStatus(new HttpHeaderProperties(readFile(HeaderFiles.postFile)))

    JsonSlurper slurper = new JsonSlurper()
    Map responseJson = slurper.parseText(response)
    if (responseJson.d != null) {
        entityUri = responseJson.d.__metadata.uri.toString()
        echo "[${STEP_NAME}] Pull Status: ${responseJson.d.status_descr.toString()}"
    } else {
        error "[${STEP_NAME}] ${responseJson?.error?.message?.value?.toString()?:"No message available}"
    }

    echo "[${STEP_NAME}] Entity URI: ${entityUri}"
    return entityUri

}

private String pollPullStatus(String url, String authToken) {

    String headerFile = "headerPoll.txt"
    String status = "R";
    while(status == "R") {

        Thread.sleep(5000)

        def pollScript = """#!/bin/bash
            curl -X GET "${url}" \
            -H 'Authorization: Basic ${authToken}' \
            -H 'Accept: application/json' \
            -D ${HeaderFiles.pollFile}
        """
        def pollResponse = sh (
            script : pollScript,
            returnStdout: true )

        checkRequestStatus(new HttpHeaderProperties(readFile(HeaderFiles.pollFile)))

        JsonSlurper slurper = new JsonSlurper()
        Map pollResponseJson = slurper.parseText(pollResponse)
        if (pollResponseJson.d != null) {
            status = pollResponseJson.d.status.toString()
        } else {
            error "[${STEP_NAME}] ${pollResponseJson?.error?.message?.value?.toString()?:"No message available"}"
        }
        echo "[${STEP_NAME}] Pull Status: ${pollResponseJson.d.status_descr.toString()}"
    }
    return status
}

private void checkRequestStatus(HttpHeaderProperties httpHeader) {
    if (httpHeader.statusCode > 201) {
        error "[${STEP_NAME}] Connection Failed: ${httpHeader.statusCode} ${httpHeader.statusMessage}"
    }
}

private void workspaceCleanup() {
    String cleanupScript = """#!/bin/bash
            rm -f ${HeaderFiles.authFile}
            rm -f ${HeaderFiles.postFile}
            rm -f ${HeaderFiles.pollFile}
        """
    sh ( script : cleanupScript, returnStdout : true )
}

public class HttpHeaderProperties{
    Integer statusCode
    String statusMessage
    String xCsrfToken

    HttpHeaderProperties(String header) {
        def statusCodeRegex = header =~ /(?<=HTTP\/1.[0-9]\s)[0-9]{3}(?=\s)/
        if (statusCodeRegex.find()) {
            statusCode = statusCodeRegex[0].toInteger()
        }
        def statusMessageRegex = header =~ /(?<=HTTP\/1.[0-9]\s[0-9]{3}\s).*/
        if (statusMessageRegex.find()) {
            statusMessage = statusMessageRegex[0]
        }
        def xCsrfTokenRegex = header =~ /(?<=x-csrf-token:\s).*/
        if (xCsrfTokenRegex.find()) {
            xCsrfToken = xCsrfTokenRegex[0]
        }
    }
}

public class HeaderFiles{
    static final String authFile = "headerFileAuth.txt"
    static final String postFile = "headerFilePost.txt"
    static final String pollFile = "headerFilePoll.txt"
}
