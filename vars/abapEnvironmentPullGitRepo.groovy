import static com.sap.piper.Prerequisites.checkScript
import com.sap.piper.ConfigurationHelper
import com.sap.piper.GenerateDocumentation
import com.sap.piper.JenkinsUtils
import com.sap.piper.Utils
import com.sap.piper.BashUtils
import groovy.json.JsonSlurper
import hudson.AbortException
import groovy.transform.Field
import org.jenkinsci.plugins.workflow.steps.FlowInterruptedException
import java.util.UUID
import java.util.regex.*

@Field def STEP_NAME = getClass().getName()
@Field Set GENERAL_CONFIG_KEYS = [
    /**
     * Specifies the host address of the SAP Cloud Platform ABAP Environment system
     */
    'host',
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
        /**
         * Cloud Foundry credentials.
         * @parentConfigKey cloudFoundry
         */
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
@Field Set STEP_CONFIG_KEYS = GENERAL_CONFIG_KEYS.plus([
    /**
     * Jenkins CredentialsId containing the communication user and password of the communciation scenario SAP_COM_0510
     */
    'credentialsId'
])
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS
/**
 * Pulls a Repository (Software Component) to a SAP Cloud Platform ABAP Environment system.
 *
 * !!! note "Git Repository and Software Component"
 *       In SAP Cloud Platform ABAP Environment Git repositories are wrapped in Software Components (which are managed in the App "Manage Software Components")
 *       Currently, those two names are used synonymous.
 * !!! note "User and Password"
 *        In the future, we want to support the user / password creation via the create-service-key funcion of cloud foundry.
 *        For this case, it is not possible to use the usual pattern with Jenkins Credentials.
 */
@GenerateDocumentation
void call(Map parameters = [:]) {

    handlePipelineStepErrors(stepName: STEP_NAME, stepParameters: parameters, failOnError: true) {

        def script = checkScript(this, parameters) ?: this

        // In the future, we want to support the user / password creation via the create-service-key funcion of cloud foundry.
        // For this case, it is not possible to use the usual pattern with Jenkins Credentials.
        Map configuration = ConfigurationHelper.newInstance(this)
            .mixinGeneralConfig(script.commonPipelineEnvironment, GENERAL_CONFIG_KEYS)
            .mixinStepConfig(script.commonPipelineEnvironment, STEP_CONFIG_KEYS)
            .mixinStageConfig(script.commonPipelineEnvironment, parameters.stageName ?: env.STAGE_NAME, STEP_CONFIG_KEYS)
            .mixin(parameters, PARAMETER_KEYS)
            .collectValidationFailures()
            .withMandatoryProperty('repositoryName', 'Repository / Software Component not provided')
            .use()

        String authToken
        String urlString
        if (configuration.credentialsId != null) {
            withCredentials([usernamePassword(credentialsId: configuration.credentialsId, usernameVariable: 'USER', passwordVariable: 'PASSWORD')]) {
                String userColonPassword = "${USER}:${PASSWORD}"
                authToken = userColonPassword.bytes.encodeBase64().toString()
                executeAbapEnvironmentPullGitRepo(configuration, urlString, authToken)
            }
        urlString = 'https://' + configuration.host + '/sap/opu/odata/sap/MANAGE_GIT_REPOSITORY/Pull'
        } else {
            // try {
                dockerExecute(script:script,dockerImage: configuration.dockerImage, dockerWorkspace: configuration.dockerWorkspace) {
                        String jsonString = getServiceKey(configuration)
                        def responseJson = readJSON(jsonString)
                        String userColPw = responseJson.abap.username + ":" + responseJson.abap.password
                        authToken = userColPw.bytes.encodeBase64().toString()
                        urlString = responseJson.url + '/sap/opu/odata/sap/MANAGE_GIT_REPOSITORY/Pull'
                        echo authToken
                        echo urlString
                        executeAbapEnvironmentPullGitRepo(configuration, urlString, authToken)
                }
            // } catch (err) {
            //     error "[${STEP_NAME}] Error: Could not get credentials from Cloud Foundry"
            // }
        }
    }
}

private String getServiceKey(Map configuration) {
    withCredentials([
        usernamePassword(credentialsId: configuration.cloudFoundry.credentialsId, passwordVariable: 'CF_PASSWORD', usernameVariable: 'CF_USERNAME')
    ]) {
        bashScript =
            """#!/bin/bash
            set +x
            set -e
            export HOME=${configuration.dockerWorkspace}
            cf login -u ${BashUtils.quoteAndEscape(CF_USERNAME)} -p ${BashUtils.quoteAndEscape(CF_PASSWORD)} -a ${configuration.cloudFoundry.apiEndpoint} -o ${BashUtils.quoteAndEscape(configuration.cloudFoundry.org)} -s ${BashUtils.quoteAndEscape(configuration.cloudFoundry.space)};
            cf service-key ${BashUtils.quoteAndEscape(configuration.cloudFoundry.serviceInstance)} ${BashUtils.quoteAndEscape(configuration.cloudFoundry.serviceKey)} > \"response.json\"
            """
        def status = sh returnStatus: true, script: bashScript
        sh "cf logout"
        echo status.toString()
        String responseString = readFile("response.json")
        echo responseString
        def p = Pattern.compile(/\{.*\}$/, Pattern.MULTILINE | Pattern.DOTALL)
        def m = responseString =~ p
        String jsonString
        if (m.find()) {
            return m[0]
        } else {
            error "No REGEX match"
        }
    }
}

private String getAuthToken()

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
