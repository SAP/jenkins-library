import com.sap.piper.ConfigurationHelper
import com.sap.piper.GenerateStageDocumentation
import com.sap.piper.StageNameProvider
import com.sap.piper.Utils
import groovy.transform.Field

import static com.sap.piper.Prerequisites.checkScript

@Field String STEP_NAME = getClass().getName()
@Field String TECHNICAL_STAGE_NAME = 'security'

@Field Set GENERAL_CONFIG_KEYS = []
@Field STAGE_STEP_KEYS = [
    /** Executes a Checkmarx scan */
    'checkmarxExecuteScan',
    /** Executes BlackDuck Detect scans */
    'detectExecuteScan',
    /** Executes a Fortify scan */
    'fortifyExecuteScan',
    /** Executes a WhiteSource scan */
    'whitesourceExecuteScan'
]
@Field Set STEP_CONFIG_KEYS = GENERAL_CONFIG_KEYS.plus(STAGE_STEP_KEYS)
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS

/**
 * In this stage important security-relevant checks will be conducted.<br />
 * This is to achieve a decent level of security for your application.
 */
@GenerateStageDocumentation(defaultStageName = 'Security')
void call(Map parameters = [:]) {
    def script = checkScript(this, parameters) ?: this
    def utils = parameters.juStabUtils ?: new Utils()
    def stageName = StageNameProvider.instance.getStageName(script, parameters, this)

    def securityScanMap = [:]

    Map config = ConfigurationHelper.newInstance(this)
        .loadStepDefaults()
        .mixinGeneralConfig(script.commonPipelineEnvironment, GENERAL_CONFIG_KEYS)
        .mixinStageConfig(script.commonPipelineEnvironment, stageName, STEP_CONFIG_KEYS)
        .mixin(parameters, PARAMETER_KEYS)
        .addIfEmpty('checkmarxExecuteScan', script.commonPipelineEnvironment.configuration.runStep?.get(stageName)?.checkmarxExecuteScan)
        .addIfEmpty('detectExecuteScan', script.commonPipelineEnvironment.configuration.runStep?.get(stageName)?.detectExecuteScan)
        .addIfEmpty('fortifyExecuteScan', script.commonPipelineEnvironment.configuration.runStep?.get(stageName)?.fortifyExecuteScan)
        .addIfEmpty('whitesourceExecuteScan', script.commonPipelineEnvironment.configuration.runStep?.get(stageName)?.whitesourceExecuteScan)
        .use()

    piperStageWrapper (script: script, stageName: stageName) {
        if (config.checkmarxExecuteScan) {
            securityScanMap['Checkmarx'] = {
                node(config.nodeLabel) {
                    try{
                        durationMeasure(script: script, measurementName: 'checkmarx_duration') {
                            checkmarxExecuteScan script: script
                        }
                    }finally{
                        deleteDir()
                    }
                }
            }
        }

        if (config.detectExecuteScan) {
            securityScanMap['Detect'] = {
                node(config.nodeLabel) {
                    try{
                        durationMeasure(script: script, measurementName: 'detect_duration') {
                            detectExecuteScan script: script
                        }
                    }finally{
                        deleteDir()
                    }
                }
            }
        }

        if (config.fortifyExecuteScan) {
            securityScanMap['Fortify'] = {
                node(config.nodeLabel) {
                    try{
                        durationMeasure(script: script, measurementName: 'fortify_duration') {
                            fortifyExecuteScan script: script
                        }
                    }finally{
                        deleteDir()
                    }
                }
            }
        }

        if (config.whitesourceExecuteScan) {
            securityScanMap['WhiteSource'] = {
                node(config.nodeLabel) {
                    try{
                        durationMeasure(script: script, measurementName: 'whitesource_duration') {
                            whitesourceExecuteScan script: script
                        }
                    }finally{
                        deleteDir()
                    }
                }
            }
        }

        if (securityScanMap.size() > 0) {
            parallel securityScanMap.plus([failFast: false])
        }
    }
}
