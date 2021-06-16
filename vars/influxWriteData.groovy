import groovy.transform.Field
import static com.sap.piper.Prerequisites.checkScript

@Field String STEP_NAME = getClass().getName()
@Field String METADATA_FILE = 'metadata/influx.yaml'

void call(Map parameters = [:]) {
        List credentials = [
        [type: 'token', id: 'influxTokenId', env: ['PIPER_influxToken']]
        ]
        parameters.piperGoPath = "./piper123"
        def script = checkScript(this, parameters) ?: this
        piperExecuteBin.prepareMetadataResource(script, METADATA_FILE)
        String piperGoPath = parameters.piperGoPath ?: './piper'
        String customDefaultConfig = piperExecuteBin.getCustomDefaultConfigsArg()
        String customConfigArg = piperExecuteBin.getCustomConfigArg(script)
        Map stepConfig = readJSON(text: sh(returnStdout: true, script: "${piperGoPath} getConfig --stepMetadata '.pipeline/tmp/${METADATA_FILE}'${customDefaultConfig}${customConfigArg}"))
        echo "Step Config: ${stepConfig}"
        parameters.dataMap = stepConfig.dataMap
        parameters.dataMapTags = stepConfig.dataMapTags
        parameters.dataMap.series_3 = [field_e: 31, field_g: 32]
        parameters.dataMapTags.series_3 = [tag_e: 'e', tag_g: 'g']
        piperExecuteBin(parameters, STEP_NAME, METADATA_FILE, credentials)
}
