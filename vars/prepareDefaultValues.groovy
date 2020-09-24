import com.sap.piper.GenerateDocumentation
import com.sap.piper.DefaultValueCache

import groovy.transform.Field

import static com.sap.piper.Prerequisites.checkScript

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
    def script = checkScript(this, parameters)
    DefaultValueCache.prepare(script, parameters)
}
