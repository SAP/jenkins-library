import com.sap.piper.BuildTool
import com.sap.piper.DownloadCacheUtils
import groovy.transform.Field

import static com.sap.piper.Prerequisites.checkScript

@Field String STEP_NAME = getClass().getName()
@Field String METADATA_FILE = 'metadata/npmExecuteLint.yaml'

//Metadata maintained in file project://resources/metadata/npmExecuteLint.yaml

void call(Map parameters = [:]) {
    final script = checkScript(this, parameters) ?: this
    parameters = DownloadCacheUtils.injectDownloadCacheInParameters(script, parameters, BuildTool.NPM)

    String eslintDefaultConfig = libraryResource ".eslintrc.json"
    writeFile file: ".pipeline/.eslintrc.json", text: eslintDefaultConfig

    piperExecuteBin(parameters, STEP_NAME, METADATA_FILE, [])

    visualizeLintingResults(script)
}

private visualizeLintingResults(Script script) {
    recordIssues blameDisabled: true,
        enabledForFailure: true,
        aggregatingResults: false,
        tool: script.checkStyle(id: "lint", name: "Lint", pattern: "*lint.xml")
}
