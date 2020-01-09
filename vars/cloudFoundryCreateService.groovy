import com.sap.piper.GenerateDocumentation
import com.sap.piper.BashUtils
import com.sap.piper.JenkinsUtils
import com.sap.piper.Utils
import com.sap.piper.ConfigurationHelper

import groovy.transform.Field

import static com.sap.piper.Prerequisites.checkScript

@Field String STEP_NAME = 'cloudFoundryCreateService'

@Field Set STEP_CONFIG_KEYS = [
    'cloudFoundry',
        /**
         * Cloud Foundry API endpoint.
         * @parentConfigKey cloudFoundry
         */
        'apiEndpoint',
        /**
         * Credentials to be used for deployment.
         * @parentConfigKey cloudFoundry
         */
        'credentialsId',
        /**
         * Defines the manifest Yaml file that contains the information about the to be created services that will be passed to a Create-Service-Push cf cli plugin.
         * @parentConfigKey cloudFoundry
         */
        'serviceManifest',
        /**
         * Defines the manifest variables Yaml files to be used to replace variable references in manifest. This parameter
         * is optional and will default to `["manifest-variables.yml"]`. This can be used to set variable files like it
         * is provided by `cf push --vars-file <file>`.
         *
         * If the manifest is present and so are all variable files, a variable substitution will be triggered that uses
         * the `cfManifestSubstituteVariables` step before deployment. The format of variable references follows the
         * [Cloud Foundry standard](https://docs.cloudfoundry.org/devguide/deploy-apps/manifest-attributes.html#variable-substitution).
         * @parentConfigKey cloudFoundry
         */
        'manifestVariablesFiles',
        /**
         * Defines a `List` of variables as key-value `Map` objects used for variable substitution within the file given by `manifest`.
         * Defaults to an empty list, if not specified otherwise. This can be used to set variables like it is provided
         * by `cf push --var key=value`.
         *
         * The order of the maps of variables given in the list is relevant in case there are conflicting variable names and values
         * between maps contained within the list. In case of conflicts, the last specified map in the list will win.
         *
         * Though each map entry in the list can contain more than one key-value pair for variable substitution, it is recommended
         * to stick to one entry per map, and rather declare more maps within the list. The reason is that
         * if a map in the list contains more than one key-value entry, and the entries are conflicting, the
         * conflict resolution behavior is undefined (since map entries have no sequence).
         *
         * Note: variables defined via `manifestVariables` always win over conflicting variables defined via any file given
         * by `manifestVariablesFiles` - no matter what is declared before. This is the same behavior as can be
         * observed when using `cf push --var` in combination with `cf push --vars-file`.
         */
        'manifestVariables',
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
    /** @see dockerExecute */
    'dockerImage',
    /** @see dockerExecute */
    'dockerWorkspace',
    /** @see dockerExecute */
    'stashContent'
]

@Field Map CONFIG_KEY_COMPATIBILITY = [cloudFoundry: [apiEndpoint: 'cfApiEndpoint', appName:'cfAppName', credentialsId: 'cfCredentialsId', serviceManifest: 'cfServiceManifest', manifestVariablesFiles: 'cfManifestVariablesFiles', manifestVariables: 'cfManifestVariables',  org: 'cfOrg', space: 'cfSpace']]
@Field Set GENERAL_CONFIG_KEYS = STEP_CONFIG_KEYS
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS

/**
 * Step that uses the CF Create-Service-Push plugin to create services in a Cloud Foundry space. The information about the services is provided in a yaml file as infrastructure as code.
 * It is possible to use variable substitution inside of the yaml file like in a CF-push manifest yaml.
 *
 * For more details how to specify the services in the yaml see the [github page of the plugin](https://github.com/dawu415/CF-CLI-Create-Service-Push-Plugin).
 *
 * The `--no-push` options is always used with the plugin. To deploy the application make use of the cloudFoundryDeploy step!
 */
