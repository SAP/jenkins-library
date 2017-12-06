import com.sap.piper.DefaultValueCache

def call(Map parameters = [:]) {
    handlePipelineStepErrors (stepName: 'prepareDefaultValues', stepParameters: parameters) {
        if(!DefaultValueCache.getInstance()) {
            Map defaultValues  = readYaml text: libraryResource('default_pipeline_environment.yml')
            DefaultValueCache.createInstance(defaultValues)
        }
    }
}
