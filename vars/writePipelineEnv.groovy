import static com.sap.piper.Prerequisites.checkScript
import groovy.transform.Field

@Field def STEP_NAME = getClass().getName()

void call(Map parameters = [:]) {
    final script = checkScript(this, parameters) ?: this
    String piperGoPath = parameters?.piperGoPath ?: './piper'
    String command = "${piperGoPath} writePipelineEnv"

    if (parameters.value) {
        command += " --value '${parameters.value}'"
    }
    Map cpe = script?.commonPipelineEnvironment?.getCPEMap(script)
    if (cpe == null) return

    def jsonMap = groovy.json.JsonOutput.toJson(cpe)
    if (!jsonMap) {
        script.echo("can't write pipelineEnvironment: empty environment")
        return
    }
    withEnv(["PIPER_pipelineEnv=${jsonMap}"]) {
        def output = script.sh(returnStdout: true, script: command)
        if (parameters?.verbose) {
            script.echo("wrote commonPipelineEnvironment: ${output}")
        }
    }
}