@GenerateDocumentation
void call(Map parameters = [:]) {
    handlePipelineStepErrors (stepName: STEP_NAME, stepParameters: parameters) {
        def script = checkScript(this, parameters) ?: this
        def utils = parameters.juStabUtils ?: new Utils()
        def jenkinsUtils = parameters.jenkinsUtilsStub ?: new JenkinsUtils()
        // load default & individual configuration
        Map config = ConfigurationHelper.newInstance(this, script)
            .loadStepDefaults()
            .mixinGeneralConfig(script.commonPipelineEnvironment, GENERAL_CONFIG_KEYS, CONFIG_KEY_COMPATIBILITY)
            .mixinStepConfig(script.commonPipelineEnvironment, STEP_CONFIG_KEYS, CONFIG_KEY_COMPATIBILITY)
            .mixinStageConfig(script.commonPipelineEnvironment, parameters.stageName?:env.STAGE_NAME, STEP_CONFIG_KEYS, CONFIG_KEY_COMPATIBILITY)
            .mixin(parameters, PARAMETER_KEYS, CONFIG_KEY_COMPATIBILITY)
            .withMandatoryProperty('cloudFoundry/org')
            .withMandatoryProperty('cloudFoundry/space')
            .withMandatoryProperty('cloudFoundry/credentialsId')
            .withMandatoryProperty('cloudFoundry/serviceManifest')
            .use()


        utils.pushToSWA([step: STEP_NAME],config)

        utils.unstashAll(config.stashContent)

        if (fileExists(config.cloudFoundry.serviceManifest)) {
            executeCreateServicePush(script, config)
        }
    }
}

private def executeCreateServicePush(script, Map config) {
    dockerExecute(script:script,dockerImage: config.dockerImage, dockerWorkspace: config.dockerWorkspace) {

        String varPart = varOptions(config)

        String varFilePart = varFileOptions(config)

        withCredentials([
            usernamePassword(credentialsId: config.cloudFoundry.credentialsId, passwordVariable: 'CF_PASSWORD', usernameVariable: 'CF_USERNAME')
        ]) {
            def returnCode = sh returnStatus: true, script: """#!/bin/bash
            set +x
            set -e
            export HOME=${config.dockerWorkspace}
            cf login -u ${BashUtils.quoteAndEscape(CF_USERNAME)} -p ${BashUtils.quoteAndEscape(CF_PASSWORD)} -a ${config.cloudFoundry.apiEndpoint} -o ${BashUtils.quoteAndEscape(config.cloudFoundry.org)} -s ${BashUtils.quoteAndEscape(config.cloudFoundry.space)};
            cf create-service-push --no-push --service-manifest ${BashUtils.quoteAndEscape(config.cloudFoundry.serviceManifest)}${varPart}${varFilePart}
            """
            sh "cf logout"
            if (returnCode!=0)  {
                error "[${STEP_NAME}] ERROR: The execution of the create-service-push plugin failed, see the logs above for more details."
            }
        }
    }
}

private varOptions(Map config) {
    String varPart = ''
    if (config.cloudFoundry.manifestVariables) {
        if (!(config.cloudFoundry.manifestVariables in List)) {
            error "[${STEP_NAME}] ERROR: Parameter config.cloudFoundry.manifestVariables is not a List!"
        }
        config.cloudFoundry.manifestVariables.each {
            if (!(it in Map)) {
                error "[${STEP_NAME}] ERROR: Parameter config.cloudFoundry.manifestVariables.$it is not a Map!"
            }
            it.keySet().each { varKey ->
                String varValue=BashUtils.quoteAndEscape(it.get(varKey).toString())
                varPart += " --var $varKey=$varValue"
            }
        }
    }
    if (varPart) echo "We will add the following string to the cf push call: '$varPart'"
    return varPart
}

private String varFileOptions(Map config) {
    String varFilePart = ''
    if (config.cloudFoundry.manifestVariablesFiles) {
        if (!(config.cloudFoundry.manifestVariablesFiles in List)) {
            error "[${STEP_NAME}] ERROR: Parameter config.cloudFoundry.manifestVariablesFiles is not a List!"
        }
        config.cloudFoundry.manifestVariablesFiles.each {
            if (fileExists(it)) {
                varFilePart += " --vars-file ${BashUtils.quoteAndEscape(it)}"
            } else {
                echo "[${STEP_NAME}] [WARNING] We skip adding not-existing file '$it' as a vars-file to the cf create-service-push call"
            }
        }
    }
    if (varFilePart) echo "We will add the following string to the cf push call: '$varFilePart'"
    return varFilePart
}
