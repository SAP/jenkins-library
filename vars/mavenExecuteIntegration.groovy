import com.sap.piper.BuildTool
import com.sap.piper.DownloadCacheUtils
import groovy.transform.Field

import static com.sap.piper.Prerequisites.checkScript

@Field String STEP_NAME = getClass().getName()
@Field String METADATA_FILE = 'metadata/mavenExecuteIntegration.yaml'

@Field Set GENERAL_CONFIG_KEYS = []
@Field Set STEP_CONFIG_KEYS = GENERAL_CONFIG_KEYS
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS.plus([
    /**
     * Specify a glob pattern where test result files will be located.
     */
    'reportLocationPattern',
])

//Metadata maintained in file project://resources/metadata/mavenExecuteIntegration.yaml

void call(Map parameters = [:]) {
    final script = checkScript(this, parameters) ?: this
    parameters = DownloadCacheUtils.injectDownloadCacheInParameters(script, parameters, BuildTool.MAVEN)

    try {
        List credentials = []
        piperExecuteBin(parameters, STEP_NAME, METADATA_FILE, credentials)
    } finally {
        testsPublishResults(script: script, junit: [allowEmptyResults: true, pattern: parameters.reportLocationPattern])
    }
}
