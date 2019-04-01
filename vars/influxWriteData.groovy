import static com.sap.piper.Prerequisites.checkScript

import com.sap.piper.ConfigurationHelper
import com.sap.piper.JsonUtils
import com.sap.piper.Utils
import com.sap.piper.analytics.InfluxData

import groovy.transform.Field

@Field def STEP_NAME = getClass().getName()

@Field Set GENERAL_CONFIG_KEYS = []
@Field Set STEP_CONFIG_KEYS = GENERAL_CONFIG_KEYS.plus([
    'artifactVersion',
    'customData',
    'customDataTags',
    'customDataMap',
    'customDataMapTags',
    'influxServer',
    'influxPrefix',
    'wrapInNode'
])
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS

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
                artifactVersion: script.commonPipelineEnvironment.getArtifactVersion(),
                influxPrefix: script.commonPipelineEnvironment.getGithubOrg() && script.commonPipelineEnvironment.getGithubRepo()
                    ? "${script.commonPipelineEnvironment.getGithubOrg()}_${script.commonPipelineEnvironment.getGithubRepo()}"
                    : null
            ])
            .mixin(parameters, PARAMETER_KEYS)
            .addIfNull('customData', InfluxData.getInstance().getFields().jenkins_custom_data)
            .addIfNull('customDataTags', InfluxData.getInstance().getTags().jenkins_custom_data)
            .addIfNull('customDataMap', InfluxData.getInstance().getFields().dropWhile{ it.key == 'jenkins_custom_data' })
            .addIfNull('customDataMapTags', InfluxData.getInstance().getTags().dropWhile{ it.key == 'jenkins_custom_data' })
            .use()

        new Utils().pushToSWA([
            step: STEP_NAME,
            stepParamKey1: 'scriptMissing',
            stepParam1: parameters?.script == null
        ], config)

        if (!config.artifactVersion)  {
            //this takes care that terminated builds due to milestone-locking do not cause an error
            echo "[${STEP_NAME}] no artifact version available -> exiting writeInflux without writing data"
            return
        }

        echo """[${STEP_NAME}]----------------------------------------------------------
Artifact version: ${config.artifactVersion}
Influx server: ${config.influxServer}
Influx prefix: ${config.influxPrefix}
InfluxDB data: ${config.customData}
InfluxDB data tags: ${config.customDataTags}
InfluxDB data map: ${config.customDataMap}
InfluxDB data map tags: ${config.customDataMapTags}
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
        try {
            step([
                $class: 'InfluxDbPublisher',
                selectedTarget: config.influxServer,
                customPrefix: config.influxPrefix,
                customData: config.customData.size()>0 ? config.customData : null,
                customDataTags: config.customDataTags.size()>0 ? config.customDataTags : null,
                customDataMap: config.customDataMap.size()>0 ? config.customDataMap : null,
                customDataMapTags: config.customDataMapTags.size()>0 ? config.customDataMapTags : null
            ])
        } catch (NullPointerException e){
            if(!e.getMessage()){
                //TODO: catch NPEs as long as https://issues.jenkins-ci.org/browse/JENKINS-55594 is not fixed & released
                error "[$STEP_NAME] NullPointerException occured, is the correct target defined?"
            }
            throw e
        }
    }

    //write results into json file for archiving - also benefitial when no InfluxDB is available yet
    def jsonUtils = new JsonUtils()
    writeFile file: 'jenkins_data.json', text: jsonUtils.getPrettyJsonString(config.customData)
    writeFile file: 'influx_data.json', text: jsonUtils.getPrettyJsonString(config.customDataMap)
    writeFile file: 'jenkins_data_tags.json', text: jsonUtils.getPrettyJsonString(config.customDataTags)
    writeFile file: 'influx_data_tags.json', text: jsonUtils.getPrettyJsonString(config.customDataMapTags)
    archiveArtifacts artifacts: '*data.json', allowEmptyArchive: true
}
