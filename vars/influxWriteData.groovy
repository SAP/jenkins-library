import static com.sap.piper.Prerequisites.checkScript

import com.sap.piper.ConfigurationHelper
import com.sap.piper.ConfigurationLoader
import com.sap.piper.ConfigurationMerger
import com.sap.piper.JsonUtils
import com.sap.piper.Utils

import groovy.transform.Field

@Field def STEP_NAME = getClass().getName()

@Field Set GENERAL_CONFIG_KEYS = []
@Field Set STEP_CONFIG_KEYS = [
    'influxServer',
    'influxPrefix',
    'wrapInNode'
]
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS.plus([
    'artifactVersion'
])

void call(Map parameters = [:]) {
    handlePipelineStepErrors (stepName: STEP_NAME, stepParameters: parameters, allowBuildFailure: true) {

        def script = checkScript(this, parameters)
        if (script == null)
            script = this

        // load default & individual configuration
        Map config = ConfigurationHelper.newInstance(this)
            .loadStepDefaults()
            .mixinGeneralConfig(script.commonPipelineEnvironment, GENERAL_CONFIG_KEYS)
            .mixinStepConfig(script.commonPipelineEnvironment, STEP_CONFIG_KEYS)
            .mixinStageConfig(script.commonPipelineEnvironment, parameters.stageName?:env.STAGE_NAME, STEP_CONFIG_KEYS)
            .mixin([
                artifactVersion: script.commonPipelineEnvironment.getArtifactVersion()
            ])
            .mixin(parameters, PARAMETER_KEYS)
            .use()

        new Utils().pushToSWA([step: STEP_NAME,
                                stepParam1: parameters?.script == null], config)

        if (!config.artifactVersion)  {
            //this takes care that terminated builds due to milestone-locking do not cause an error
            echo "[${STEP_NAME}] no artifact version available -> exiting writeInflux without writing data"
            return
        }

        echo """[${STEP_NAME}]----------------------------------------------------------
Artifact version: ${config.artifactVersion}
Influx server: ${config.influxServer}
Influx prefix: ${config.influxPrefix}
InfluxDB data: ${script.commonPipelineEnvironment.getInfluxCustomData()}
InfluxDB data map: ${script.commonPipelineEnvironment.getInfluxCustomDataMap()}
[${STEP_NAME}]----------------------------------------------------------"""

        if(config.wrapInNode){
            node(''){
                try{
                    writeToInflux(config, script)
                }finally{
                    deleteDir()
                }
            }
        } else {
            writeToInflux(config, script)
        }
    }
}

private void writeToInflux(config, script){
    if (config.influxServer) {
        step([
            $class: 'InfluxDbPublisher',
            selectedTarget: config.influxServer,
            customPrefix: config.influxPrefix,
            customData: script.commonPipelineEnvironment.getInfluxCustomData(),
            customDataMap: script.commonPipelineEnvironment.getInfluxCustomDataMap()
        ])
    }

    //write results into json file for archiving - also benefitial when no InfluxDB is available yet
    def jsonUtils = new JsonUtils()
    writeFile file: 'jenkins_data.json', text: jsonUtils.getPrettyJsonString(script.commonPipelineEnvironment.getInfluxCustomData())
    writeFile file: 'pipeline_data.json', text: jsonUtils.getPrettyJsonString(script.commonPipelineEnvironment.getInfluxCustomDataMap())
    archiveArtifacts artifacts: '*data.json', allowEmptyArchive: true

}
