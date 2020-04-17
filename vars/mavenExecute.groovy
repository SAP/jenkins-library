import com.sap.piper.DownloadCacheUtils
import groovy.transform.Field

@Field def STEP_NAME = getClass().getName()
@Field String METADATA_FILE = 'metadata/mavenExecute.yaml'

def call(Map parameters = [:]) {
    List credentials = [ ]
    parameters = DownloadCacheUtils.injectDownloadCacheInMavenParameters(parameters.script, parameters)
    piperExecuteBin(parameters, STEP_NAME, METADATA_FILE, credentials)

    String output = ''
    if (parameters.returnStdout) {
        String outputFile = '.pipeline/maven_output.txt'
        if (!fileExists(outputFile)) {
            error "Text file with contents of maven output does not exist at '$outputFile'"
        }
        output = readFile(outputFile)
    }
    return output
}
