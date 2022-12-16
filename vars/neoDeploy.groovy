import com.cloudbees.groovy.cps.NonCPS

import com.sap.piper.GenerateDocumentation
import com.sap.piper.ConfigurationHelper
import com.sap.piper.Utils
import com.sap.piper.StepAssertions
import com.sap.piper.tools.neo.DeployMode
import com.sap.piper.tools.neo.NeoCommandHelper
import com.sap.piper.tools.neo.WarAction
import groovy.transform.Field

import static com.sap.piper.Prerequisites.checkScript

@Field String STEP_NAME = getClass().getName()

@Field Set GENERAL_CONFIG_KEYS = [
    'neo',
    /**
     * The SAP BTP account to deploy to.
     * @parentConfigKey neo
     * @mandatory for deployMode=warParams
     */
    'account',
    /**
     * Name of the application you want to manage, configure, or deploy.
     * @parentConfigKey neo
     * @mandatory for deployMode=warParams
     */
    'application',
    /**
     * The Jenkins credentials containing either user and password (UsernamePassword type credential) or json containing clientId, client secret and oauth service url (SecretFile type credential) used for SAP CP deployment.
     * @parentConfigKey neo
     */
    'credentialsId',
    /**
     * The Jenkins credential of type 'UsernamePassword' or 'SecretFile'.
     * @parentConfigKey neo
     */
    'credentialType',
    /**
     * Map of environment variables in the form of KEY: VALUE.
     * @parentConfigKey neo
     */
    'environment',
    /**
     * The SAP BTP host to deploy to.
     * @parentConfigKey neo
     * @mandatory for deployMode=warParams
     */
    'host',
    /**
     * The path to the .properties file in which all necessary deployment properties for the application are defined.
     * @parentConfigKey neo
     * @mandatory for deployMode=warPropertiesFile
     */
    'propertiesFile',
    /**
     * Name of SAP BTP application runtime.
     * @parentConfigKey neo
     * @mandatory for deployMode=warParams
     */
    'runtime',
    /**
     * Version of SAP BTP application runtime.
     * @parentConfigKey neo
     * @mandatory for deployMode=warParams
     */
    'runtimeVersion',
    /**
     * Compute unit (VM) size. Acceptable values: lite, pro, prem, prem-plus.
     * @parentConfigKey neo
     */
    'size',
    /**
     * String of VM arguments passed to the JVM.
     * @parentConfigKey neo
     */
    'vmArguments',
    /**
     * Boolean to enable/disable invalidating the cache after deployment.
     * @possibleValues `true`, `false`
     * @parentConfigKey neo
     */
    'invalidateCache',
    /**
     * Portal landscape region subscribed to in SAP BTP.
     * @parentConfigKey neo
     */
    'portalLandscape',
    /**
     * UsernamePassword type credential containing SAP BTP OAuth client ID and client secret.
     * @parentConfigKey neo
     */
    'oauthCredentialId',
    /**
     * Site ID of the SAP Fiori Launchpad containing the SAP Fiori app. If not set, the cache of the default site, as defined in the Portal service, is invalidated.
     * @parentConfigKey neo
     */
    'siteId',
    /**
     * Availability zone of BTP NEO into which the application is deployed.
     * @possibleValues 1, 2
     * @parentConfigKey neo
     */
    'azDistribution'
]

