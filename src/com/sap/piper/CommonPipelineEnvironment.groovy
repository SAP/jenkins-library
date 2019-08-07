package com.sap.piper;

import com.sap.piper.analytics.InfluxData

public class CommonPipelineEnvironment {

    private static CommonPipelineEnvironment INSTANCE = new CommonPipelineEnvironment()

    static CommonPipelineEnvironment getInstance() {
        INSTANCE
    }

    Map defaultConfiguration = [:]

    // The project config
    Map configuration = [:]

    private Map valueMap = [:]

    //stores properties for a pipeline which build an artifact and then bundles it into a container
    private Map appContainerProperties = [:]

    //stores version of the artifact which is build during pipeline run
    def artifactVersion

    //Stores the current buildResult
    String buildResult = 'SUCCESS'

    //stores the gitCommitId as well as additional git information for the build during pipeline run
    String gitCommitId
    String gitCommitMessage
    String gitSshUrl
    String gitHttpsUrl
    String gitBranch

    //GiutHub specific information
    String githubOrg
    String githubRepo

    String mtarFilePath

    String changeDocumentId

    void setValue(String property, value) {
        valueMap[property] = value
    }

    def getValue(String property) {
        return valueMap.get(property)
    }

    def setAppContainerProperty(property, value) {
        appContainerProperties[property] = value
    }

    def getAppContainerProperty(property) {
        return appContainerProperties[property]
    }

    // goes into measurement jenkins_custom_data
    def setInfluxCustomDataEntry(key, value) {
        InfluxData.addField('jenkins_custom_data', key, value)
    }
    // goes into measurement jenkins_custom_data
    @Deprecated // not used in library
    def getInfluxCustomData() {
        return InfluxData.getInstance().getFields().jenkins_custom_data
    }

    // goes into measurement jenkins_custom_data
    def setInfluxCustomDataTagsEntry(key, value) {
        InfluxData.addTag('jenkins_custom_data', key, value)
    }
    // goes into measurement jenkins_custom_data
    @Deprecated // not used in library
    def getInfluxCustomDataTags() {
        return InfluxData.getInstance().getTags().jenkins_custom_data
    }

    void setInfluxCustomDataMapEntry(measurement, field, value) {
        InfluxData.addField(measurement, field, value)
    }
    @Deprecated // not used in library
    def getInfluxCustomDataMap() {
        return InfluxData.getInstance().getFields()
    }

    def setInfluxCustomDataMapTagsEntry(measurement, tag, value) {
        InfluxData.addTag(measurement, tag, value)
    }
    @Deprecated // not used in library
    def getInfluxCustomDataMapTags() {
        return InfluxData.getInstance().getTags()
    }

    @Deprecated // not used in library
    def setInfluxStepData(key, value) {
        InfluxData.addField('step_data', key, value)
    }
    @Deprecated // not used in library
    def getInfluxStepData(key) {
        return InfluxData.getInstance().getFields()['step_data'][key]
    }

    @Deprecated // not used in library
    def setInfluxPipelineData(key, value) {
        InfluxData.addField('pipeline_data', key, value)
    }
    @Deprecated // not used in library
    def setPipelineMeasurement(key, value){
        setInfluxPipelineData(key, value)
    }
    @Deprecated // not used in library
    def getPipelineMeasurement(key) {
        return InfluxData.getInstance().getFields()['pipeline_data'][key]
    }

    def reset() {
        appContainerProperties = [:]
        configuration = [:]
        artifactVersion = null

        gitCommitId = null
        gitCommitMessage = null
        gitSshUrl = null
        gitHttpsUrl = null
        gitBranch = null

        githubOrg = null
        githubRepo = null

        mtarFilePath = null
        valueMap = [:]

        changeDocumentId = null

        InfluxData.reset()
    }

    Map getStepConfiguration(stepName, stageName = env.STAGE_NAME, includeDefaults = true) {
        Map defaults = [:]
        if (includeDefaults) {
            defaults = DefaultValueCache.getInstance()?.getDefaultValues()?.general ?: [:]
            defaults = ConfigurationMerger.merge(ConfigurationLoader.defaultStepConfiguration([commonPipelineEnvironment: this], stepName), null, defaults)
            defaults = ConfigurationMerger.merge(ConfigurationLoader.defaultStageConfiguration([commonPipelineEnvironment: this], stageName), null, defaults)
        }
        Map config = ConfigurationMerger.merge(configuration.get('general') ?: [:], null, defaults)
        config = ConfigurationMerger.merge(configuration.get('steps')?.get(stepName) ?: [:], null, config)
        config = ConfigurationMerger.merge(configuration.get('stages')?.get(stageName) ?: [:], null, config)
        return config
    }
}
