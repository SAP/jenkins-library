class commonPipelineEnvironment implements Serializable {
    Map configProperties = [:]

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
    Map influxCustomDataMapTags = [pipeline_data: [:]]
    //influxCustomData represents measurement jenkins_custom_data in Influx. Metrics can be written into this map
    private Map influxCustomData = [:]
    //influxCustomDataTags represents tags in Influx. Tags are required in Influx for easier querying data
    Map influxCustomDataTags = [:]

    String mtarFilePath

    String transportRequestId
    String changeDocumentId

    def reset() {
        appContainerProperties = [:]
        artifactVersion = null

        configProperties = [:]
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

    def setConfigProperty(property, value) {
        configProperties[property] = value
    }

    def getConfigProperty(property) {
        if (configProperties[property] != null)
            return configProperties[property].trim()
        else
            return configProperties[property]
    }

    def getInfluxCustomData() {
        return influxCustomData
    }

    def getInfluxCustomDataMap() {
        return influxCustomDataMap
    }

    def setInfluxCustomDataTag(tag, value) {
        influxCustomDataTags[tag] = value
    }
    def setInfluxCustomDataMapTag(measurement, tag, value) {
        if (!influxCustomDataMapTags[measurement]) {
            influxCustomDataMapTags[measurement] = [:]
        }
        influxCustomDataMapTags[measurement][tag] = value
    }

    def setInfluxStepData (dataKey, value) {
        influxCustomDataMap.step_data[dataKey] = value
    }
    def getInfluxStepData (dataKey) {
        return influxCustomDataMap.step_data[dataKey]
    }

    def setPipelineMeasurement (measurementName, value) {
        influxCustomDataMap.pipeline_data[measurementName] = value
    }

    def getPipelineMeasurement (measurementName) {
        return influxCustomDataMap.pipeline_data[measurementName]
    }
}
