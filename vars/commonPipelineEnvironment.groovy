class commonPipelineEnvironment implements Serializable {
    private Map configProperties = [:]

    //stores version of the artifact which is build during pipeline run
    def artifactVersion

    Map configuration = [:]
    Map defaultConfiguration = [:]

    //each Map in influxCustomDataMap represents a measurement in Influx. Additional measurements can be added as a new Map entry of influxCustomDataMap
    private Map influxCustomDataMap = [pipeline_data: [:]]
    //influxCustomData represents measurement jenkins_custom_data in Influx. Metrics can be written into this map
    private Map influxCustomData = [:]

    private String mtarFilePath

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
}