@Field Set STEP_CONFIG_KEYS = GENERAL_CONFIG_KEYS.plus([
    /**
     * The deployment mode which should be used. Available options are:
     * *`'mta'` - default,
     * *`'warParams'` - deploying WAR file and passing all the deployment parameters via the function call,
     * *`'warPropertiesFile'` - deploying WAR file and putting all the deployment parameters in a .properties file.
     * @possibleValues 'mta', 'warParams', 'warPropertiesFile'
     */
    'deployMode',
    /**
     * @see dockerExecute
     */
    'dockerEnvVars',
    /**
     * @see dockerExecute
     */
    'dockerImage',
    /**
     * @see dockerExecute
     */
    'dockerOptions',
    /**
      * Extension files. Provided to the neo command via parameter `--extensions` (`-e`). Only valid for deploy mode `mta`.
      */
    'extensions',
    /**
     * The path to the archive for deployment to SAP CP. If not provided the following defaults are used based on the deployMode:
     * *`'mta'` - The `mtarFilePath` from common pipeline environment is used instead.
     * *`'warParams'` and `'warPropertiesFile'` - The following template will be used "<mavenDeploymentModule>/target/<artifactId>.<packaging>"
     */
    'source',
    /**
     * Path to the maven module which contains the deployment artifact.
     */
    'mavenDeploymentModule'
])

@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS.plus([
    /**
     * Action mode when using WAR file mode. Available options are `deploy` (default) and `rolling-update` which performs update of an application without downtime in one go.
     * @possibleValues 'deploy', 'rolling-update'
     */
    'warAction'
])

/**
 * Deploys an Application to SAP BTP (SAP CP) using the SAP BTP Console Client (Neo Java Web SDK).
 */
@GenerateDocumentation
void call(parameters = [:]) {
    handlePipelineStepErrors(stepName: STEP_NAME, stepParameters: parameters) {

        def script = checkScript(this, parameters) ?: this
        def utils = parameters.utils ?: new Utils()
        String stageName = parameters.stageName ?: env.STAGE_NAME

        // load default & individual configuration
        ConfigurationHelper configHelper = ConfigurationHelper.newInstance(this)
            .loadStepDefaults([:], stageName)
            .mixinGeneralConfig(script.commonPipelineEnvironment, GENERAL_CONFIG_KEYS)
            .mixinStepConfig(script.commonPipelineEnvironment, STEP_CONFIG_KEYS)
            .mixinStageConfig(script.commonPipelineEnvironment, stageName, STEP_CONFIG_KEYS)
            .mixin(parameters, PARAMETER_KEYS)
            .collectValidationFailures()
            .withPropertyInValues('deployMode', DeployMode.stringValues())

        Map configuration = configHelper.use()

        DeployMode deployMode = DeployMode.fromString(configuration.deployMode)

        def isWarParamsDeployMode = { deployMode == DeployMode.WAR_PARAMS },
            isNotWarPropertiesDeployMode = {deployMode != DeployMode.WAR_PROPERTIES_FILE}

        if(!configuration.source){
            configHelper.mixin([source: getDefaultSource(script, configuration, deployMode)])
        }

        configuration = configHelper
            .withMandatoryProperty('source')
            .withMandatoryProperty('neo/credentialsId')
            .withMandatoryProperty('neo/credentialType')
            .withMandatoryProperty('neo/application', null, isWarParamsDeployMode)
            .withMandatoryProperty('neo/runtime', null, isWarParamsDeployMode)
            .withMandatoryProperty('neo/runtimeVersion', null, isWarParamsDeployMode)
            .withMandatoryProperty('neo/host', null, isNotWarPropertiesDeployMode)
            .withMandatoryProperty('neo/account', null, isNotWarPropertiesDeployMode)
            .use()

        Set extensionFileNames

        if(configuration.extensions == null) {
            extensionFileNames = []
        } else {
            extensionFileNames = configuration.extensions in Collection ? configuration.extensions : [configuration.extensions]
        }

        if( ! extensionFileNames.findAll { it == null || it.isEmpty() }.isEmpty() )
            error "At least one extension file name was null or empty: ${extensionFileNames}."

        if(deployMode != DeployMode.MTA && ! extensionFileNames.isEmpty())
            error "Extensions (${extensionFileNames} found for deploy mode ${deployMode}. Extensions are only supported for deploy mode '${DeployMode.MTA}')"

        utils.pushToSWA([
            step: STEP_NAME,
            stepParamKey1: 'deployMode',
            stepParam1: configuration.deployMode == 'mta'?'mta':'war', // ['mta', 'warParams', 'warPropertiesFile']
            stepParamKey2: 'warAction',
            stepParam2: configuration.warAction == 'rolling-update'?'blue-green':'standard', // ['deploy', 'deploy-mta', 'rolling-update']
            stepParamKey3: 'scriptMissing',
            stepParam3: parameters?.script == null,
        ], configuration)

        if(configuration.neo.credentialType == 'UsernamePassword'){
            withCredentials([usernamePassword(
                credentialsId: configuration.neo.credentialsId,
                passwordVariable: 'NEO_PASSWORD',
                usernameVariable: 'NEO_USERNAME')]) {

                assertPasswordRules(NEO_PASSWORD)

                dockerExecute(
                    script: script,
                    dockerImage: configuration.dockerImage,
                    dockerEnvVars: configuration.dockerEnvVars,
                    dockerOptions: configuration.dockerOptions
                ) {

                    StepAssertions.assertFileExists(this, configuration.source)

                    for(CharSequence extensionFile in extensionFileNames) {
                        StepAssertions.assertFileExists(this, extensionFile)
                    }

                    NeoCommandHelper neoCommandHelper = new NeoCommandHelper(
                        this,
                        deployMode,
                        configuration.neo,
                        extensionFileNames,
                        NEO_USERNAME,
                        NEO_PASSWORD,
                        configuration.source
                    )

                    lock("$STEP_NAME:${neoCommandHelper.resourceLock()}") {
                        deploy(script, configuration, neoCommandHelper, configuration.dockerImage, deployMode)
                    }
                    if(configuration.neo.invalidateCache == true) {
                        if (configuration.deployMode == 'mta') {
                            echo "Triggering invalidation of cache for html5 applications"
                            invalidateCache(configuration)
                        } else {
                            echo "Invalidation of cache is ignored. It is performed only for html5 applications."
                        }
                    }
                }
            }
        }
        else if(configuration.neo.credentialType == 'SecretFile'){

            dockerExecute(
                script: script,
                dockerImage: configuration.dockerImage,
                dockerEnvVars: configuration.dockerEnvVars,
                dockerOptions: configuration.dockerOptions
            ) {
                withCredentials([file(credentialsId: configuration.neo.credentialsId, variable: 'oauth_deploy_cred')]) {
                    deployWithBearerToken(oauth_deploy_cred, configuration, script)
                }
            }
        }
        else {
            error "Unsupported type of neo deploy credential."
        }
    }
}

