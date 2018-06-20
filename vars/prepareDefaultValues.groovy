import com.sap.piper.DefaultValueCache
import com.sap.piper.MapUtils

def call(Map parameters = [:]) {
    handlePipelineStepErrors (stepName: 'prepareDefaultValues', stepParameters: parameters) {
        if(!DefaultValueCache.getInstance() || parameters.customDefaults) {
            def defaultValues = [:]
            def configurationFiles = ['default_pipeline_environment.yml']
            def customDefaults = parameters.customDefaults

            if(customDefaults in String)
                customDefaults = [customDefaults]
            if(customDefaults in List)
                configurationFiles += customDefaults
            for (def configFileName : configurationFiles){
                if(configurationFiles.size() > 1) echo "Loading configuration file '${}'"
                def configuration = readYaml text: libraryResource(configFileName)
                defaultValues = MapUtils.merge(
                        MapUtils.pruneNulls(defaultValues),
                        MapUtils.pruneNulls(configuration))
            }
            DefaultValueCache.createInstance(defaultValues)
        }
    }
}
