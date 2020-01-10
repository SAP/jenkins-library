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
            defaults = ConfigurationLoader.defaultGeneralConfiguration()
            defaults = ConfigurationMerger.merge(ConfigurationLoader.defaultStepConfiguration(null, stepName), null, defaults)
            defaults = ConfigurationMerger.merge(ConfigurationLoader.defaultStageConfiguration(null, stageName), null, defaults)
        }
        Map config = ConfigurationMerger.merge(configuration.get('general') ?: [:], null, defaults)
        config = ConfigurationMerger.merge(configuration.get('steps')?.get(stepName) ?: [:], null, config)
        config = ConfigurationMerger.merge(configuration.get('stages')?.get(stageName) ?: [:], null, config)
        return config
    }

    void writeToDisk(script) {

        def files = [
            [filename: '.pipeline/piperEnvironment/artifactVersion', content: artifactVersion],
            [filename: '.pipeline/piperEnvironment/github/owner', content: githubOrg],
            [filename: '.pipeline/piperEnvironment/github/repository', content: githubRepo],
            [filename: '.pipeline/piperEnvironment/git/branch', content: gitBranch],
            [filename: '.pipeline/piperEnvironment/git/commitId', content: gitCommitId],
            [filename: '.pipeline/piperEnvironment/git/commitMessage', content: gitCommitMessage],
        ]

        files.each({f  ->
            if (f.content && !script.fileExists(f.filename)) {
                script.writeFile file: f.filename, text: f.content
            }
        })

        valueMap.each({key, value ->
            def fileName = ".pipeline/piperEnvironment/custom/${key}"
            if (value && !script.fileExists(fileName)) {
                //ToDo: check for value type and act accordingly?
                script.writeFile file: fileName, text: value
            }
        })
    }

    void readFromDisk() {
        def file = '.pipeline/piperEnvironment/artifactVersion'
        if (fileExists(file)) {
            artifactVersion = readFile(file)
        }

        file = '.pipeline/piperEnvironment/github/owner'
        if (fileExists(file)) {
            githubOrg = readFile(file)
        }

        file = '.pipeline/piperEnvironment/github/repository'
        if (fileExists(file)) {
            githubRepo = readFile(file)
        }

        file = '.pipeline/piperEnvironment/git/branch'
        if (fileExists(file)) {
            gitBranch = readFile(file)
        }

        file = '.pipeline/piperEnvironment/git/commitId'
        if (fileExists(file)) {
            gitCommitId = readFile(file)
        }

        file = '.pipeline/piperEnvironment/git/commitMessage'
        if (fileExists(file)) {
            gitCommitMessage = readFile(file)
        }

        def customValues = findFiles(glob: '.pipeline/piperEnvironment/custom/*')

        customValues.each({f ->
            def fileName = f.getName()
            def param = fileName.split('/')[fileName.split('\\/').size()-1]
            valueMap[param] = readFile(f.getPath())
        })
    }
}
