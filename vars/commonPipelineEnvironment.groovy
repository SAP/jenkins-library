import com.sap.piper.ConfigurationLoader
import com.sap.piper.ConfigurationMerger
import com.sap.piper.analytics.InfluxData

class commonPipelineEnvironment implements Serializable {

    //stores version of the artifact which is build during pipeline run
    def artifactVersion

    //Stores the current buildResult
    String buildResult = 'SUCCESS'

    //stores the gitCommitId as well as additional git information for the build during pipeline run
    String gitCommitId
    String gitSshUrl
    String gitHttpsUrl
    String gitBranch

    //GiutHub specific information
    String githubOrg
    String githubRepo

    //stores properties for a pipeline which build an artifact and then bundles it into a container
    private Map appContainerProperties = [:]

    Map configuration = [:]
    Map defaultConfiguration = [:]

    //influxCustomData represents measurement jenkins_custom_data in Influx. Metrics can be written into this map
    private Map influxCustomData = [:]
    //influxCustomDataTags represents tags in Influx. Tags are required in Influx for easier querying data
    private Map influxCustomDataTags = [:]

    String mtarFilePath

    String transportRequestId
    String changeDocumentId

    def reset() {
        appContainerProperties = [:]
        artifactVersion = null

        configuration = [:]

        gitCommitId = null
        gitSshUrl = null
        gitHttpsUrl = null
        gitBranch = null

        githubOrg = null
        githubRepo = null

        influxCustomData = [:]
        influxCustomDataTags = [:]

        mtarFilePath = null

        transportRequestId = null
        changeDocumentId = null
    }

    def setAppContainerProperty(property, value) {
        appContainerProperties[property] = value
    }

    def getAppContainerProperty(property) {
        return appContainerProperties[property]
    }

    // goes into measurement jenkins_custom_data
    def setInfluxCustomDataEntry(field, value) {
        influxCustomData[field] = value
    }
    // goes into measurement jenkins_custom_data
    def getInfluxCustomData() {
        return influxCustomData
    }

    // goes into measurement jenkins_custom_data
    def setInfluxCustomDataTagsEntry(tag, value) {
        influxCustomDataTags[tag] = value
    }

    // goes into measurement jenkins_custom_data
    def getInfluxCustomDataTags() {
        return influxCustomDataTags
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
            defaults = ConfigurationLoader.defaultGeneralConfiguration()
            defaults = ConfigurationMerger.merge(ConfigurationLoader.defaultStepConfiguration(null, stepName), null, defaults)
            defaults = ConfigurationMerger.merge(ConfigurationLoader.defaultStageConfiguration(null, stageName), null, defaults)
        }
        Map config = ConfigurationMerger.merge(configuration.get('general') ?: [:], null, defaults)
        config = ConfigurationMerger.merge(configuration.get('steps')?.get(stepName) ?: [:], null, config)
        config = ConfigurationMerger.merge(configuration.get('stages')?.get(stageName) ?: [:], null, config)
        return config
    }
}
