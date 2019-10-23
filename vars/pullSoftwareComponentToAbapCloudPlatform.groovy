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
     * Specifies the host address
     */
    'host',
    /**
     * Specifies the name of the Software Component
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
 * Pulls a Software Component to a SAP Cloud Platform ABAP Environment System.
 *
 * Prerequisite: the Communication Arrangement for the Communication Scenario SAP_COM_0510 has to be set up, including a Communication System and Communication Arrangement
 */
@GenerateDocumentation
void call(Map parameters = [:]) {

    handlePipelineStepErrors(stepName: STEP_NAME, stepParameters: parameters, failOnError: true) {

        def script = checkScript(this, parameters) ?: this

        ConfigurationHelper configHelper = ConfigurationHelper.newInstance(this)
            .mixinGeneralConfig(script.commonPipelineEnvironment, GENERAL_CONFIG_KEYS)
            .mixinStepConfig(script.commonPipelineEnvironment, STEP_CONFIG_KEYS)
            .mixin(parameters, PARAMETER_KEYS)

        Map configuration = configHelper.use()

        configHelper
            .withMandatoryProperty('host', 'Host not provided')
            .withMandatoryProperty('repositoryName', 'Repository / Software Component not provided')
            .withMandatoryProperty('username')
            .withMandatoryProperty('password')

        String usernameColonPassword = configuration.username + ":" + configuration.password
        String authToken = usernameColonPassword.bytes.encodeBase64().toString()
        String urlString = configuration.host + '/sap/opu/odata/sap/MANAGE_GIT_REPOSITORY/Pull'
        echo "[${STEP_NAME}] General Parameters: URL = \"${urlString}\", repositoryName = \"${configuration.repositoryName}\""

        String urlPullEntity = triggerPull(configuration, urlString, authToken)

        if (urlPullEntity != null) {
            String finalStatus = pollPullStatus(urlPullEntity, authToken)
            if (finalStatus != 'S') {
                error "[${STEP_NAME}] Pull Failed"
            }
        }
    }
}

private String triggerPull(Map configuration, String url, String authToken) {

    String entityUri = null
    String headerFile = "header.txt"

    def xCsrfTokenScript = """#!/bin/bash
        curl -I -X GET ${url} \
        -H 'Authorization: Basic ${authToken}' \
        -H 'Accept: application/json' \
        -H 'x-csrf-token: fetch' \
        -D ${headerFile} \
        | awk 'BEGIN {FS=": "}/^x-csrf-token/{print \$2}'
    """

    def xCsrfToken = sh (
        script : xCsrfTokenScript,
        returnStdout: true )

    checkRequestStatus(headerFile)

    def scriptPull = """#!/bin/bash
        curl -X POST \"${url}\" \
        -H 'Authorization: Basic ${authToken}' \
        -H 'Accept: application/json' \
        -H 'Content-Type: application/json' \
        -H 'x-csrf-token: ${xCsrfToken.trim()}' \
        -D ${headerFile}
        --cookie ${headerFile} \
        -d '{ \"sc_name\": \"${configuration.repositoryName}\" }'
    """
    def response = sh (
        script : scriptPull,
        returnStdout: true )

    checkRequestStatus(headerFile)

    JsonSlurper slurper = new JsonSlurper()
    Map responseJson = slurper.parseText(response)
    if (responseJson.d != null) {
        entityUri = responseJson.d.__metadata.uri.toString()
        echo "[${STEP_NAME}] Pull Status: ${responseJson.d.status_descr.toString()}"
    } else {
        error "[${STEP_NAME}] ${responseJson.error.message.value.toString()}"
    }

    echo "[${STEP_NAME}] Entity URI: ${entityUri}"
    return entityUri

}

private String pollPullStatus(String url, String authToken) {

    String headerFile = "header.txt"
    String status = "R";
    while(status == "R") {

        Thread.sleep(5000)

        def pollScript = """#!/bin/bash
            curl -X GET "${url}" \
            -H 'Authorization: Basic ${authToken}' \
            -H 'Accept: application/json' \
            -D ${headerFile}
        """
        def pollResponse = sh (
            script : pollScript,
            returnStdout: true )

        checkRequestStatus(headerFile)

        JsonSlurper slurper = new JsonSlurper()
        Map pollResponseJson = slurper.parseText(pollResponse)
        if (pollResponseJson.d != null) {
            status = pollResponseJson.d.status.toString()
        } else {
            error "[${STEP_NAME}] ${pollResponseJson.error.message.value.toString()}"
        }
        echo "[${STEP_NAME}] Pull Status: ${pollResponseJson.d.status_descr.toString()}"
    }
    return status
}

private void checkRequestStatus(String headerFile) {
    String headerString = readFile(headerFile)
    HttpHeader httpHeader = new HttpHeader(headerString)

    if (httpHeader.statusCode > 201) {
        echo readFile(headerFile)
        error "[${STEP_NAME}] Connection Failed: ${httpHeader.statusCode} ${httpHeader.statusMessage}"
    }
}

public class HttpHeader{
    Integer statusCode
    String statusMessage

    HttpHeader(String header) {
        def statusCodeRegex = header =~ /(?<=HTTP\/1.1\s)[0-9]{3}(?=\s)/
        if (statusCodeRegex.find()) {
            statusCode = statusCodeRegex[0].toInteger()
        }
        def statusMessageRegex = header =~ /(?<=HTTP\/1.1\s[0-9]{3}\s).*/
        if (statusMessageRegex.find()) {
            statusMessage = statusMessageRegex[0]
        }
    }

}