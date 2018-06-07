import com.sap.piper.DefaultValueCache

import hudson.AbortException

def call(Map parameters = [:]) {
    handlePipelineStepErrors (stepName: 'prepareDefaultValues', stepParameters: parameters) {
        if(!DefaultValueCache.getInstance()) {
            Map defaultValues  = readYaml text: libraryResource('default_pipeline_environment.yml')
            Map customDefaultValues = null

            def customDefaults = null

            try {
              customDefaults = libraryResource('pipeline_environment.yml')
            } catch(AbortException e) {
                // custom defaults file not found, that's OK the file is optional.
            }

            if( customDefaults) {
                customDefaultValues = readYaml text: customDefaults
            }
            echo "CUSTOM_DEFAULT_VALUES: '${customDefaultValues}'."
            DefaultValueCache.createInstance(defaultValues, customDefaultValues)
        }
    }
}
