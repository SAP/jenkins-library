import com.sap.piper.ConfigurationLoader
import com.sap.piper.ConfigurationMerger
import com.sap.piper.DefaultValueCache
import com.sap.piper.analytics.InfluxData
import groovy.json.JsonOutput

class commonPipelineEnvironment implements Serializable {

    //Project identifier which might be used to distinguish resources which are available globally, e.g. for locking
    def projectName

    //stores version of the artifact which is build during pipeline run
    def artifactVersion
    def originalArtifactVersion

    // stores additional artifact coordinates
    def artifactId
    def groupId
    def packaging

    //stores the build tools if it inferred automatically, e.g. in the SAP Cloud SDK pipeline
    String buildTool

    //Stores the current buildResult
    String buildResult = 'SUCCESS'

    //stores the gitCommitId and gitRemoteCommitId as additional git information for the build during pipeline run
    /*
       Incase of 'Merging the pull request with the current target branch revision' stratergy in jenkins,
       the jenkins creates its own local merge commit which is stored in gitCommitId.
       gitRemoteCommitId will contain the actual remote merge commitId on git rather than local commitId
    */
    String gitCommitId
    String gitRemoteCommitId
    String gitHeadCommitId
    String gitCommitMessage
    String gitSshUrl
    String gitHttpsUrl
    String gitBranch
    String gitRef

    String xsDeploymentId

    //GitHub specific information
    String githubOrg
    String githubRepo

    // GitHub deployment ID
    String githubDeploymentId

    //stores properties for a pipeline which build an artifact and then bundles it into a container
    private Map appContainerProperties = [:]

    Map configuration = [:]
    Map containerProperties = [:]
    Map defaultConfiguration = [:]

    // Location of the file from where the configuration was parsed. See setupCommonPipelineEnvironment.groovy
    // Useful for making sure that the piper binary uses the same file when called from Jenkins.
    String configurationFile = ''

    String mtarFilePath = null

    String abapAddonDescriptor


    private Map valueMap = [:]

    void setValue(String property, value) {
        valueMap[property] = value
    }

    void removeValue(String property) {
        valueMap.remove(property)
    }

    def getValue(String property) {
        return valueMap.get(property)
    }

    String changeDocumentId

    def reset() {

        projectName = null

        abapAddonDescriptor = null

        appContainerProperties = [:]
        artifactVersion = null
        originalArtifactVersion = null

        artifactId = null
        groupId = null
        packaging = null

        buildTool = null

        configuration = [:]
        containerProperties = [:]

        gitCommitId = null
        gitRemoteCommitId = null
        gitHeadCommitId = null
        gitCommitMessage = null
        gitSshUrl = null
        gitHttpsUrl = null
        gitBranch = null
        gitRef = null

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
        Map config = ConfigurationMerger.merge(configuration.get('general') ?: [:] as Map, null, defaults)
        config = ConfigurationMerger.merge(configuration.get('steps')?.get(stepName) ?: [:], null, config)
        config = ConfigurationMerger.merge(configuration.get('stages')?.get(stageName) ?: [:], null, config)
        return config
    }

    def files = [
        [filename: '.pipeline/commonPipelineEnvironment/artifactVersion', property: 'artifactVersion'],
        [filename: '.pipeline/commonPipelineEnvironment/artifactId', property: 'artifactId'],
        [filename: '.pipeline/commonPipelineEnvironment/groupId', property: 'groupId'],
        [filename: '.pipeline/commonPipelineEnvironment/packaging', property: 'packaging'],
        [filename: '.pipeline/commonPipelineEnvironment/buildTool', property: 'buildTool'],
        [filename: '.pipeline/commonPipelineEnvironment/originalArtifactVersion', property: 'originalArtifactVersion'],
        [filename: '.pipeline/commonPipelineEnvironment/github/owner', property: 'githubOrg'],
        [filename: '.pipeline/commonPipelineEnvironment/github/repository', property: 'githubRepo'],
        [filename: '.pipeline/commonPipelineEnvironment/git/branch', property: 'gitBranch'],
        [filename: '.pipeline/commonPipelineEnvironment/git/commitId', property: 'gitCommitId'],
        [filename: '.pipeline/commonPipelineEnvironment/git/remoteCommitId', property: 'gitRemoteCommitId'],
        [filename: '.pipeline/commonPipelineEnvironment/git/headCommitId', property: 'gitHeadCommitId'],
        [filename: '.pipeline/commonPipelineEnvironment/git/httpsUrl', property: 'gitHttpsUrl'],
        [filename: '.pipeline/commonPipelineEnvironment/git/ref', property: 'gitRef'],
        [filename: '.pipeline/commonPipelineEnvironment/git/commitMessage', property: 'gitCommitMessage'],
        [filename: '.pipeline/commonPipelineEnvironment/mtarFilePath', property: 'mtarFilePath'],
        [filename: '.pipeline/commonPipelineEnvironment/abap/addonDescriptor', property: 'abapAddonDescriptor'],
        [filename: '.pipeline/commonPipelineEnvironment/git/github_deploymentId', property: 'githubDeploymentId'],
    ]

    Map getCPEMap(script) {
        def cpeMap = [:]
        files.each({f ->
            createMapEntry(script, cpeMap, f.filename, this[f.property])
        })

        containerProperties.each({key, value ->
            def filename = ".pipeline/commonPipelineEnvironment/container/${key}"
            createMapEntry(script, cpeMap, filename, value)
        })

        valueMap.each({key, value ->
            def filename = ".pipeline/commonPipelineEnvironment/custom/${key}"
            createMapEntry(script, cpeMap, filename, value)
        })
        return cpeMap
    }

    void createMapEntry(script, Map resMap, String filename, value) {
        // net.sf.json.JSONNull can come in through readPipelineEnv via readJSON()
        // leaving them in will create a StackOverflowError further down in writePipelineEnv()
        // thus removing them from the map for now
        if (value != null && !(value instanceof net.sf.json.JSONNull)) {
            // prefix is assumed by step if nothing else is specified
            def prefix = ~/^.pipeline\/commonPipelineEnvironment\//
            filename -= prefix
            resMap[filename] = value
        }
    }

    def setCPEMap(script, Map cpeMap) {
        if (cpeMap == null) return
        def prefix = ~/^.pipeline\/commonPipelineEnvironment\//
        files.each({f ->
                        def key = f.filename - prefix
                        if (cpeMap.containsKey(key)) this[f.property] = cpeMap[key]
                   })

        cpeMap.each {
            if (it.key.startsWith("custom/")) valueMap[it.key - ~/^custom\//] = it.value
            if (it.key.startsWith("container/")) containerProperties[it.key - ~/^container\//] = it.value
        }
    }

    List getCustomDefaults() {
        DefaultValueCache.getInstance().getCustomDefaults()
    }
}
