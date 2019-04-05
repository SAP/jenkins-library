import com.sap.piper.GenerateDocumentation
import com.sap.piper.DefaultValueCache
import com.sap.piper.MapUtils

import groovy.transform.Field

@Field STEP_NAME = getClass().getName()

/**
 * Loads the pipeline library default values from the file `resources/default_pipeline_environment.yml`.
 * Afterwards the values can be loaded by the method: `ConfigurationLoader.defaultStepConfiguration`
 */
@GenerateDocumentation
void call(Map parameters = [:]) {
    if(!DefaultValueCache.getInstance() || parameters.customDefaults) {
        def defaultValues = [:]
        def configFileList = ['default_pipeline_environment.yml']
        def customDefaults = parameters.customDefaults

        if(customDefaults in String)
            customDefaults = [customDefaults]
        if(customDefaults in List)
            configFileList += customDefaults
        for (def configFileName : configFileList){
            if(configFileList.size() > 1) echo "Loading configuration file '${configFileName}'"
            def configuration = readYaml text: libraryResource(configFileName)
            defaultValues = MapUtils.merge(
                    MapUtils.pruneNulls(defaultValues),
                    MapUtils.pruneNulls(configuration))
        }
        DefaultValueCache.createInstance(defaultValues)
    }
}
