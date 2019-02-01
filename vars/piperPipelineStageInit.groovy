import com.sap.piper.ConfigurationHelper
import com.sap.piper.Utils
import groovy.transform.Field

import static com.sap.piper.Prerequisites.checkScript

@Field String STEP_NAME = getClass().getName()

@Field Set GENERAL_CONFIG_KEYS = [
    'buildTool',
    'productiveBranch',
    'stashSettings',
    'verbose'
]
@Field Set STEP_CONFIG_KEYS = GENERAL_CONFIG_KEYS
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS

void call(Map parameters = [:]) {

    def script = checkScript(this, parameters) ?: this
    def utils = parameters.juStabUtils ?: new Utils()

    def stageName = parameters.stageName?:env.STAGE_NAME

    piperStageWrapper (script: script, stageName: stageName, stashContent: [], ordinal: 1) {
        checkout scm

        setupCommonPipelineEnvironment script: script, customDefaults: parameters.customDefaults

        Map config = ConfigurationHelper.newInstance(this)
            .loadStepDefaults()
            .mixinGeneralConfig(script.commonPipelineEnvironment, GENERAL_CONFIG_KEYS)
            .mixinStageConfig(script.commonPipelineEnvironment, stageName, STEP_CONFIG_KEYS)
            .mixin(parameters, PARAMETER_KEYS)
            .withMandatoryProperty('buildTool')
            .withMandatoryProperty('stashSettings')
            .use()

        //perform stashing based on libray resource piper-stash-settings.yml if not configured otherwise
        initStashConfiguration(script, config)

        if (config.verbose) {
            echo "piper-lib-os  configuration: ${script.commonPipelineEnvironment.configuration}"
        }

        // telemetry reporting
        utils.pushToSWA([step: STEP_NAME], config)

        checkBuildTool(config)

        piperInitRunStageConfiguration script: script

        if (env.BRANCH_NAME == config.productiveBranch) {
            setVersion script: script
        }

        pipelineStashFilesBeforeBuild script: script

    }
}

private void checkBuildTool(config) {
    def buildDescriptorPattern = ''
    switch (config.buildTool) {
        case 'maven':
            buildDescriptorPattern = 'pom.xml'
            break
        case 'npm':
            buildDescriptorPattern = 'package.json'
            break
        case 'mta':
            buildDescriptorPattern = 'mta.yaml'
            break
    }
    if (buildDescriptorPattern && !findFiles(glob: buildDescriptorPattern)) {
        error "[${STEP_NAME}] buildTool configuration '${config.buildTool}' does not fit to your project, please set buildTool as genereal setting in your .pipeline/config.yml correctly, see also https://github.wdf.sap.corp/pages/ContinuousDelivery/piper-doc/configuration/"
    }
}

private void initStashConfiguration (script, config) {
    Map stashConfiguration = readYaml(text: libraryResource(config.stashSettings))
    echo "Stash config: stashConfiguration"
    script.commonPipelineEnvironment.configuration.stageStashes = stashConfiguration
}