private invalidateCache(configuration){
    def account = configuration.neo.account
    def host = configuration.neo.host
    def portalLandscape = configuration.neo.portalLandscape

    withCredentials([usernamePassword(
        credentialsId: configuration.neo.oauthCredentialId,
        passwordVariable: 'OAUTH_NEO_CLIENT_SECRET',
        usernameVariable: 'OAUTH_NEO_CLIENT_ID')]) {
        def bearerTokenResponse = sh(
            script: """#!/bin/bash
                        curl -X POST -u "${OAUTH_NEO_CLIENT_ID}:${OAUTH_NEO_CLIENT_SECRET}" \
                            --fail \
                            "https://oauthasservices-${account}.${host}/oauth2/api/v1/token?grant_type=client_credentials&scope=write,read"
                    """,
            returnStdout: true)
        def bearerToken = readJSON(text: bearerTokenResponse).access_token

        echo "Retrieved bearer token."

        def fetchXcsrfTokenResponse = sh(
            script: """#!/bin/bash
                        curl -i -L \
                            -c 'cookies.jar' \
                            -H 'X-CSRF-Token: Fetch' \
                            -H "Authorization: Bearer ${bearerToken}" \
                            --fail \
                            "https://${portalLandscape}-${account}.${host}/fiori/api/v1/csrf"
                    """,
            returnStdout: true)

        def xcsrfToken = readProperties(text: fetchXcsrfTokenResponse)["X-CSRF-Token"]
        def siteId = configuration.neo.siteId ?: ""

        if(! siteId){
            echo "Using the default site defined in Portal service and invalidating the cache."
        }
        else{
            echo "Invalidating the cache for site with Id: ${siteId}."
        }
        def statusCode = sh(
            script: """#!/bin/bash
                        curl -X POST -L \
                            -b 'cookies.jar'  \
                            -H "X-CSRF-Token: ${xcsrfToken}" \
                            -H "Authorization: Bearer ${bearerToken}" \
                            -d "{\"siteId\":${siteId}}" \
                            -so /dev/null \
                            -w '%{response_code}' \
                            "https://${portalLandscape}-${account}.${host}/fiori/v1/operations/invalidateCache"
                    """,
            returnStdout: true).trim()

        if(! siteId && statusCode == "500") {
            error "Invalidating the cache failed. " +
                    "As no siteId is set, the default site defined in the portal UI is used. " +
                    "Please verify a default site is defined in Portal service. " +
                    "Alternatively, configure the siteId parameter for this step to invalidate the cache of that specific site."
        } else if(! statusCode == "200" || ! statusCode == "201" ){
            error "Invalidating the cache failed with response code: ${statusCode}."
        }
        echo "Successfully invalidated the cache."
    }
}

