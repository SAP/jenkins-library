import com.sap.piper.ConfigurationLoader
import com.sap.piper.ConfigurationMerger
import com.sap.piper.DefaultValueCache
import com.sap.piper.analytics.InfluxData
import com.sap.piper.JsonUtils

class commonPipelineEnvironment implements Serializable {

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

    String xsDeploymentId

    //GiutHub specific information
    String githubOrg
    String githubRepo

    //stores properties for a pipeline which build an artifact and then bundles it into a container
    private Map appContainerProperties = [:]

    Map configuration = [:]
    Map containerProperties = [:]
    Map defaultConfiguration = [:]


    String mtarFilePath
    private Map valueMap = [:]

    void setValue(String property, value) {
        valueMap[property] = value
    }

    def getValue(String property) {
        return valueMap.get(property)
    }

    String changeDocumentId

    def reset() {
        appContainerProperties = [:]
        artifactVersion = null

        configuration = [:]
        containerProperties = [:]

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

    def setContainerProperty(property, value) {
        containerProperties[property] = value
    }

    def getContainerProperty(property) {
        return containerProperties[property]
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
            defaults = ConfigurationLoader.defaultGeneralConfiguration()
            defaults = ConfigurationMerger.merge(ConfigurationLoader.defaultStepConfiguration(null, stepName), null, defaults)
            defaults = ConfigurationMerger.merge(ConfigurationLoader.defaultStageConfiguration(null, stageName), null, defaults)
        }
        Map config = ConfigurationMerger.merge(configuration.get('general') ?: [:], null, defaults)
        config = ConfigurationMerger.merge(configuration.get('steps')?.get(stepName) ?: [:], null, config)
        config = ConfigurationMerger.merge(configuration.get('stages')?.get(stageName) ?: [:], null, config)
        return config
    }

    def files = [
        [filename: '.pipeline/commonPipelineEnvironment/artifactVersion', property: 'artifactVersion'],
        [filename: '.pipeline/commonPipelineEnvironment/github/owner', property: 'githubOrg'],
        [filename: '.pipeline/commonPipelineEnvironment/github/repository', property: 'githubRepo'],
        [filename: '.pipeline/commonPipelineEnvironment/git/branch', property: 'gitBranch'],
        [filename: '.pipeline/commonPipelineEnvironment/git/commitId', property: 'gitCommitId'],
        [filename: '.pipeline/commonPipelineEnvironment/git/commitMessage', property: 'gitCommitMessage'],
    ]

    void writeToDisk(script) {

        files.each({f  ->
            if (this[f.property] && !script.fileExists(f.filename)) {
                script.writeFile file: f.filename, text: this[f.property]
            }
        })

        containerProperties.each({key, value ->
            def fileName = ".pipeline/commonPipelineEnvironment/container/${key}"
            if (value && !script.fileExists(fileName)) {
                if(value instanceof String) {
                    script.writeFile file: fileName, text: value
                } else {
                    script.writeFile file: fileName, text: JsonUtils.groovyObjectToJsonString(value)
                }
            }
        })

        valueMap.each({key, value ->
            def fileName = ".pipeline/commonPipelineEnvironment/custom/${key}"
            if (value && !script.fileExists(fileName)) {
                if(value instanceof String) {
                    script.writeFile file: fileName, text: value
                } else {
                    script.writeFile file: fileName, text: JsonUtils.groovyObjectToJsonString(value)
                }
            }
        })
    }

    void readFromDisk(script) {

        files.each({f  ->
            if (script.fileExists(f.filename)) {
                this[f.property] = script.readFile(f.filename)
            }
        })

        def customValues = script.findFiles(glob: '.pipeline/commonPipelineEnvironment/custom/*')

        customValues.each({f ->
            def fileName = f.getName()
            def param = fileName.split('/')[fileName.split('\\/').size()-1]
            valueMap[param] = script.readFile(f.getPath())
        })
    }

    List getCustomDefaults() {
        DefaultValueCache.getInstance().getCustomDefaults()
    }
}
