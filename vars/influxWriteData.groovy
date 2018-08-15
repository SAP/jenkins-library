import com.sap.piper.ConfigurationLoader
import com.sap.piper.ConfigurationMerger
import com.sap.piper.JsonUtils

import groovy.transform.Field

@Field def STEP_NAME = 'influxWriteData'

@Field Set GENERAL_CONFIG_KEYS = []
@Field Set STEP_CONFIG_KEYS = [
    'influxServer',
    'influxPrefix'
]
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS.plus([
    'artifactVersion'
])

def call(Map parameters = [:]) {
    handlePipelineStepErrors (stepName: STEP_NAME, stepParameters: parameters, allowBuildFailure: true) {
        def script = parameters.script ?: [commonPipelineEnvironment: commonPipelineEnvironment]

        // load default & individual configuration
        Map configuration = ConfigurationHelper
            .loadStepDefaults(this)
            .mixinGeneralConfig(script.commonPipelineEnvironment, GENERAL_CONFIG_KEYS)
            .mixinStepConfig(script.commonPipelineEnvironment, STEP_CONFIG_KEYS)
            .mixinStageConfig(script.commonPipelineEnvironment, parameters.stageName?:env.STAGE_NAME, STEP_CONFIG_KEYS)
            .mixin([
                artifactVersion: commonPipelineEnvironment.getArtifactVersion()
            ])
            .mixin(parameters, PARAMETER_KEYS)
            .use()

        Map pipelineDataMap = [
            artifactVersion: commonPipelineEnvironment.getArtifactVersion()
        ]

        Map configuration = ConfigurationMerger.merge(script, STEP_NAME, parameters, parameterKeys, pipelineDataMap, stepConfigurationKeys)

        def artifactVersion = configuration.artifactVersion
        if (!artifactVersion)  {
            //this takes care that terminated builds due to milestone-locking do not cause an error
            echo "[${STEP_NAME}] no artifact version available -> exiting writeInflux without writing data"
            return
        }

        def influxServer = configuration.influxServer
        def influxPrefix = configuration.influxPrefix

        echo """[${STEP_NAME}]----------------------------------------------------------
Artifact version: ${artifactVersion}
Influx server: ${influxServer}
Influx prefix: ${influxPrefix}
InfluxDB data: ${script.commonPipelineEnvironment.getInfluxCustomData()}
InfluxDB data map: ${script.commonPipelineEnvironment.getInfluxCustomDataMap()}
[${STEP_NAME}]----------------------------------------------------------"""

        if (influxServer)
            step([$class: 'InfluxDbPublisher', selectedTarget: influxServer, customPrefix: influxPrefix, customData: script.commonPipelineEnvironment.getInfluxCustomData(), customDataMap: script.commonPipelineEnvironment.getInfluxCustomDataMap()])

        //write results into json file for archiving - also benefitial when no InfluxDB is available yet
        def jsonUtils = new JsonUtils()
        writeFile file: 'jenkins_data.json', text: jsonUtils.getPrettyJsonString(script.commonPipelineEnvironment.getInfluxCustomData())
        writeFile file: 'pipeline_data.json', text: jsonUtils.getPrettyJsonString(script.commonPipelineEnvironment.getInfluxCustomDataMap())
        archiveArtifacts artifacts: '*data.json', allowEmptyArchive: true
    }
}