private deploy(script, Map configuration, NeoCommandHelper neoCommandHelper, dockerImage, DeployMode deployMode) {

    String logFolder = "logs/neo/${UUID.randomUUID()}"

    try {
        sh "mkdir -p ${logFolder}"
        withEnv(["neo_logging_location=${pwd()}/${logFolder}"]) {
            if (deployMode.isWarDeployment()) {
                ConfigurationHelper.newInstance(this, configuration).withPropertyInValues('warAction', WarAction.stringValues())
                WarAction warAction = WarAction.fromString(configuration.warAction)

                if (warAction == WarAction.ROLLING_UPDATE) {
                    if (!isAppRunning(neoCommandHelper)) {
                        warAction = WarAction.DEPLOY
                        echo "Rolling update not possible because application is not running. Falling back to standard deployment."
                    }
                }

                echo "Link to the application dashboard: ${neoCommandHelper.cloudCockpitLink()}"

                if (warAction == WarAction.ROLLING_UPDATE) {
                    try {
                        sh neoCommandHelper.rollingUpdateCommand()
                    } catch (e) {
                        error "[ERROR][${STEP_NAME}] The execution of the deploy command failed, see the log for details."
                    }
                } else {
                    try {
                        sh neoCommandHelper.deployCommand()
                    } catch (e) {
                        error "[ERROR][${STEP_NAME}] The execution of the deploy command failed, see the log for details."
                    }
                    sh neoCommandHelper.restartCommand()
                }


            } else if (deployMode == DeployMode.MTA) {
                try {
                    sh neoCommandHelper.deployMta()
                } catch (e) {
                    error "[ERROR][${STEP_NAME}] The execution of the deploy command failed, see the log for details."
                }
            }
        }
    }
    catch (Exception ex) {

        echo "Error while deploying to SAP BTP. Here are the neo.sh logs:"
        try {
            sh "cat ${logFolder}/*"
        } catch(Exception e) {
            echo "Unable to provide the logs."
            ex.addSuppressed(e)
        }
        throw ex
    }
}

