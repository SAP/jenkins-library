import groovy.transform.Field
import com.cloudbees.groovy.cps.NonCPS
import com.sap.piper.GenerateStageDocumentation
import groovy.transform.Field
import com.sap.piper.Utils
import com.sap.piper.ConfigurationHelper
import com.sap.piper.ConfigurationLoader

import static com.sap.piper.Prerequisites.checkScript

@Field String STEP_NAME = getClass().getName()
@Field Set GENERAL_CONFIG_KEYS = []
@Field STAGE_STEP_KEYS = [
    'cloudFoundryCreateServiceKey',
    'abapEnvironmentAssemblePackages',
    'abapEnvironmentBuild',
    'abapAddonAssemblyKitRegisterPackages',
    'abapAddonAssemblyKitReleasePackages',
    'abapEnvironmentAssembleConfirm',
    'abapAddonAssemblyKitCreateTargetVector',
    'abapAddonAssemblyKitPublishTargetVector',
    /** Parameter for host config */
    'host',
    'generateTagForAddonProductVersion',
    'generateTagForAddonComponentVersion'
]
@Field Set STEP_CONFIG_KEYS = GENERAL_CONFIG_KEYS.plus(STAGE_STEP_KEYS)
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS
/**
 * This stage builds an AddOn for the SAP Cloud Platform ABAP Environment
 */
void call(Map parameters = [:]) {
    def script = checkScript(this, parameters) ?: this
    def stageName = parameters.stageName?:env.STAGE_NAME

    // load default & individual configuration
    Map config = ConfigurationHelper.newInstance(this)
        .loadStepDefaults([:], stageName)
        .mixin(ConfigurationLoader.defaultStageConfiguration(script, stageName))
        .mixinGeneralConfig(script.commonPipelineEnvironment, GENERAL_CONFIG_KEYS)
        .mixinStepConfig(script.commonPipelineEnvironment, STEP_CONFIG_KEYS)
        .mixinStageConfig(script.commonPipelineEnvironment, stageName, STEP_CONFIG_KEYS)
        .mixin(parameters, PARAMETER_KEYS)
        .use()

    piperStageWrapper (script: script, stageName: stageName, stashContent: [], stageLocking: false) {
        if (!config.host) {
            cloudFoundryCreateServiceKey script: parameters.script
        }
        abapEnvironmentAssemblePackages script: parameters.script
        abapEnvironmentBuild(script: parameters.script, phase: 'GENERATION', downloadAllResultFiles: true, useFieldsOfAddonDescriptor: '[{"use":"Name","renameTo":"SWC"}]')
        abapAddonAssemblyKitRegisterPackages script: parameters.script
        abapAddonAssemblyKitReleasePackages script: parameters.script
        abapEnvironmentAssembleConfirm script: parameters.script
        abapAddonAssemblyKitCreateTargetVector script: parameters.script
        abapAddonAssemblyKitPublishTargetVector(script: parameters.script, targetVectorScope: 'T')
        if (config.generateTagForAddonComponentVersion || config.generateTagForAddonProductVersion) {
            try {
                Set keys = [ 'cfServiceKeyName' ]
                Map configClone = ConfigurationHelper.newInstance(this)
                    // .loadStepDefaults([:], 'Clone Repositories')
                    .mixin(ConfigurationLoader.defaultStageConfiguration(script, 'Clone Repositories'))
                    .mixinGeneralConfig(script.commonPipelineEnvironment, keys)
                    // .mixinStepConfig(script.commonPipelineEnvironment, keys)
                    .mixinStageConfig(script.commonPipelineEnvironment, 'Clone Repositories', keys)
                    // .mixin(parameters, keys)
                    .use()
                    configClone.each{ k, v -> println "${k}:${v}" }
                // Map configClone2 = ConfigurationHelper.newInstance(this)
                //     // .loadStepDefaults([:], 'Clone Repositories')
                //     .mixin(ConfigurationLoader.defaultStageConfiguration(script, 'Clone Repositories'))
                //     .mixinGeneralConfig(script.commonPipelineEnvironment, keys)
                //     // .mixinStepConfig(script.commonPipelineEnvironment, keys)
                //     // .mixinStageConfig(script.commonPipelineEnvironment, 'Clone Repositories', keys)
                //     // .mixin(parameters, keys)
                //     .use()
                //     configClone2.each{ k, v -> println "${k}:${v}" }
                // Map configClone3 = ConfigurationHelper.newInstance(this)
                //     // .loadStepDefaults([:], 'Clone Repositories')
                //     .mixin(ConfigurationLoader.defaultStageConfiguration(script, 'Clone Repositories'))
                //     // .mixinGeneralConfig(script.commonPipelineEnvironment, keys)
                //     // .mixinStepConfig(script.commonPipelineEnvironment, keys)
                //     // .mixinStageConfig(script.commonPipelineEnvironment, 'Clone Repositories', keys)
                //     // .mixin(parameters, keys)
                //     .use()
                //     configClone3.each{ k, v -> println "${k}:${v}" }
                abapEnvironmentCreateTag(script: parameters.script, cfServiceKeyName: configClone.cfServiceKeyName)
            } catch (e) {
                echo 'Tag creation failed: ' + e.message
            }
        }
    }
}
