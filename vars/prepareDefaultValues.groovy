import com.sap.piper.DefaultValueCache
import com.sap.piper.MapUtils

def call(Map parameters = [:]) {
    handlePipelineStepErrors (stepName: 'prepareDefaultValues', stepParameters: parameters) {
        if(!DefaultValueCache.getInstance() || parameters.customDefaults) {
            def defaultValues = [:]
            def configurationFiles = ['default_pipeline_environment.yml']
            def customDefaults = parameters.customDefaults
            
            if(customDefaults in String) // >> filename resolves to Map
                customDefaults = [].plus(customDefaults)
            if(customDefaults in List)
                configurationFiles += customDefaults
        /*
        if(defaults instanceof Map) // >> config map
            defaults = [].plus(defaults)
        */
            for (def configFileName : configurationFiles){
                def configuration = readYaml text: libraryResource(configFileName)
                defaultValues = MapUtils.merge(
                        MapUtils.pruneNull(defaultValues),
                        MapUtils.pruneNull(configuration))
            }
            DefaultValueCache.createInstance(defaultValues)
        }
    }
}
