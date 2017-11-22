class commonPipelineEnvironment implements Serializable {
    private Map configProperties = [:]

    private String mtarFilePath
    private Map gitCoordinates

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
    def getMtarFilePath() {
        return mtarFilePath
    }
    void setMtarFilePath(mtarFilePath) {
        this.mtarFilePath = mtarFilePath
    }
    def getGitCoordinates() {
        return gitCoordinates
    }
    void setGitCoordinates(gitCoordinates) {
        this.gitCoordinates = gitCoordinates
    }
}
