import com.sap.piper.ConfigurationLoader
import com.sap.piper.ConfigurationMerger
import com.sap.piper.JsonUtils

def call(Map parameters = [:]) {

    def stepName = 'influxWriteData'

    handlePipelineStepErrors (stepName: stepName, stepParameters: parameters, allowBuildFailure: true) {

        def script = parameters.script
        if (script == null)
            script = [commonPipelineEnvironment: commonPipelineEnvironment]

        prepareDefaultValues script: script

        final Map stepDefaults = ConfigurationLoader.defaultStepConfiguration(script, stepName)
        final Map stepConfiguration = ConfigurationLoader.stepConfiguration(script, stepName)
        final Map generalConfiguration = ConfigurationLoader.generalConfiguration(script)

        List parameterKeys = [
            'artifactVersion',
            'influxServer',
            'influxPrefix'
        ]
        Map pipelineDataMap = [
            artifactVersion: commonPipelineEnvironment.getArtifactVersion()
        ]
        List stepConfigurationKeys = [
            'influxServer',
            'influxPrefix'
        ]

        Map configuration = ConfigurationMerger.mergeWithPipelineData(parameters, parameterKeys, pipelineDataMap, stepConfiguration, stepConfigurationKeys, stepDefaults)

        def artifactVersion = configuration.artifactVersion
        if (artifactVersion == null)  {
            //this takes care that terminated builds due to milestone-locking do not cause an error
            echo "[${stepName}] no artifact version available -> exiting writeInflux without writing data"
            return
        }

        def influxServer = configuration.influxServer
        def influxPrefix = configuration.influxPrefix

        echo """[${stepName}]----------------------------------------------------------
Artifact version: ${artifactVersion}
Influx server: ${influxServer}
Influx prefix: ${influxPrefix}
InfluxDB data: ${script.commonPipelineEnvironment.getInfluxCustomData()}
InfluxDB data map: ${script.commonPipelineEnvironment.getInfluxCustomDataMap()}
[${stepName}]----------------------------------------------------------"""

        if (influxServer)
            step([$class: 'InfluxDbPublisher', selectedTarget: influxServer, customPrefix: influxPrefix, customData: script.commonPipelineEnvironment.getInfluxCustomData(), customDataMap: script.commonPipelineEnvironment.getInfluxCustomDataMap()])

        //write results into json file for archiving - also benefitial when no InfluxDB is available yet
        //write results into json file for archiving - also benefitial when no InfluxDB is available yet
        def jsonUtils = new JsonUtils()
        writeFile file: 'jenkins_data.json', text: jsonUtils.getPrettyJsonString(script.commonPipelineEnvironment.getInfluxCustomData())
        writeFile file: 'pipeline_data.json', text: jsonUtils.getPrettyJsonString(script.commonPipelineEnvironment.getInfluxCustomDataMap())
        archiveArtifacts artifacts: '*data.json', allowEmptyArchive: true
    }
}
