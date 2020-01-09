import static com.sap.piper.Prerequisites.checkScript
import com.sap.piper.ConfigurationHelper
import com.sap.piper.GenerateDocumentation
import com.sap.piper.JenkinsUtils
import com.sap.piper.Utils
import com.sap.piper.BashUtils
import groovy.json.JsonSlurper
import hudson.AbortException
import groovy.transform.Field
import java.util.UUID
import java.util.regex.*

@Field def STEP_NAME = getClass().getName()
@Field Set STEP_CONFIG_KEYS = [
    /**
     * Specifies the host address of the SAP Cloud Platform ABAP Environment system
     */
    'host',
    /**
     * Jenkins CredentialsId containing the communication user and password of the communciation scenario SAP_COM_0510
     */
    'credentialsId',
    /**
     * Specifies the name of the Repository (Software Component) on the SAP Cloud Platform ABAP Environment system
     */
    'repositoryName',
    'cloudFoundry',
        /**
         * Cloud Foundry API endpoint.
         * @parentConfigKey cloudFoundry
         */
        'apiEndpoint',
        'credentialsId',
        /**
         * Cloud Foundry target organization.
         * @parentConfigKey cloudFoundry
         */
        'org',
        /**
         * Cloud Foundry target space.
         * @parentConfigKey cloudFoundry
         */
        'space',
        /**
         * Cloud Foundry service instance, for which the service key will be created.
         * @parentConfigKey cloudFoundry
         */
        'serviceInstance',
        /**
         * Cloud Foundry service key, which will be created.
         * @parentConfigKey cloudFoundry
         */
        'serviceKey',
    /** @see dockerExecute */
    'dockerImage',
    /** @see dockerExecute */
    'dockerWorkspace'
]
@Field Set GENERAL_CONFIG_KEYS = STEP_CONFIG_KEYS
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS
@Field Map CONFIG_KEY_COMPATIBILITY = [cloudFoundry: [apiEndpoint: 'cfApiEndpoint', credentialsId: 'cfCredentialsId', org: 'cfOrg', space: 'cfSpace', serviceInstance: 'cfServiceInstance', serviceKey: 'cfServiceKey']]
/**
 * Pulls a Repository (Software Component) to a SAP Cloud Platform ABAP Environment system.
 *
 * This is either possible by providing the host and the credentialsId of the communication arrangement or by providing access to a service key for the communication arrangement SAP_COM_0510 on cloud foundry.
 *
 * !!! note "Git Repository and Software Component"
 *       In SAP Cloud Platform ABAP Environment Git repositories are wrapped in Software Components (which are managed in the App "Manage Software Components")
 *       Currently, those two names are used synonymous.
 */
