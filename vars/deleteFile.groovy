import com.sap.piper.ConfigurationHelper
import groovy.transform.Field

import static com.sap.piper.Prerequisites.checkScript

@Field String STEP_NAME = getClass().getName()
@Field Set GENERAL_CONFIG_KEYS = []
@Field Set STEP_CONFIG_KEYS = GENERAL_CONFIG_KEYS + [
    /**
     * The `String` path of the file to delete
     */
    'path'
]
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS
/**
 * Deletes a file at a given path, if it exists, or silently ignores the call, if if does not.
 * If the file exists but cannot be deleted, the script will fail the build with an according error.
 * @param arguments - the `Map` of arguments specifying what to delete.
 */
void call(Map<String, String> arguments) {
    handlePipelineStepErrors (stepName: STEP_NAME, stepParameters: arguments) {
        def script = checkScript(this, arguments)  ?: this

        // load default & individual configuration
        Map config = ConfigurationHelper.newInstance(this)
            .loadStepDefaults()
            .mixinStepConfig(script.commonPipelineEnvironment, STEP_CONFIG_KEYS)
            .mixinStageConfig(script.commonPipelineEnvironment, arguments.stageName ?: env.STAGE_NAME, STEP_CONFIG_KEYS)
            .mixin(arguments, PARAMETER_KEYS)
            .use()

        String path = config?.path
        if(!path) {
            error "[DeleteFile] File path must not be null or empty."
        }

        deleteFileIfPresent(path)
    }
}

/**
 * Removes the given file, if it exists.
 * @param filePath the path to the file to remove.
 */
private void deleteFileIfPresent(String filePath) {

    Boolean fileExists = fileExists file: filePath
    if(fileExists) {
        Boolean failure = sh script: "rm '${filePath}'", returnStatus: true
        if(!failure) {
            echo "[DeleteFile] Successfully deleted file '${filePath}'."
        }
        else {
            error "[DeleteFile] Could not delete file '${filePath}'. Check file permissions."
        }
    }
}
