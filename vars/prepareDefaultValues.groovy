import com.sap.piper.GenerateDocumentation
import com.sap.piper.DefaultValueCache
import com.sap.piper.MapUtils
import com.sap.piper.config.ConfigCache

import groovy.transform.Field

@Field STEP_NAME = getClass().getName()

@Field Set GENERAL_CONFIG_KEYS = []
@Field Set STEP_CONFIG_KEYS = []
@Field Set PARAMETER_KEYS = []

/**
 * Loads the pipeline library default values from the file `resources/default_pipeline_environment.yml`.
 * Afterwards the values can be loaded by the method: `ConfigurationLoader.defaultStepConfiguration`
 */
@GenerateDocumentation
void call(Map parameters = [:]) {
    if(!DefaultValueCache.getInstance() || parameters.customDefaults) {
        def defaultValues = ConfigCache.getInstance(this).getPiperDefaults()

        def configFileList = []
        def customDefaults = parameters.customDefaults

        if(customDefaults in String)
            customDefaults = [customDefaults]
        if(customDefaults in List)
            configFileList += customDefaults
        for (def configFileName : configFileList){
            echo "Loading configuration file '${configFileName}'"
            def configuration = readYaml text: libraryResource(configFileName)
            defaultValues = MapUtils.merge(defaultValues, MapUtils.pruneNulls(configuration))
        }
        DefaultValueCache.createInstance(defaultValues)
    }
}
