import static com.sap.piper.Prerequisites.checkScript

import groovy.transform.Field

@Field def STEP_NAME = getClass().getName()

void call(Map parameters = [:]) {
    final script = checkScript(this, parameters) ?: this
    String piperGoPath = parameters?.piperGoPath ?: './piper'
    def output = script.sh(returnStdout: true, script: "${piperGoPath} readPipelineEnv")
    Map cpeMap = script.readJSON(text: output)
    script?.commonPipelineEnvironment?.setCPEMap(script, cpeMap)
}
