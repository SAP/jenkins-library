import com.sap.piper.ConfigurationLoader
import com.sap.piper.ConfigurationMerger
import com.sap.piper.DefaultValueCache
import com.sap.piper.analytics.InfluxData

class commonPipelineEnvironment implements Serializable {

    //
    // We forward to cpe declared on DefaultValueCache
    def methodMissing(String name, def args) {
        DefaultValueCache.commonPipelineEnvironment.invokeMethod(name, args)
    }

    def propertyMissing(def name) {
       DefaultValueCache.commonPipelineEnvironment[name]
    }

    def propertyMissing(def name, def value) {
       DefaultValueCache.commonPipelineEnvironment[name] = value
    }
    // End forwarding to DefaultValueCache
    //

    /*
     * Should only be used by tests
     */
    void setConfiguration(Map configuration) {
        DefaultValueCache.createInstance(DefaultValueCache.getInstance()?.getDefaultValues() ?: [:], configuration)
    }

    def getConfiguration() {
        DefaultValueCache.getInstance().getProjectConfig()
    }

    def reset() {
        DefaultValueCache.commonPipelineEnvironment.appContainerProperties = [:]
        artifactVersion = null

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

    Map getStepConfiguration(stepName, stageName = env.STAGE_NAME, includeDefaults = true) {
        Map defaults = [:]
        if (includeDefaults) {
            defaults = DefaultValueCache.getInstance()?.getDefaultValues()?.general ?: [:]
            defaults = ConfigurationMerger.merge(ConfigurationLoader.defaultStepConfiguration(stepName), null, defaults)
            defaults = ConfigurationMerger.merge(ConfigurationLoader.defaultStageConfiguration(stageName), null, defaults)
        }
        Map projectConfig = DefaultValueCache.getInstance().getProjectConfig()
        Map config = ConfigurationMerger.merge(projectConfig.get('general') ?: [:], null, defaults)
        config = ConfigurationMerger.merge(projectConfig.get('steps')?.get(stepName) ?: [:], null, config)
        config = ConfigurationMerger.merge(projectConfig.get('stages')?.get(stageName) ?: [:], null, config)
        return config
    }
}
