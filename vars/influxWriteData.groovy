import com.sap.piper.JenkinsUtils

import static com.sap.piper.Prerequisites.checkScript

import com.sap.piper.GenerateDocumentation
import com.sap.piper.ConfigurationHelper
import com.sap.piper.JsonUtils
import com.sap.piper.Utils
import com.sap.piper.analytics.InfluxData

import groovy.transform.Field

@Field def STEP_NAME = getClass().getName()

@Field Set GENERAL_CONFIG_KEYS = []
@Field Set STEP_CONFIG_KEYS = GENERAL_CONFIG_KEYS.plus([
    /**
     * Defines the version of the current artifact. Defaults to `commonPipelineEnvironment.getArtifactVersion()`
     */
    'artifactVersion',
    /**
     * Defines custom data (map of key-value pairs) to be written to Influx into measurement `jenkins_custom_data`. Defaults to `commonPipelineEnvironment.getInfluxCustomData()`
     */
    'customData',
    /**
     * Defines tags (map of key-value pairs) to be written to Influx into measurement `jenkins_custom_data`. Defaults to `commonPipelineEnvironment.getInfluxCustomDataTags()`
     */
    'customDataTags',
    /**
     * Defines a map of measurement names containing custom data (map of key-value pairs) to be written to Influx. Defaults to `commonPipelineEnvironment.getInfluxCustomDataMap()`
     */
    'customDataMap',
    /**
     * Defines a map of measurement names containing tags (map of key-value pairs) to be written to Influx. Defaults to `commonPipelineEnvironment.getInfluxCustomDataTags()`
     */
    'customDataMapTags',
    /**
     * Defines the name of the Influx server as configured in Jenkins global configuration.
     */
    'influxServer',
    /**
     * Defines a custom prefix.
     * For example in multi branch pipelines, where every build is named after the branch built and thus you have different builds called 'master' that report different metrics.
     */
    'influxPrefix',
    'sonarTokenCredentialsId',
    /**
     * Defines if a dedicated node/executor should be created in the pipeline run.
     * This is especially relevant when running the step in a declarative `POST` stage where by default no executor is available.
     */
    'wrapInNode'
])
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS

/**
 * Since your Continuous Delivery Pipeline in Jenkins provides your productive development and delivery infrastructure you should monitor the pipeline to ensure it runs as expected. How to setup this monitoring is described in the following.
 *
 * You basically need three components:
 *
 * - The [InfluxDB Jenkins plugin](https://wiki.jenkins-ci.org/display/JENKINS/InfluxDB+Plugin) which allows you to send build metrics to InfluxDB servers
 * - The [InfluxDB](https://www.influxdata.com/time-series-platform/influxdb/) to store this data (Docker available)
 * - A [Grafana](http://grafana.org/) dashboard to visualize the data stored in InfluxDB (Docker available)
 *
 * !!! note "no InfluxDB available?"
 *     If you don't have an InfluxDB available yet this step will still provide you some benefit.
 *
 *     It will create following files for you and archive them into your build:
 *
 *     * `jenkins_data.json`: This file gives you build-specific information, like e.g. build result, stage where the build failed
 *     * `influx_data.json`: This file gives you detailed information about your pipeline, e.g. stage durations, steps executed, ...
 */
@GenerateDocumentation
void call(Map parameters = [:]) {
    handlePipelineStepErrors (stepName: STEP_NAME, stepParameters: parameters, allowBuildFailure: true) {

        def script = checkScript(this, parameters) ?: this
        def jenkinsUtils = parameters.jenkinsUtilsStub ?: new JenkinsUtils()
        String stageName = parameters.stageName ?: env.STAGE_NAME

        // load default & individual configuration
        Map config = ConfigurationHelper.newInstance(this)
            .loadStepDefaults([:], stageName)
            .mixinGeneralConfig(script.commonPipelineEnvironment, GENERAL_CONFIG_KEYS)
            .mixinStepConfig(script.commonPipelineEnvironment, STEP_CONFIG_KEYS)
            .mixinStageConfig(script.commonPipelineEnvironment, stageName, STEP_CONFIG_KEYS)
            .mixin([
                artifactVersion: script.commonPipelineEnvironment.getArtifactVersion(),
                influxPrefix: script.commonPipelineEnvironment.getGithubOrg() && script.commonPipelineEnvironment.getGithubRepo()
                    ? "${script.commonPipelineEnvironment.getGithubOrg()}_${script.commonPipelineEnvironment.getGithubRepo()}"
                    : null
            ])
            .mixin(parameters, PARAMETER_KEYS)
            .addIfNull('customData', InfluxData.getInstance().getFields().jenkins_custom_data)
            .addIfNull('customDataTags', InfluxData.getInstance().getTags().jenkins_custom_data)
            .addIfNull('customDataMap', InfluxData.getInstance().getFields().findAll({ it.key != 'jenkins_custom_data' }))
            .addIfNull('customDataMapTags', InfluxData.getInstance().getTags().findAll({ it.key != 'jenkins_custom_data' }))
            .use()

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
                    writeToInflux(config, jenkinsUtils, script)
                }finally{
                    deleteDir()
                }
            }
        } else {
            writeToInflux(config, jenkinsUtils, script)
        }
    }
}

private void writeToInflux(config, JenkinsUtils jenkinsUtils, script){
    if (config.influxServer) {

        def influxPluginVersion = jenkinsUtils.getPluginVersion('influxdb')

        try {
            def credentialList = []
            def influxParams = [
                selectedTarget: config.influxServer,
                customPrefix: config.influxPrefix,
                customData: config.customData.size()>0 ? config.customData : null,
                customDataTags: config.customDataTags.size()>0 ? config.customDataTags : null,
                customDataMap: config.customDataMap.size()>0 ? config.customDataMap : null,
                customDataMapTags: config.customDataMapTags.size()>0 ? config.customDataMapTags : null
            ]
            if(config.sonarTokenCredentialsId){
                credentialList.add(string(
                    credentialsId: config.sonarTokenCredentialsId,
                    variable: 'SONAR_AUTH_TOKEN'
                ))
            }
            withCredentials(credentialList){
                if (!influxPluginVersion || influxPluginVersion.startsWith('1.')) {
                    influxParams['$class'] = 'InfluxDbPublisher'
                    step(influxParams)
                } else {
                    influxDbPublisher(influxParams)
                }
            }

        } catch (NullPointerException e){
            if(!e.getMessage()){
                //TODO: catch NPEs as long as https://issues.jenkins-ci.org/browse/JENKINS-55594 is not fixed & released
                error "[$STEP_NAME] NullPointerException occurred, is the correct target defined?"
            }
            throw e
        }
    }

    //write results into json file for archiving - also beneficial when no InfluxDB is available yet
    def jsonUtils = new JsonUtils()
    writeFile file: 'jenkins_data.json', text: jsonUtils.groovyObjectToPrettyJsonString(config.customData)
    writeFile file: 'influx_data.json', text: jsonUtils.groovyObjectToPrettyJsonString(config.customDataMap)
    writeFile file: 'jenkins_data_tags.json', text: jsonUtils.groovyObjectToPrettyJsonString(config.customDataTags)
    writeFile file: 'influx_data_tags.json', text: jsonUtils.groovyObjectToPrettyJsonString(config.customDataMapTags)
    archiveArtifacts artifacts: '*data*.json', allowEmptyArchive: true
}
