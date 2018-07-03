class commonPipelineEnvironment implements Serializable {
    private Map configProperties = [:]

    //stores version of the artifact which is build during pipeline run
    def artifactVersion

    //stores the gitCommitId as well as additional git information for the build during pipeline run
    private String gitCommitId
    private String gitSshUrl

    //stores properties for a pipeline which build an artifact and then bundles it into a container
    private Map appContainerProperties = [:]

    Map configuration = [:]
    Map defaultConfiguration = [:]

    //each Map in influxCustomDataMap represents a measurement in Influx. Additional measurements can be added as a new Map entry of influxCustomDataMap
    private Map influxCustomDataMap = [pipeline_data: [:], step_data: [:]]
    //influxCustomData represents measurement jenkins_custom_data in Influx. Metrics can be written into this map
    private Map influxCustomData = [:]

    private String mtarFilePath

    private String transportRequestId

    def reset() {
        appContainerProperties = [:]
        artifactVersion = null

        configProperties = [:]
        configuration = [:]

        gitCommitId = null
        gitSshUrl = null

        influxCustomData = [:]
        influxCustomDataMap = [pipeline_data: [:], step_data: [:]]

        mtarFilePath = null
        transportRequestId = null
    }

    def setAppContainerProperty(property, value) {
        appContainerProperties[property] = value
    }

    def getAppContainerProperty(property) {
        return appContainerProperties[property]
    }

    def setArtifactVersion(version) {
        artifactVersion = version
    }

    def getArtifactVersion() {
        return artifactVersion
    }

    def setConfigProperties(map) {
        configProperties = map
    }

    def getConfigProperties() {
        return configProperties
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

    def setGitCommitId(commitId) {
        gitCommitId = commitId
    }

    def getGitCommitId() {
        return gitCommitId
    }

    def setGitSshUrl(url) {
        gitSshUrl = url
    }

    def getGitSshUrl() {
        return gitSshUrl
    }

    def getInfluxCustomData() {
        return influxCustomData
    }

    def getInfluxCustomDataMap() {
        return influxCustomDataMap
    }

    def getMtarFilePath() {
        return mtarFilePath
    }

    void setMtarFilePath(mtarFilePath) {
        this.mtarFilePath = mtarFilePath
    }

    def setPipelineMeasurement (measurementName, value) {
        influxCustomDataMap.pipeline_data[measurementName] = value
    }

    def getPipelineMeasurement (measurementName) {
        return influxCustomDataMap.pipeline_data[measurementName]
    }

    def setTransportRequestId(transportRequestId) {
        this.transportRequestId = transportRequestId
    }

    def getTransportRequestId() {
        return this.transportRequestId
    }
}
