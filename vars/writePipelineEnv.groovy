import static com.sap.piper.Prerequisites.checkScript
import groovy.transform.Field

@Field def STEP_NAME = getClass().getName()

void call(Map parameters = [:]) {
    final script = checkScript(this, parameters) ?: this
    String piperGoPath = parameters?.piperGoPath ?: './piper'
    echo "before getCPEMap()"
    Map cpe = script?.commonPipelineEnvironment?.getCPEMap(script)
    echo "after getCPEMap()"
    echo "CPE: ${cpe}"
    if (cpe == null) {
        return
    }
    try {
        def jsonMap = groovy.json.JsonOutput.toJson(cpe)
    } catch (ex java.lang.StackOverflowError) {
        echo "stack overflow error occured - ignoring it for now"
    }
    
    def writePipelineEnvCommand = """
${piperGoPath} writePipelineEnv <<EOF
${jsonMap}
EOF
"""

    def output = script.sh(returnStdout: true, script: writePipelineEnvCommand)
    if (parameters?.verbose) {
        script.echo("wrote commonPipelineEnvironment: ${output}")
    }
}
