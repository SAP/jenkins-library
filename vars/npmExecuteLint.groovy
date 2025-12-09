import com.sap.piper.BuildTool
import com.sap.piper.DownloadCacheUtils
import groovy.transform.Field
import hudson.AbortException

import static com.sap.piper.Prerequisites.checkScript

@Field String STEP_NAME = getClass().getName()
@Field String METADATA_FILE = 'metadata/npmExecuteLint.yaml'

//Metadata maintained in file project://resources/metadata/npmExecuteLint.yaml

void call(Map parameters = [:]) {
    final script = checkScript(this, parameters) ?: this
    parameters = DownloadCacheUtils.injectDownloadCacheInParameters(script, parameters, BuildTool.NPM)

    try {
        piperExecuteBin(parameters, STEP_NAME, METADATA_FILE, [])
    } catch (Exception exception) {
        error("Linter execution failed. Please examine the reports which are also available in the Jenkins user interface.")
    }
    finally {
        visualizeLintingResults(script)
    }
}

private visualizeLintingResults(Script script) {
    try {
        recordIssues skipBlames: true,
            enabledForFailure: true,
            aggregatingResults: false,
            tool: script.checkStyle(id: "lint", name: "Lint", pattern: "*lint.xml")
    } catch (e) {
        echo "recordIssues has failed. Possibly due to an outdated version of the warnings-ng plugin."
        e.printStackTrace()
    }
}
