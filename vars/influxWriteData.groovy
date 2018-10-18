import com.sap.piper.ConfigurationHelper
import com.sap.piper.ConfigurationLoader
import com.sap.piper.ConfigurationMerger
import com.sap.piper.JsonUtils
import com.sap.piper.Utils

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
        def script = parameters.script
        if (script == null)
             script = [commonPipelineEnvironment: commonPipelineEnvironment]

        // load default & individual configuration
        Map configuration = ConfigurationHelper.newInstance(this)
            .loadStepDefaults()
            .mixinGeneralConfig(script.commonPipelineEnvironment, GENERAL_CONFIG_KEYS)
            .mixinStepConfig(script.commonPipelineEnvironment, STEP_CONFIG_KEYS)
            .mixinStageConfig(script.commonPipelineEnvironment, parameters.stageName?:env.STAGE_NAME, STEP_CONFIG_KEYS)
            .mixin([
                artifactVersion: commonPipelineEnvironment.getArtifactVersion()
            ])
            .mixin(parameters, PARAMETER_KEYS)
            .use()

        new Utils().pushToSWA([step: STEP_NAME], configuration)

        if (!configuration.artifactVersion)  {
            //this takes care that terminated builds due to milestone-locking do not cause an error
            echo "[${STEP_NAME}] no artifact version available -> exiting writeInflux without writing data"
            return
        }

        echo """[${STEP_NAME}]----------------------------------------------------------
Artifact version: ${configuration.artifactVersion}
Influx server: ${configuration.influxServer}
Influx prefix: ${configuration.influxPrefix}
InfluxDB data: ${script.commonPipelineEnvironment.getInfluxCustomData()}
InfluxDB data map: ${script.commonPipelineEnvironment.getInfluxCustomDataMap()}
[${STEP_NAME}]----------------------------------------------------------"""

        if (configuration.influxServer)
            step([$class: 'InfluxDbPublisher', selectedTarget: configuration.influxServer, customPrefix: configuration.influxPrefix, customData: script.commonPipelineEnvironment.getInfluxCustomData(), customDataMap: script.commonPipelineEnvironment.getInfluxCustomDataMap()])

        //write results into json file for archiving - also benefitial when no InfluxDB is available yet
        def jsonUtils = new JsonUtils()
        writeFile file: 'jenkins_data.json', text: jsonUtils.getPrettyJsonString(script.commonPipelineEnvironment.getInfluxCustomData())
        writeFile file: 'pipeline_data.json', text: jsonUtils.getPrettyJsonString(script.commonPipelineEnvironment.getInfluxCustomDataMap())
        archiveArtifacts artifacts: '*data.json', allowEmptyArchive: true
    }
}
