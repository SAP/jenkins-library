import static com.sap.piper.Prerequisites.checkScript
import groovy.transform.Field

@Field def STEP_NAME = getClass().getName()

void call(Map parameters = [:]) {
    final script = checkScript(this, parameters) ?: this
    String piperGoPath = parameters.piperGoPath ?: './piper'
    Map cpe = script.commonPipelineEnvironment.getCPEMap(script)
    if (cpe == null) {
        return
    }
    def jsonMap = groovy.json.JsonOutput.toJson(cpe)
    def writePipelineEnvCommand = """
${piperGoPath} writePipelineEnv <<EOF
${jsonMap}
EOF
"""

    def output = script.sh(returnStdout: true, script: writePipelineEnvCommand)
}
