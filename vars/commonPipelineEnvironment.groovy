import com.sap.piper.ConfigurationLoader
import com.sap.piper.ConfigurationMerger
import com.sap.piper.DefaultValueCache
import com.sap.piper.analytics.InfluxData

class commonPipelineEnvironment implements Serializable {

    //stores properties for a pipeline which build an artifact and then bundles it into a container
    private Map appContainerProperties = [:]

    Map defaultConfiguration = [:]

    //
    // We forward to cpe declared on DefaultValueCache
    def methodMissing(String name, def args) {
        DefaultValueCache.commonPipelineEnvironment.invokeMethod(name, args)
    }

    def propertyMissing(def name) {
       DefaultValueCache.commonPipelineEnvironment[name]
    }

    def propertyMissing(def name, def value) {
       DefaultValueCache.commonPipelineEnvironment[name] = value
    }
    // End forwarding to DefaultValueCache
    //

    /*
     * Should only be used by tests
     */
    void setConfiguration(Map configuration) {
        DefaultValueCache.createInstance(DefaultValueCache.getInstance()?.getDefaultValues() ?: [:], configuration)
    }

    def getConfiguration() {
        DefaultValueCache.getInstance().getProjectConfig()
    }

    void setValue(String property, value) {
        DefaultValueCache.commonPipelineEnvironment.setValue(property, value)
    }

    def getValue(String property) {
        DefaultValueCache.commonPipelineEnvironment.getValue(property)
    }

    def reset() {
        appContainerProperties = [:]
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

    Map getStepConfiguration(stepName, stageName = env.STAGE_NAME, includeDefaults = true) {
        Map defaults = [:]
        if (includeDefaults) {
            defaults = DefaultValueCache.getInstance()?.getDefaultValues()?.general ?: [:]
            defaults = ConfigurationMerger.merge(ConfigurationLoader.defaultStepConfiguration(stepName), null, defaults)
            defaults = ConfigurationMerger.merge(ConfigurationLoader.defaultStageConfiguration(stageName), null, defaults)
        }
        Map projectConfig = DefaultValueCache.getInstance().getProjectConfig()
        Map config = ConfigurationMerger.merge(projectConfig.get('general') ?: [:], null, defaults)
        config = ConfigurationMerger.merge(projectConfig.get('steps')?.get(stepName) ?: [:], null, config)
        config = ConfigurationMerger.merge(projectConfig.get('stages')?.get(stageName) ?: [:], null, config)
        return config
    }
}