private deployWithBearerToken(def credentialFilePath, Map configuration, Script script){

    def deployArchive = script.commonPipelineEnvironment.getMtarFilePath()
    def host = configuration.neo.host
    def account = configuration.neo.account

    def credentialFileContent = readFile(credentialFilePath)
    def credentialsMap = parseJson(credentialFileContent)

    def oauthClientId = credentialsMap.oauthClientId
    def oauthClientSecret = credentialsMap.oauthClientSecret
    def oauthUrl = credentialsMap.oauthServiceUrl

    echo "[${STEP_NAME}] Retrieving oauth token..."

    def myCurl = "curl --fail --silent --show-error --retry 12"
    def token_json = sh(
        script: """#!/bin/bash
                    ${myCurl} -XPOST -u \"${oauthClientId}:${oauthClientSecret}\" \"${oauthUrl}/apitoken/v1?grant_type=client_credentials"
                """,
        returnStdout: true
    )
    def responseJson = readJSON text: token_json
    def token = responseJson.access_token

    echo "[${STEP_NAME}] Deploying '${deployArchive}' to '${account}'..."

    def deploymentContentResponse = sh(
        script: """#!/bin/bash
                    ${myCurl} -XPOST -H \"Authorization: Bearer ${token}\" -F file=@\"${deployArchive}\" \"https://slservice.${host}/slservice/v1/oauth/accounts/${account}/mtars\"
                """,
        returnStdout: true
    )
    def deploymentJson = readJSON text: deploymentContentResponse
    def deploymentId = deploymentJson.id

    echo "[${STEP_NAME}] Deployment Id is '${deploymentId}'."

    def statusPollScript = """#!/bin/bash
                                ${myCurl} -XGET -H \"Authorization: Bearer ${token}\" \"https://slservice.${host}/slservice/v1/oauth/accounts/${account}/mtars/${deploymentId}\"
                            """
    def statusResponse = sh(script: statusPollScript, returnStdout: true)
    def statusJson = readJSON text: statusResponse
    def state = statusJson.state

    while (state == 'RUNNING') {
        sleep(10)
        statusResponse = sh(script: statusPollScript, returnStdout: true)
        statusJson = readJSON text: statusResponse
        state = statusJson.state
        echo "${STEP_NAME}] Deployment is still running..."
    }

    if (state == 'DONE') {
        echo "[${STEP_NAME}] Deployment has succeeded."
    } else if (state == 'FAILED') {
        if(statusJson.progress[0]?.modules[0]?.error?.internalMessage) {
            def message = statusJson.progress[0].modules[0].error.internalMessage
            echo "[${STEP_NAME}] Deployment has failed with the message: ${message}"
            error "[${STEP_NAME}] Deployment failure message: ${message}"
        } else {
            echo "[${STEP_NAME}] Deployment has failed with response: ${statusResponse}"
            error "[${STEP_NAME}] Deployment failure reason: ${statusResponse}"
        }
    } else {
        echo "[${STEP_NAME}] Unknown status '${state}'"
        error "[${STEP_NAME}] Deployment failed with unknown status: ${state}"
    }
}

private boolean isAppRunning(NeoCommandHelper commandHelper) {
    def status = sh script: "${commandHelper.statusCommand()} || true", returnStdout: true
    return status.contains('Status: STARTED')
}

private assertPasswordRules(String password) {
    if (password.startsWith("@")) {
        error("Your password for the deployment to SAP BTP contains characters which are not " +
            "supported by the neo tools. " +
            "For example it is not allowed that the password starts with @. " +
            "Please consult the documentation for the neo command line tool for more information: " +
            "https://help.sap.com/viewer/65de2977205c403bbc107264b8eccf4b/Cloud/en-US/8900b22376f84c609ee9baf5bf67130a.html")
    }
}

private getDefaultSource(Script script, Map configuration, DeployMode deployMode){
    if(deployMode == DeployMode.MTA) {
        return script.commonPipelineEnvironment.getMtarFilePath()
    }

    String pomFile = "${configuration.mavenDeploymentModule}/pom.xml"

    if(!fileExists(pomFile)){
        error("The configured mavenDeploymentModule (${configuration.mavenDeploymentModule}) does not contain a pom file.")
    }

    def pom = readMavenPom file: pomFile

    String source = "${configuration.mavenDeploymentModule}/target/${pom.artifactId}.${pom.packaging}"

    return source
}

//Convert LazyMap instance produced after jsonSluper to a groovy based LinkedHashMap to overcome serialization issue
@NonCPS
def parseJson(credentialFileContent) {
    def lazyMap = new groovy.json.JsonSlurper().parseText(credentialFileContent)
    def map = [:]
    for (prop in lazyMap) {
        map[prop.key] = prop.value
    }
    return map
}
