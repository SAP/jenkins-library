import com.sap.piper.DefaultValueCache
import com.sap.piper.MapUtils

import hudson.AbortException

def call(Map parameters = [:]) {
    handlePipelineStepErrors (stepName: 'prepareDefaultValues', stepParameters: parameters) {
        if(!DefaultValueCache.getInstance() || parameters.customDefaults) {
            def configurationFiles = ['default_pipeline_environment.yml']
            def defaultConfiguration = [:]

            def customDefaults = parameters.customDefaults
            if(customDefaults in String) // >> filename resolves to Map
                customDefaults = [].plus(customDefaults)
            // customDefaults is Map / null
            configurationFiles += customDefaults
        /*
        if(defaults instanceof Map) // >> config map
            defaults = [].plus(defaults)
        */
        //if(configurationFiles in List) // >> list of String / Map
            for (def configFileName : configurationFiles){
                def configuration = readYaml text: libraryResource(configFileName)
                defaultConfiguration = MapUtils.merge(
                        MapUtils.pruneNull(defaultConfiguration),
                        MapUtils.pruneNull(configuration))
            }

            DefaultValueCache.createInstance(defaultConfiguration)
        }
    }
}