@GenerateDocumentation
void call(Map parameters = [:]) {
    handlePipelineStepErrors(stepName: STEP_NAME, stepParameters: parameters, failOnError: true) {

        def script = checkScript(this, parameters) ?: this
        Map configuration = ConfigurationHelper.newInstance(this, script)
            .loadStepDefaults()
            .mixinGeneralConfig(script.commonPipelineEnvironment, GENERAL_CONFIG_KEYS, CONFIG_KEY_COMPATIBILITY)
            .mixinStepConfig(script.commonPipelineEnvironment, STEP_CONFIG_KEYS, CONFIG_KEY_COMPATIBILITY)
            .mixinStageConfig(script.commonPipelineEnvironment, parameters.stageName ?: env.STAGE_NAME, STEP_CONFIG_KEYS, CONFIG_KEY_COMPATIBILITY)
            .mixin(parameters, PARAMETER_KEYS, CONFIG_KEY_COMPATIBILITY)
            .collectValidationFailures()
            .withMandatoryProperty('repositoryName')
            .use()

        String userColonPassword
        String urlString
        if (configuration.credentialsId != null && configuration.host != null) {
            echo "[${STEP_NAME}] Info: Using configuration: credentialsId: $configuration.credentialsId and host: $configuration.host"
            withCredentials([usernamePassword(credentialsId: configuration.credentialsId, usernameVariable: 'USER', passwordVariable: 'PASSWORD')]) {
                userColonPassword = "${USER}:${PASSWORD}"
                urlString = 'https://' + configuration.host + '/sap/opu/odata/sap/MANAGE_GIT_REPOSITORY/Pull'
            }
        } else {
            echo "[${STEP_NAME}] Info: Using Cloud Foundry service key $configuration.cloudFoundry.serviceKey for service instance $configuration.cloudFoundry.serviceInstance"
            dockerExecute(script:script,dockerImage: configuration.dockerImage, dockerWorkspace: configuration.dockerWorkspace) {
                String jsonString = getServiceKey(configuration)
                Map responseJson = readJSON(text : jsonString)
                userColonPassword = responseJson.abap.username + ":" + responseJson.abap.password
                urlString = responseJson.url + '/sap/opu/odata/sap/MANAGE_GIT_REPOSITORY/Pull'
            }
        }
        if (userColonPassword != null && urlString != null) {
            String authToken = userColonPassword.bytes.encodeBase64().toString()
            executeAbapEnvironmentPullGitRepo(configuration, urlString, authToken)
        } else {
            error "[${STEP_NAME}] Error: Necessary parameters not available"
        }
    }
}

private String getServiceKey(Map configuration) {

    String responseFile = "response-${UUID.randomUUID().toString()}.txt"
    withCredentials([
        usernamePassword(credentialsId: configuration.cloudFoundry.credentialsId, passwordVariable: 'CF_PASSWORD', usernameVariable: 'CF_USERNAME')
    ]) {
        bashScript =
            """#!/bin/bash
            set +x
            set -e
            export HOME=${configuration.dockerWorkspace}
            cf login -u ${BashUtils.quoteAndEscape(CF_USERNAME)} -p ${BashUtils.quoteAndEscape(CF_PASSWORD)} -a ${configuration.cloudFoundry.apiEndpoint} -o ${BashUtils.quoteAndEscape(configuration.cloudFoundry.org)} -s ${BashUtils.quoteAndEscape(configuration.cloudFoundry.space)};
            cf service-key ${BashUtils.quoteAndEscape(configuration.cloudFoundry.serviceInstance)} ${BashUtils.quoteAndEscape(configuration.cloudFoundry.serviceKey)} > \"${responseFile}\"
            """
        String responseString
        try {
            def status = sh returnStatus: true, script: bashScript
            if (status != 0) {
                echo "[${STEP_NAME}] Info: Could not get the service key $configuration.cloudFoundry.serviceKey for service instance $configuration.cloudFoundry.serviceInstance"
            }
            responseString = readFile(responseFile)
        } finally {
            sh "cf logout"
            sh script : """#!/bin/bash
                rm -f ${responseFile}
                """
        }
        def p = Pattern.compile(/\{.*\}$/, Pattern.MULTILINE | Pattern.DOTALL)
        def m = responseString =~ p
        String jsonString
        if (m.find()) {
            return m[0]
        } else {
            echo "[${STEP_NAME}] Info: Could not parse the service key $configuration.cloudFoundry.serviceKey"
            return null
        }
    }
}

private executeAbapEnvironmentPullGitRepo(Map configuration, String urlString, String authToken) {
    echo "[${STEP_NAME}] General Parameters: URL = \"${urlString}\", repositoryName = \"${configuration.repositoryName}\""
    HeaderFiles headerFiles = new HeaderFiles()
    try {
        String urlPullEntity = triggerPull(configuration, urlString, authToken, headerFiles)
        if (urlPullEntity != null) {
            String finalStatus = pollPullStatus(urlPullEntity, authToken, headerFiles)
            if (finalStatus != 'S') {
                error "[${STEP_NAME}] Pull Failed"
            }
        } else {
            error "[${STEP_NAME}] Pull Failed"
        }
    } finally {
        workspaceCleanup(headerFiles)
    }
}

