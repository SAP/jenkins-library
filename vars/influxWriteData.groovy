import groovy.transform.Field
import static com.sap.piper.Prerequisites.checkScript

@Field String STEP_NAME = getClass().getName()
@Field String METADATA_FILE = 'metadata/influx.yaml'

void call(Map parameters = [:]) {
        List credentials = [
        [type: 'token', id: 'influxTokenId', env: ['PIPER_influxToken']]
        ]
        def script = checkScript(this, parameters) ?: this
        piperExecuteBin.prepareMetadataResource(script, METADATA_FILE)
        String piperGoPath = parameters.piperGoPath ?: './piper'
        String customDefaultConfig = piperExecuteBin.getCustomDefaultConfigsArg()
        String customConfigArg = piperExecuteBin.getCustomConfigArg(script)
        Map stepConfig = readJSON(text: sh(returnStdout: true, script: "${piperGoPath} getConfig --stepMetadata '.pipeline/tmp/${METADATA_FILE}'${customDefaultConfig}${customConfigArg}"))
        echo "Step Config: ${stepConfig}"
        // Create groovy-side metrics
        stepDataMap = stepConfig.dataMap ?: [:]
        parametersDataMap = parameters.dataMap ?: [:]
        dataMap = stepDataMap + parametersDataMap
        parameters.dataMap = dataMap

        stepDataMapTags = stepConfig.dataMapTags ?: [:]
        parametersDataMapTags = parameters.dataMapTags ?: [:]
        dataMap = stepDataMapTags + parametersDataMapTags
        parameters.dataMapTags = stepDataMapTags
        
        parameters.dataMap.series_3 = [field_e: 31, field_g: 32]
        parameters.dataMapTags.series_3 = [tag_e: 'e', tag_g: 'g']
        piperExecuteBin(parameters, STEP_NAME, METADATA_FILE, credentials)
}
