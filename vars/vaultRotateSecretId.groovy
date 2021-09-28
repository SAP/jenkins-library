import groovy.transform.Field
import static com.sap.piper.Prerequisites.checkScript

@Field String STEP_NAME = getClass().getName()
@Field String METADATA_FILE = 'metadata/vaultRotateSecretId.yaml'

void call(Map parameters = [:]) {
        def script = checkScript(this, parameters) ?: this
        piperExecuteBin(parameters, STEP_NAME, METADATA_FILE, [], false, false, false)
}