private String triggerPull(Map configuration, String url, String authToken, HeaderFiles headerFiles) {
    String entityUri = null

    def xCsrfTokenScript = """#!/bin/bash
        curl -I -X GET ${url} \
        -H 'Authorization: Basic ${authToken}' \
        -H 'Accept: application/json' \
        -H 'x-csrf-token: fetch' \
        -D ${headerFiles.authFile} \
    """

    sh ( script : xCsrfTokenScript, returnStdout: true )

    HttpHeaderProperties headerProperties = new HttpHeaderProperties(readFile(headerFiles.authFile))
    checkRequestStatus(headerProperties)

    def scriptPull = """#!/bin/bash
        curl -X POST \"${url}\" \
        -H 'Authorization: Basic ${authToken}' \
        -H 'Accept: application/json' \
        -H 'Content-Type: application/json' \
        -H 'x-csrf-token: ${headerProperties.xCsrfToken}' \
        --cookie ${headerFiles.authFile} \
        -D ${headerFiles.postFile} \
        -d '{ \"sc_name\": \"${configuration.repositoryName}\" }'
    """
    def response = sh (
        script : scriptPull,
        returnStdout: true )

    checkRequestStatus(new HttpHeaderProperties(readFile(headerFiles.postFile)))

    JsonSlurper slurper = new JsonSlurper()
    Map responseJson = slurper.parseText(response)
    if (responseJson.d != null) {
        entityUri = responseJson.d.__metadata.uri.toString()
        echo "[${STEP_NAME}] Pull Status: ${responseJson.d.status_descr.toString()}"
    } else {
        error "[${STEP_NAME}] Error: ${responseJson?.error?.message?.value?.toString()?:'No message available'}"
    }

    echo "[${STEP_NAME}] Entity URI: ${entityUri}"
    return entityUri
}

private String pollPullStatus(String url, String authToken, HeaderFiles headerFiles) {
    String status = "R";
    while(status == "R") {

        Thread.sleep(5000)

        def pollScript = """#!/bin/bash
            curl -X GET "${url}" \
            -H 'Authorization: Basic ${authToken}' \
            -H 'Accept: application/json' \
            -D ${headerFiles.pollFile}
        """
        def pollResponse = sh (
            script : pollScript,
            returnStdout: true )

        checkRequestStatus(new HttpHeaderProperties(readFile(headerFiles.pollFile)))

        JsonSlurper slurper = new JsonSlurper()
        Map pollResponseJson = slurper.parseText(pollResponse)
        if (pollResponseJson.d != null) {
            status = pollResponseJson.d.status.toString()
        } else {
            error "[${STEP_NAME}] Error: ${pollResponseJson?.error?.message?.value?.toString()?:'No message available'}"
        }
        echo "[${STEP_NAME}] Pull Status: ${pollResponseJson.d.status_descr.toString()}"
    }
    return status
}

private void checkRequestStatus(HttpHeaderProperties httpHeader) {
    if (httpHeader.statusCode == 400) {
        echo "[${STEP_NAME}] Info: ${httpHeader.statusCode} ${httpHeader.statusMessage}"
    } else if (httpHeader.statusCode > 201) {
        error "[${STEP_NAME}] Error: ${httpHeader.statusCode} ${httpHeader.statusMessage}"
    }
}

private void workspaceCleanup(HeaderFiles headerFiles) {
    String cleanupScript = """#!/bin/bash
            rm -f ${headerFiles.authFile} ${headerFiles.postFile} ${headerFiles.pollFile}
        """
    sh ( script : cleanupScript )
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

    String authFile
    String postFile
    String pollFile

    HeaderFiles() {
        String uuid = UUID.randomUUID().toString()
        this.authFile = "headerFileAuth-${uuid}.txt"
        this.postFile = "headerFilePost-${uuid}.txt"
        this.pollFile = "headerFilePoll-${uuid}.txt"
    }
}
