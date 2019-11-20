import com.sap.piper.CommonPipelineEnvironment
import com.sap.piper.ConfigurationLoader
import com.sap.piper.ConfigurationMerger
import com.sap.piper.analytics.InfluxData

class commonPipelineEnvironment implements Serializable {

    //
    // Instances of this step does not keep any state. Everything is redirected to
    // the singleton CommonPipelineEnvironment. Since all instances of this step
    // share the same state through the singleton CommonPipelineEnvironment all
    // instances can be used in the same way.
    //
    // [Q] Why not simplify this using reflection, e.g. methodMissing/invokeMethod and
    //     similar? Wouln't this be simpler e.g. wrt. adding new properties? With the
    //     approach here we have to add the new property here and in the singleton class.
    //     And in general this look like boiler plate code ...
    // [A] Does not work with Jenkins since a security manager prohibits this.
    //

    void setArtifactVersion(String artifactVersion) {
        CommonPipelineEnvironment.getInstance().artifactVersion = artifactVersion
    }
    String getArtifactVersion() {
        CommonPipelineEnvironment.getInstance().artifactVersion
    }

    void setBuildResult(String buildResult) {
        CommonPipelineEnvironment.getInstance().buildResult = buildResult
    }
    String getBuildResult() {
        CommonPipelineEnvironment.getInstance().buildResult
    }

    void setGitCommitId(String gitCommitId) {
        CommonPipelineEnvironment.getInstance().gitCommitId = gitCommitId
    }
    String getGitCommitId() {
        CommonPipelineEnvironment.getInstance().gitCommitId
    }

    void setGitCommitMessage(String gitCommitMessage) {
        CommonPipelineEnvironment.getInstance().gitCommitMessage = gitCommitMessage
    }
    String getGitCommitMessage() {
        CommonPipelineEnvironment.getInstance().gitCommitMessage
    }

    void setGitSshUrl(String gitSshUrl) {
        CommonPipelineEnvironment.getInstance().gitSshUrl = gitSshUrl
    }
    String getGitSshUrl() {
        CommonPipelineEnvironment.getInstance().gitSshUrl
    }

    void setGitHttpsUrl(String gitHttpsUrl) {
        CommonPipelineEnvironment.getInstance().gitHttpsUrl = gitHttpsUrl
    }
    String getGitHttpsUrl() {
        CommonPipelineEnvironment.getInstance().gitHttpsUrl
    }

    void setGitBranch(String gitBranch) {
        CommonPipelineEnvironment.getInstance().gitBranch = gitBranch
    }
    String getGitBranch() {
        CommonPipelineEnvironment.getInstance().gitBranch
    }

    void setXsDeploymentId(String xsDeploymentId) {
        CommonPipelineEnvironment.getInstance().xsDeploymentId = xsDeploymentId
    }
    String getXsDeploymentId() {
        CommonPipelineEnvironment.getInstance().xsDeploymentId
    }

    void setGithubOrg(String githubOrg) {
        CommonPipelineEnvironment.getInstance().githubOrg = githubOrg
    }
    String getGithubOrg() {
        CommonPipelineEnvironment.getInstance().githubOrg
    }


    void setGithubRepo(String githubRepo) {
        CommonPipelineEnvironment.getInstance().githubRepo = githubRepo
    }
    String getGithubRepo() {
        CommonPipelineEnvironment.getInstance().githubRepo
    }

    Map getConfiguration() {
        CommonPipelineEnvironment.getInstance().configuration
    }
    void setConfiguration(Map configuration) {
        CommonPipelineEnvironment.getInstance().configuration = configuration
    }

    Map getDefaultConfiguration() {
        CommonPipelineEnvironment.getInstance().defaultConfiguration
    }
    void setDefaultConfiguration(Map defaultConfiguration) {
        CommonPipelineEnvironment.getInstance().defaultConfiguration = defaultConfiguration
    }

    String getMtarFilePath() {
        CommonPipelineEnvironment.getInstance().mtarFilePath
    }
    void setMtarFilePath(String mtarFilePath) {
        CommonPipelineEnvironment.getInstance().mtarFilePath = mtarFilePath
    }

