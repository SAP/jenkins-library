import com.sap.piper.GenerateDocumentation
import com.sap.piper.ConfigurationLoader
import com.sap.piper.ConfigurationMerger

@Field String STEP_NAME = getClass().getName()

@Field Set GENERAL_CONFIG_KEYS = []

@Field Set STEP_CONFIG_KEYS = []

@Field Set PARAMETER_KEYS = []

/**
 * Provides project specific settings.
 */
@GenerateDocumentation
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

    //each Map in influxCustomDataMap represents a measurement in Influx. Additional measurements can be added as a new Map entry of influxCustomDataMap
    private Map influxCustomDataMap = [pipeline_data: [:], step_data: [:]]
    //each Map in influxCustomDataMapTags represents tags for certain measurement in Influx. Tags are required in Influx for easier querying data
    private Map influxCustomDataMapTags = [pipeline_data: [:]]
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
        influxCustomDataMap = [pipeline_data: [:], step_data: [:]]
        influxCustomDataMapTags = [pipeline_data: [:]]

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

    // goes into measurement jenkins_data
    def setInfluxCustomDataEntry(field, value) {
        influxCustomData[field] = value
    }
    // goes into measurement jenkins_data
    /**
     * Returns the Influx custom data which can be collected during pipeline run.
     *
     * @return A `Map` containing the data collected.
     */
    def getInfluxCustomData() {
        return influxCustomData
    }

    // goes into measurement jenkins_data
    def setInfluxCustomDataTagsEntry(tag, value) {
        influxCustomDataTags[tag] = value
    }

    // goes into measurement jenkins_data
    def getInfluxCustomDataTags() {
        return influxCustomDataTags
    }

    void setInfluxCustomDataMapEntry(measurement, field, value) {
        if (!influxCustomDataMap[measurement]) {
            influxCustomDataMap[measurement] = [:]
        }
        influxCustomDataMap[measurement][field] = value
    }

    /**
     * Returns the Influx custom data map which can be collected during pipeline run.
     * It is used for example by step [`influxWriteData`](../steps/influxWriteData.md).
     * The data map is a map of maps, like `[pipeline_data: [:], my_measurement: [:]]`
     * Each map inside the map represents a dedicated measurement in the InfluxDB.
     *
     * @return A `Map` containing a `Map`s with data collected.
     */
    def getInfluxCustomDataMap() {
        return influxCustomDataMap
    }

    def setInfluxCustomDataMapTagsEntry(measurement, tag, value) {
        if (!influxCustomDataMapTags[measurement]) {
            influxCustomDataMapTags[measurement] = [:]
        }
        influxCustomDataMapTags[measurement][tag] = value
    }
    def getInfluxCustomDataMapTags() {
        return influxCustomDataMapTags
    }

    def setInfluxStepData(key, value) {
        setInfluxCustomDataMapEntry('step_data', key, value)
    }
    def getInfluxStepData(key) {
        return influxCustomDataMap.step_data[key]
    }

    def setInfluxPipelineData(key, value) {
        setInfluxCustomDataMapEntry('pipeline_data', key, value)
    }

    /**
     * **This is an internal function!**
     * Sets the value of a specific pipeline measurement.
     * Please use the step [`durationMeasure`](../steps/durationMeasure.md) in a pipeline, instead.
     */
    def setPipelineMeasurement(key, value){
        setInfluxPipelineData(key, value)
    }

    /**
     * Returns the value of a specific pipeline measurement.
     * The measurements are collected with step [`durationMeasure`](../steps/durationMeasure.md)
     *
     * @return The value of the measurement
     */
    def getPipelineMeasurement(key) {
        return influxCustomDataMap.pipeline_data[key]
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