    Map getValueMap() {
        CommonPipelineEnvironment.getInstance().valueMap
    }
    void setValueMap(Map valueMap) {
        CommonPipelineEnvironment.getInstance().valueMap = valueMap
    }

    void setValue(String property, value) {
        valueMap[property] = value
    }

    def getValue(String property) {
        return valueMap.get(property)
    }

    String getChangeDocumentId() {
        CommonPipelineEnvironment.getInstance().changeDocumentId
    }
    void setChangeDocumentId(String changeDocumentId) {
        CommonPipelineEnvironment.getInstance().changeDocumentId = changeDocumentId
    }

    def reset() {
        CommonPipelineEnvironment.getInstance().reset()
    }

    Map getAppContainerProperties() {
        CommonPipelineEnvironment.getInstance().appContainerProperties
    }
    void setAppContainerProperties(Map appContainerProperties) {
        CommonPipelineEnvironment.getInstance().appContainerProperties = appContainerProperties
    }

    def setAppContainerProperty(property, value) {
        getAppContainerProperties()[property] = value
    }

    def getAppContainerProperty(property) {
        return getAppContainerProperties()[property]
    }

    // goes into measurement jenkins_custom_data
    def setInfluxCustomDataEntry(key, value) {
        CommonPipelineEnvironment.getInstance().setInfluxCustomDataEntry(key, value)
    }
    // goes into measurement jenkins_custom_data
    @Deprecated // not used in library
    def getInfluxCustomData() {
        CommonPipelineEnvironment.getInstance().getInfluxCustomData()
    }

    // goes into measurement jenkins_custom_data
    def setInfluxCustomDataTagsEntry(key, value) {
        CommonPipelineEnvironment.getInstance().setInfluxCustomDataTagsEntry(key, value)
    }
    // goes into measurement jenkins_custom_data
    @Deprecated // not used in library
    def getInfluxCustomDataTags() {
        CommonPipelineEnvironment.getInstance().getInfluxCustomDataTags()
    }

    void setInfluxCustomDataMapEntry(measurement, field, value) {
        CommonPipelineEnvironment.getInstance().setInfluxCustomDataMapEntry(measurement, field, value)
    }
    @Deprecated // not used in library
    def getInfluxCustomDataMap() {
        CommonPipelineEnvironment.getInstance().getInfluxCustomDataMap()
    }

    def setInfluxCustomDataMapTagsEntry(measurement, tag, value) {
        CommonPipelineEnvironment.getInstance().setInfluxCustomDataMapTagsEntry(measurement, tag, value)
    }
    @Deprecated // not used in library
    def getInfluxCustomDataMapTags() {
        CommonPipelineEnvironment.getInstance().getInfluxCustomDataMapTags()
    }

    @Deprecated // not used in library
    def setInfluxStepData(key, value) {
        CommonPipelineEnvironment.getInstance().setInfluxStepData(key, value)
    }
    @Deprecated // not used in library
    def getInfluxStepData(key) {
        CommonPipelineEnvironment.getInstance().getInfluxStepData(key)
    }

    @Deprecated // not used in library
    def setInfluxPipelineData(key, value) {
        CommonPipelineEnvironment.getInstance().setInfluxPipelineData(key, value)
    }
    @Deprecated // not used in library
    def setPipelineMeasurement(key, value){
        CommonPipelineEnvironment.getInstance().setPipelineMeasurement(key, value)
    }
    @Deprecated // not used in library
    def getPipelineMeasurement(key) {
        CommonPipelineEnvironment.getInstance().getPipelineMeasurement(key)
    }

    Map getStepConfiguration(stepName, stageName = env.STAGE_NAME, includeDefaults = true) {
        CommonPipelineEnvironment.getInstance().getStepConfiguration(stepName, stageName, includeDefaults)
    }
    
    void setPipelineDefaults(pipelineDefaults) {
        CommonPipelineEnvironment.getInstance().pipelineDefaults = pipelineDefaults
    }
    
    def getPipelineDefaults(pipelineDefaults) {
        return CommonPipelineEnvironment.getInstance().pipelineDefaults
    }
}
